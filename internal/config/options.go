package config

// ID3Tags contains metadata for MP3 file.
type ID3Tags struct {
	Title   string
	Artist  string
	Album   string
	Year    string
	Genre   string
	Track   string
	Comment string
	Cover   string // path to cover file
}

// ConvertOptions contains all conversion parameters.
type ConvertOptions struct {
	InputPath  string
	OutputPath string // empty string - generate from InputPath

	Tags ID3Tags

	// Encoding mode
	VBR        bool
	VBRQuality float64 // 0.0 (better) – 9.9 (worse), VBR only
	Bitrate    int     // CBR in kbps (32–320), VBR only when !VBR
	Quality    int     // LAME algorithmic quality 0–9 (0=better)

	Verbose bool
	Quiet   bool
}

// DefaultConvertOptions returns default parameters: VBR V2, quality 2.
func DefaultConvertOptions() ConvertOptions {
	return ConvertOptions{
		VBR:        true,
		VBRQuality: 2.0,
		Quality:    2,
	}
}
