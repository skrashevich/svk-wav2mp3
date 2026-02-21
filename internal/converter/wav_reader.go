package converter

import (
	"fmt"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const chunkSamples = 4096 // сэмплов на канал за один чанк (итого * numChannels в буфере)

// WAVInfo содержит метаданные WAV-файла.
type WAVInfo struct {
	SampleRate  int
	NumChannels int
	BitDepth    int
	Duration    time.Duration
	FileSizeB   int64
}

// WAVReader читает WAV-файл и предоставляет PCM-данные в формате int16 чанками.
type WAVReader struct {
	file    *os.File
	decoder *wav.Decoder
	pcmBuf  *audio.IntBuffer // переиспользуемый буфер для чанкового чтения
	Info    WAVInfo
}

// NewWAVReader открывает WAV-файл и читает заголовок.
func NewWAVReader(path string) (*WAVReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть WAV файл: %w", err)
	}

	dec := wav.NewDecoder(f)
	dec.ReadInfo()
	if !dec.IsValidFile() {
		f.Close()
		return nil, fmt.Errorf("некорректный WAV файл: %s", path)
	}

	fi, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, err
	}

	dur, err := dec.Duration()
	if err != nil {
		dur = 0
	}

	info := WAVInfo{
		SampleRate:  int(dec.SampleRate),
		NumChannels: int(dec.NumChans),
		BitDepth:    int(dec.BitDepth),
		Duration:    dur,
		FileSizeB:   fi.Size(),
	}

	numChans := int(dec.NumChans)
	if numChans < 1 {
		numChans = 1
	}

	// Буфер на chunkSamples фреймов * numChannels сэмплов
	pcmBuf := &audio.IntBuffer{
		Data: make([]int, chunkSamples*numChans),
		Format: &audio.Format{
			NumChannels: numChans,
			SampleRate:  int(dec.SampleRate),
		},
		SourceBitDepth: int(dec.BitDepth),
	}

	return &WAVReader{
		file:    f,
		decoder: dec,
		pcmBuf:  pcmBuf,
		Info:    info,
	}, nil
}

// Close закрывает файл.
func (r *WAVReader) Close() error {
	return r.file.Close()
}

// ReadSamplesInt16 читает следующий чанк PCM-данных в формате int16 interleaved.
// Возвращает nil, nil при EOF.
func (r *WAVReader) ReadSamplesInt16() ([]int16, error) {
	n, err := r.decoder.PCMBuffer(r.pcmBuf)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения PCM: %w", err)
	}
	if n == 0 {
		return nil, nil // EOF
	}
	samples := normalizePCM(r.pcmBuf.Data[:n], r.Info.BitDepth)
	return samples, nil
}

// normalizePCM конвертирует PCM-сэмплы различной битности в int16.
func normalizePCM(samples []int, bitDepth int) []int16 {
	out := make([]int16, len(samples))
	for i, s := range samples {
		switch bitDepth {
		case 8:
			// 8-bit unsigned (0–255) → int16 со знаком
			out[i] = int16((s - 128) << 8)
		case 16:
			out[i] = int16(s)
		case 24:
			// 24-bit → 16-bit: отбрасываем 8 младших бит
			out[i] = int16(s >> 8)
		case 32:
			// 32-bit → 16-bit: отбрасываем 16 младших бит
			out[i] = int16(s >> 16)
		default:
			out[i] = int16(s)
		}
	}
	return out
}

// TotalSamples возвращает общее число фреймов на канал (для progressbar).
func (r *WAVReader) TotalSamples() int {
	if r.Info.Duration == 0 || r.Info.SampleRate == 0 {
		return 0
	}
	return int(r.Info.Duration.Seconds() * float64(r.Info.SampleRate))
}
