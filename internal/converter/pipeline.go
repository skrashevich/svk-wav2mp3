package converter

import (
	"context"
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

// RunPipeline читает WAV и пишет MP3 чанками по chunkSamples фреймов,
// обновляя progressbar. Корректно реагирует на отмену контекста между чанками.
func RunPipeline(ctx context.Context, reader *WAVReader, writer *MP3Writer, quiet bool) error {
	totalSamples := reader.TotalSamples()

	var bar *progressbar.ProgressBar
	if !quiet {
		if totalSamples > 0 {
			bar = progressbar.NewOptions(totalSamples,
				progressbar.OptionSetDescription("Кодирование"),
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
				progressbar.OptionSetDescription("Кодирование"),
				progressbar.OptionSetWriter(os.Stderr),
				progressbar.OptionSpinnerType(14),
			)
		}
	}

	for {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("операция отменена: %w", err)
		}

		samples, err := reader.ReadSamplesInt16()
		if err != nil {
			return err
		}
		if samples == nil {
			break
		}

		if err := writer.WriteSamples(samples); err != nil {
			return err
		}

		if bar != nil {
			frames := len(samples) / reader.Info.NumChannels
			_ = bar.Add(frames)
		}
	}

	if bar != nil {
		_ = bar.Finish()
	}

	return nil
}
