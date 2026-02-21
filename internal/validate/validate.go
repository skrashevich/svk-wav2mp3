package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/svk/wav2mp3/internal/config"
)

// ConvertOptions проверяет корректность параметров конвертации.
func ConvertOptions(opts *config.ConvertOptions) error {
	if opts.InputPath == "" {
		return fmt.Errorf("входной файл не указан (используйте -i)")
	}
	if _, err := os.Stat(opts.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("входной файл не найден: %s", opts.InputPath)
	}
	if !strings.EqualFold(getExt(opts.InputPath), ".wav") {
		return fmt.Errorf("входной файл должен иметь расширение .wav: %s", opts.InputPath)
	}

	if opts.VBR {
		if opts.VBRQuality < 0.0 || opts.VBRQuality > 9.9 {
			return fmt.Errorf("VBR quality должно быть в диапазоне 0.0–9.9, получено: %.1f", opts.VBRQuality)
		}
	} else {
		if opts.Bitrate < 32 || opts.Bitrate > 320 {
			return fmt.Errorf("bitrate должен быть в диапазоне 32–320 kbps, получено: %d", opts.Bitrate)
		}
	}

	if opts.Quality < 0 || opts.Quality > 9 {
		return fmt.Errorf("quality должен быть в диапазоне 0–9, получено: %d", opts.Quality)
	}

	if opts.Tags.Cover != "" {
		if _, err := os.Stat(opts.Tags.Cover); os.IsNotExist(err) {
			return fmt.Errorf("файл обложки не найден: %s", opts.Tags.Cover)
		}
	}

	return nil
}

func getExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ""
}
