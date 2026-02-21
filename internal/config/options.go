package config

// ID3Tags содержит метаданные для MP3-файла.
type ID3Tags struct {
	Title   string
	Artist  string
	Album   string
	Year    string
	Genre   string
	Track   string
	Comment string
	Cover   string // путь к файлу обложки
}

// ConvertOptions содержит все параметры конвертации.
type ConvertOptions struct {
	InputPath  string
	OutputPath string // пустая строка — генерировать из InputPath

	Tags ID3Tags

	// Режим кодирования
	VBR        bool
	VBRQuality float64 // 0.0 (лучше) — 9.9 (хуже), только при VBR
	Bitrate    int     // CBR в kbps (32–320), только при !VBR
	Quality    int     // алгоритмическое качество LAME 0–9 (0=лучше)

	Verbose bool
	Quiet   bool
}

// DefaultConvertOptions возвращает параметры по умолчанию: VBR V2, quality 2.
func DefaultConvertOptions() ConvertOptions {
	return ConvertOptions{
		VBR:        true,
		VBRQuality: 2.0,
		Quality:    2,
	}
}
