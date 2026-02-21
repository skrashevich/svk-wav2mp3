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

// Stats содержит статистику после конвертации.
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

// Convert выполняет полный цикл конвертации WAV → MP3.
func Convert(ctx context.Context, opts config.ConvertOptions) (*Stats, error) {
	outputPath := resolveOutputPath(opts.InputPath, opts.OutputPath)

	// Защита: удалить недозаписанный файл при ошибке или отмене
	success := false
	defer func() {
		if !success {
			os.Remove(outputPath)
		}
	}()

	// Открываем WAV
	reader, err := NewWAVReader(opts.InputPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	// Создаём MP3-энкодер
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

	// Кодируем
	if err := RunPipeline(ctx, reader, writer, opts.Quiet); err != nil {
		writer.Close()
		return nil, err
	}

	// Закрываем энкодер (flush + close)
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ошибка завершения кодирования: %w", err)
	}

	elapsed := time.Since(start)

	// Записываем ID3v2-теги после закрытия энкодера
	hasAnyTag := opts.Tags.Title != "" || opts.Tags.Artist != "" || opts.Tags.Album != "" ||
		opts.Tags.Year != "" || opts.Tags.Genre != "" || opts.Tags.Track != "" ||
		opts.Tags.Comment != "" || opts.Tags.Cover != ""

	if hasAnyTag {
		if err := tagger.Apply(outputPath, opts.Tags); err != nil {
			return nil, err
		}
	}

	// Размер выходного файла
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

// resolveOutputPath генерирует путь выходного файла если не задан явно.
func resolveOutputPath(inputPath, outputPath string) string {
	if outputPath != "" {
		return outputPath
	}
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return base + ".mp3"
}
