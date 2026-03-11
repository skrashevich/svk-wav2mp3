package converter

import (
	"context"
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

// chunk holds a pre-read block of PCM samples for pipelined processing.
type chunk struct {
	samples []int16
	err     error
}

// RunPipeline reads WAV and writes MP3 with pipelined I/O: the next WAV
// chunk is read concurrently while the current chunk is being encoded.
// Updates the progressbar and responds to context cancellation.
func RunPipeline(ctx context.Context, reader *WAVReader, writer *MP3Writer, quiet bool) error {
	totalSamples := reader.TotalSamples()

	var bar *progressbar.ProgressBar
	if !quiet {
		if totalSamples > 0 {
			bar = progressbar.NewOptions(totalSamples,
				progressbar.OptionSetDescription("Encoding"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionShowBytes(false),
				progressbar.OptionSetWidth(40),
				progressbar.OptionThrottle(0),
				progressbar.OptionShowCount(),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionSetRenderBlankState(true),
			)
		} else {
			bar = progressbar.NewOptions(-1,
				progressbar.OptionSetDescription("Encoding"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionSpinnerType(14),
			)
		}
	}

	// Buffered channel of 1: allows the reader goroutine to stay one chunk ahead.
	ch := make(chan chunk, 1)

	go func() {
		defer close(ch)
		for {
			if ctx.Err() != nil {
				return
			}
			samples, err := reader.ReadSamplesInt16()
			if err != nil {
				ch <- chunk{err: err}
				return
			}
			if samples == nil {
				return // EOF
			}
			// Copy the samples since ReadSamplesInt16 reuses its buffer.
			copied := make([]int16, len(samples))
			copy(copied, samples)
			ch <- chunk{samples: copied}
		}
	}()

	for c := range ch {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("operation cancelled: %w", err)
		}
		if c.err != nil {
			return c.err
		}

		if err := writer.WriteSamples(c.samples); err != nil {
			return err
		}

		if bar != nil {
			frames := len(c.samples) / reader.Info.NumChannels
			_ = bar.Add(frames)
		}
	}

	if bar != nil {
		_ = bar.Finish()
	}

	return nil
}
