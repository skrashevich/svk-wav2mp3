package converter

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/viert/go-lame"
)

// MP3Writer управляет LAME-энкодером и записью MP3-данных в файл.
type MP3Writer struct {
	file    *os.File
	encoder *lame.Encoder
}

// EncoderConfig параметры энкодера.
type EncoderConfig struct {
	SampleRate  int
	NumChannels int
	VBR         bool
	VBRQuality  float64 // 0.0–9.9
	Bitrate     int     // kbps, только CBR
	Quality     int     // алгоритмическое 0–9
}

// NewMP3Writer создаёт файл и инициализирует LAME-энкодер.
func NewMP3Writer(path string, cfg EncoderConfig) (*MP3Writer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать MP3 файл: %w", err)
	}

	enc := lame.NewEncoder(f)

	// Отключаем автоматическую запись ID3-тегов LAME —
	// теги будем писать вручную через bogem/id3v2 после кодирования.
	enc.SetWriteID3TagAutomatic(false)

	if err := enc.SetInSamplerate(cfg.SampleRate); err != nil {
		f.Close()
		return nil, fmt.Errorf("SetInSamplerate: %w", err)
	}
	if err := enc.SetNumChannels(cfg.NumChannels); err != nil {
		f.Close()
		return nil, fmt.Errorf("SetNumChannels: %w", err)
	}
	if err := enc.SetQuality(cfg.Quality); err != nil {
		f.Close()
		return nil, fmt.Errorf("SetQuality: %w", err)
	}

	if cfg.VBR {
		if err := enc.SetVBR(lame.VBRDefault); err != nil {
			f.Close()
			return nil, fmt.Errorf("SetVBR: %w", err)
		}
		if err := enc.SetVBRQuality(cfg.VBRQuality); err != nil {
			f.Close()
			return nil, fmt.Errorf("SetVBRQuality: %w", err)
		}
	} else {
		if err := enc.SetVBR(lame.VBROff); err != nil {
			f.Close()
			return nil, fmt.Errorf("SetVBR(off): %w", err)
		}
		if err := enc.SetBrate(cfg.Bitrate); err != nil {
			f.Close()
			return nil, fmt.Errorf("SetBrate: %w", err)
		}
	}

	return &MP3Writer{
		file:    f,
		encoder: enc,
	}, nil
}

// WriteSamples кодирует и записывает чанк int16-сэмплов.
// go-lame.Write принимает []byte (int16 little-endian interleaved).
func (w *MP3Writer) WriteSamples(samples []int16) error {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}
	if _, err := w.encoder.Write(buf); err != nil {
		return fmt.Errorf("ошибка кодирования: %w", err)
	}
	return nil
}

// Close сбрасывает буферы LAME и закрывает файл.
func (w *MP3Writer) Close() error {
	if _, err := w.encoder.Flush(); err != nil {
		return fmt.Errorf("ошибка flush LAME: %w", err)
	}
	w.encoder.Close()
	return w.file.Close()
}

// FilePath возвращает путь к выходному файлу.
func (w *MP3Writer) FilePath() string {
	return w.file.Name()
}
