package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/svk/wav2mp3/internal/config"
	"github.com/svk/wav2mp3/internal/tagger"
)

// Stats contains conversion statistics.
type Stats struct {
	InputPath   string
	OutputPath  string
	WAVInfo     WAVInfo
	OutputSizeB int64
	Elapsed     time.Duration
	VBR         bool
	VBRQuality  float64
	Bitrate     int
	Tags        config.ID3Tags
}

// Convert performs full conversion cycle WAV → MP3.
func Convert(ctx context.Context, opts config.ConvertOptions) (*Stats, error) {
	outputPath := resolveOutputPath(opts.InputPath, opts.OutputPath)

	// Safety: remove partial output file on error or cancellation
	success := false
	defer func() {
		if !success {
			os.Remove(outputPath)
		}
	}()

	// Open WAV
	reader, err := NewWAVReader(opts.InputPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Create MP3 encoder
	encCfg := EncoderConfig{
		SampleRate:  reader.Info.SampleRate,
		NumChannels: reader.Info.NumChannels,
		VBR:         opts.VBR,
		VBRQuality:  opts.VBRQuality,
		Bitrate:     opts.Bitrate,
		Quality:     opts.Quality,
	}
	writer, err := NewMP3Writer(outputPath, encCfg)
	if err != nil {
		return nil, err
	}

	start := time.Now()

	// Encode
	if err := RunPipeline(ctx, reader, writer, opts.Quiet); err != nil {
		writer.Close()
		return nil, err
	}

	// Close encoder (flush + close)
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("encoder flush/close error: %w", err)
	}

	elapsed := time.Since(start)

	// Write ID3v2 tags after encoder is closed
	hasAnyTag := opts.Tags.Title != "" || opts.Tags.Artist != "" || opts.Tags.Album != "" ||
		opts.Tags.Year != "" || opts.Tags.Genre != "" || opts.Tags.Track != "" ||
		opts.Tags.Comment != "" || opts.Tags.Cover != ""

	if hasAnyTag {
		if err := tagger.Apply(outputPath, opts.Tags); err != nil {
			return nil, err
		}
	}

	// Output file size
	fi, err := os.Stat(outputPath)
	if err != nil {
		return nil, err
	}

	success = true

	return &Stats{
		InputPath:   opts.InputPath,
		OutputPath:  outputPath,
		WAVInfo:     reader.Info,
		OutputSizeB: fi.Size(),
		Elapsed:     elapsed,
		VBR:         opts.VBR,
		VBRQuality:  opts.VBRQuality,
		Bitrate:     opts.Bitrate,
		Tags:        opts.Tags,
	}, nil
}

// resolveOutputPath generates output file path if not explicitly provided.
func resolveOutputPath(inputPath, outputPath string) string {
	if outputPath != "" {
		return outputPath
	}
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + ".mp3"
}
