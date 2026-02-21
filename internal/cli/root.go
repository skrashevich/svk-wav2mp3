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

// NewRootCmd creates and returns root cobra command.
func NewRootCmd(version string) *cobra.Command {
	root := &cobra.Command{
		Use:   "wav2mp3 -i INPUT [flags]",
		Short: "High-quality WAV to MP3 converter",
		Long: `wav2mp3 converts WAV files to MP3 via libmp3lame with support for
ID3v2 tags and album cover.

VBR V2 is used by default (equivalent to lame -V2 -q2).
For CBR use --bitrate; --bitrate and --vbr-quality are incompatible.`,
		Version:      version,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConvert(cmd)
		},
	}

	f := root.Flags()

	// Main parameters
	f.StringVarP(&opts.InputPath, "input", "i", "", "Input WAV file (required)")
	f.StringVarP(&opts.OutputPath, "output", "o", "", "Output MP3 file (default: next to input)")

	// ID3 tags
	f.StringVar(&opts.Tags.Title, "title", "", "Track title")
	f.StringVar(&opts.Tags.Artist, "artist", "", "Artist")
	f.StringVar(&opts.Tags.Album, "album", "", "Album")
	f.StringVar(&opts.Tags.Year, "year", "", "Year")
	f.StringVar(&opts.Tags.Genre, "genre", "", "Genre")
	f.StringVar(&opts.Tags.Track, "track", "", "Track number")
	f.StringVar(&opts.Tags.Comment, "comment", "", "Comment")
	f.StringVar(&opts.Tags.Cover, "cover", "", "Cover file path (JPEG, PNG, GIF, WebP)")

	// Encoding parameters
	f.Float64Var(&opts.VBRQuality, "vbr-quality", 2.0,
		"VBR: quality 0.0 (better) – 9.9 (worse), default 2.0")
	f.IntVar(&opts.Bitrate, "bitrate", 0,
		"CBR bitrate in kbps (32–320); specifying disables VBR")
	f.IntVar(&opts.Quality, "quality", 2,
		"LAME algorithmic quality: 0 (better) – 9 (faster)")

	// Output
	f.BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose output (encoder parameters)")
	f.BoolVarP(&opts.Quiet, "quiet", "q", false, "Minimal output (no progress bar and stats)")

	_ = root.MarkFlagRequired("input")

	return root
}

func runConvert(cmd *cobra.Command) error {
	ctx := cmd.Context()

	bitrateSet := cmd.Flags().Changed("bitrate")
	vbrQualitySet := cmd.Flags().Changed("vbr-quality")

	if bitrateSet && vbrQualitySet {
		return fmt.Errorf("--bitrate and --vbr-quality are incompatible: choose one mode")
	}

	// CBR if --bitrate is set, otherwise VBR
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
	fmt.Fprintf(os.Stderr, "Input:    %s\n", opts.InputPath)
	if opts.OutputPath != "" {
		fmt.Fprintf(os.Stderr, "Output:   %s\n", opts.OutputPath)
	}
	if opts.VBR {
		fmt.Fprintf(os.Stderr, "Mode:     VBR V%.0f (LAME quality q=%d)\n", opts.VBRQuality, opts.Quality)
	} else {
		fmt.Fprintf(os.Stderr, "Mode:     CBR %d kbps (LAME quality q=%d)\n", opts.Bitrate, opts.Quality)
	}
	if opts.Tags.Cover != "" {
		fmt.Fprintf(os.Stderr, "Cover:    %s\n", opts.Tags.Cover)
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

	fmt.Printf("\nInput:  %s (%d Hz, %s, %d-bit, %dm %02ds, %s)\n",
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

	fmt.Printf("Output: %s (%s, elapsed %.1fs, %s, compression %.2fx)\n",
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
		fmt.Printf("Tags:   %s\n", strings.Join(tagParts, ", "))
	}
}

// formatSize adaptsively formats size: B/KB/MB.
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
