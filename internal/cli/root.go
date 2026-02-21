package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/svk/wav2mp3/internal/config"
	"github.com/svk/wav2mp3/internal/converter"
	"github.com/svk/wav2mp3/internal/validate"
)

var opts = config.DefaultConvertOptions()

// NewRootCmd создаёт и возвращает корневую команду cobra.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "wav2mp3 -i INPUT [flags]",
		Short: "Конвертер WAV в MP3 с максимальным качеством",
		Long: `wav2mp3 конвертирует WAV-файлы в MP3 через libmp3lame с поддержкой
ID3v2-тегов и обложки альбома.

По умолчанию используется VBR V2 (эквивалент lame -V2 -q2).
Для CBR укажите --bitrate; флаги --bitrate и --vbr-quality несовместимы.`,
		Version:      version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConvert(cmd)
		},
	}

	f := root.Flags()

	// Основные параметры
	f.StringVarP(&opts.InputPath, "input", "i", "", "Входной WAV файл (обязательный)")
	f.StringVarP(&opts.OutputPath, "output", "o", "", "Выходной MP3 файл (по умолчанию: рядом с входным)")

	// ID3-теги
	f.StringVar(&opts.Tags.Title, "title", "", "Название трека")
	f.StringVar(&opts.Tags.Artist, "artist", "", "Исполнитель")
	f.StringVar(&opts.Tags.Album, "album", "", "Альбом")
	f.StringVar(&opts.Tags.Year, "year", "", "Год")
	f.StringVar(&opts.Tags.Genre, "genre", "", "Жанр")
	f.StringVar(&opts.Tags.Track, "track", "", "Номер трека")
	f.StringVar(&opts.Tags.Comment, "comment", "", "Комментарий")
	f.StringVar(&opts.Tags.Cover, "cover", "", "Путь к файлу обложки (JPEG, PNG, GIF, WebP)")

	// Параметры кодирования
	f.Float64Var(&opts.VBRQuality, "vbr-quality", 2.0,
		"VBR: качество 0.0 (лучше) – 9.9 (хуже), по умолчанию 2.0")
	f.IntVar(&opts.Bitrate, "bitrate", 0,
		"CBR битрейт в kbps (32–320); при указании отключает VBR")
	f.IntVar(&opts.Quality, "quality", 2,
		"Алгоритмическое качество LAME: 0 (лучше) – 9 (быстрее)")

	// Вывод
	f.BoolVarP(&opts.Verbose, "verbose", "v", false, "Подробный вывод (параметры энкодера)")
	f.BoolVarP(&opts.Quiet, "quiet", "q", false, "Минимальный вывод (без прогресс-бара и статистики)")

	_ = root.MarkFlagRequired("input")

	return root
}

func runConvert(cmd *cobra.Command) error {
	ctx := cmd.Context()

	bitrateSet := cmd.Flags().Changed("bitrate")
	vbrQualitySet := cmd.Flags().Changed("vbr-quality")

	if bitrateSet && vbrQualitySet {
		return fmt.Errorf("--bitrate и --vbr-quality несовместимы: выберите один режим")
	}

	// CBR если --bitrate задан явно, иначе VBR
	opts.VBR = !bitrateSet

	if err := validate.ConvertOptions(&opts); err != nil {
		return err
	}

	if opts.Verbose {
		printVerboseInfo()
	}

	stats, err := converter.Convert(ctx, opts)
	if err != nil {
		return err
	}

	if !opts.Quiet {
		printStats(stats)
	}

	return nil
}

func printVerboseInfo() {
	fmt.Fprintf(os.Stderr, "Вход:    %s\n", opts.InputPath)
	if opts.OutputPath != "" {
		fmt.Fprintf(os.Stderr, "Выход:   %s\n", opts.OutputPath)
	}
	if opts.VBR {
		fmt.Fprintf(os.Stderr, "Режим:   VBR V%.0f (качество LAME q=%d)\n", opts.VBRQuality, opts.Quality)
	} else {
		fmt.Fprintf(os.Stderr, "Режим:   CBR %d kbps (качество LAME q=%d)\n", opts.Bitrate, opts.Quality)
	}
	if opts.Tags.Cover != "" {
		fmt.Fprintf(os.Stderr, "Обложка: %s\n", opts.Tags.Cover)
	}
}

func printStats(s *converter.Stats) {
	wav := s.WAVInfo
	channels := "Mono"
	if wav.NumChannels == 2 {
		channels = "Stereo"
	} else if wav.NumChannels > 2 {
		channels = fmt.Sprintf("%dch", wav.NumChannels)
	}

	dur := wav.Duration
	mins := int(dur.Minutes())
	secs := int(dur.Seconds()) % 60

	fmt.Printf("\nВход:  %s (%d Hz, %s, %d-bit, %dm %02ds, %s)\n",
		s.InputPath, wav.SampleRate, channels, wav.BitDepth, mins, secs,
		formatSize(wav.FileSizeB))

	modeStr := fmt.Sprintf("VBR V%.0f", s.VBRQuality)
	if !s.VBR {
		modeStr = fmt.Sprintf("CBR %d kbps", s.Bitrate)
	}

	var compression float64
	if s.OutputSizeB > 0 {
		compression = float64(wav.FileSizeB) / float64(s.OutputSizeB)
	}

	fmt.Printf("Выход: %s (%s, elapsed %.1fs, %s, сжатие %.2fx)\n",
		s.OutputPath, modeStr, s.Elapsed.Seconds(),
		formatSize(s.OutputSizeB), compression)

	var tagParts []string
	if s.Tags.Title != "" {
		tagParts = append(tagParts, fmt.Sprintf("Title=%q", s.Tags.Title))
	}
	if s.Tags.Artist != "" {
		tagParts = append(tagParts, fmt.Sprintf("Artist=%q", s.Tags.Artist))
	}
	if s.Tags.Album != "" {
		tagParts = append(tagParts, fmt.Sprintf("Album=%q", s.Tags.Album))
	}
	if s.Tags.Cover != "" {
		tagParts = append(tagParts, fmt.Sprintf("Cover=%s", s.Tags.Cover))
	}
	if len(tagParts) > 0 {
		fmt.Printf("Теги:  %s\n", strings.Join(tagParts, ", "))
	}
}

// formatSize адаптивно форматирует размер: B / KB / MB.
func formatSize(b int64) string {
	switch {
	case b >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(b)/1024/1024)
	case b >= 1024:
		return fmt.Sprintf("%.1f KB", float64(b)/1024)
	default:
		return fmt.Sprintf("%d B", b)
	}
}
