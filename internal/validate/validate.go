package validate

import (
	"fmt"
	"os"
	"strings"

	"github.com/svk/wav2mp3/internal/config"
)

// ConvertOptions validates conversion parameter correctness.
func ConvertOptions(opts *config.ConvertOptions) error {
	if opts.InputPath == "" {
		return fmt.Errorf("input file not specified (use -i)")
	}
	if _, err := os.Stat(opts.InputPath); os.IsNotExist(err) {
		return fmt.Errorf("input file not found: %s", opts.InputPath)
	}
	if !strings.EqualFold(getExt(opts.InputPath), ".wav") {
		return fmt.Errorf("input file must have .wav extension: %s", opts.InputPath)
	}

	if opts.VBR {
		if opts.VBRQuality < 0.0 || opts.VBRQuality > 9.9 {
			return fmt.Errorf("VBR quality must be in range 0.0–9.9, got: %.1f", opts.VBRQuality)
		}
	} else {
		if opts.Bitrate < 32 || opts.Bitrate > 320 {
			return fmt.Errorf("bitrate must be in range 32–320 kbps, got: %d", opts.Bitrate)
		}
	}

	if opts.Quality < 0 || opts.Quality > 9 {
		return fmt.Errorf("quality must be in range 0–9, got: %d", opts.Quality)
	}

	if opts.Tags.Cover != "" {
		if _, err := os.Stat(opts.Tags.Cover); os.IsNotExist(err) {
			return fmt.Errorf("cover file not found: %s", opts.Tags.Cover)
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
