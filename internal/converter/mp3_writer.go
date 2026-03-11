package converter

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/svk/wav2mp3/lame"
)

// MP3Writer manages LAME encoder and MP3 data writing to file.
type MP3Writer struct {
	file    *os.File
	encoder *lame.Encoder // pure-Go transpiled LAME encoder
}

// EncoderConfig encoder parameters.
type EncoderConfig struct {
	SampleRate  int
	NumChannels int
	VBR         bool
	VBRQuality  float64 // 0.0–9.9
	Bitrate     int     // kbps, CBR only
	Quality     int     // algorithmic 0–9
}

// NewMP3Writer creates file and initializes LAME encoder.
func NewMP3Writer(path string, cfg EncoderConfig) (*MP3Writer, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create MP3 file: %w", err)
	}

	enc := lame.NewEncoder(f)

	// Disable automatic ID3 tag writing by LAME
	// We'll write tags manually via bogem/id3v2 after encoding.
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

// WriteSamples encodes and writes chunk of int16 samples.
// go-lame.Write expects []byte (int16 little-endian interleaved).
func (w *MP3Writer) WriteSamples(samples []int16) error {
	buf := make([]byte, len(samples)*2)
	for i, s := range samples {
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(s))
	}
	if _, err := w.encoder.Write(buf); err != nil {
		return fmt.Errorf("encoding error: %w", err)
	}
	return nil
}

// Close flushes LAME buffers, writes the VBR/Xing header, and closes file.
func (w *MP3Writer) Close() error {
	if _, err := w.encoder.Flush(); err != nil {
		return fmt.Errorf("LAME flush error: %w", err)
	}

	// Write LAME/Xing VBR header at the beginning of the file.
	// This header contains accurate duration and bitrate information.
	if tag := w.encoder.GetLametagFrame(); len(tag) > 0 {
		if _, err := w.file.Seek(0, 0); err != nil {
			return fmt.Errorf("seek for lametag: %w", err)
		}
		if _, err := w.file.Write(tag); err != nil {
			return fmt.Errorf("write lametag: %w", err)
		}
	}

	w.encoder.Close()
	return w.file.Close()
}

// FilePath returns output file path.
func (w *MP3Writer) FilePath() string {
	return w.file.Name()
}
