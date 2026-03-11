package converter

import (
	"fmt"
	"os"
	"time"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

const chunkSamples = 4096 // samples per channel per chunk (total count = 4096 * numChannels in buffer)

// WAVInfo contains WAV file metadata.
type WAVInfo struct {
	SampleRate  int
	NumChannels int
	BitDepth    int
	Duration    time.Duration
	FileSizeB   int64
}

// WAVReader reads a WAV file and provides PCM data in int16 chunks.
type WAVReader struct {
	file      *os.File
	decoder   *wav.Decoder
	pcmBuf    *audio.IntBuffer // reusable buffer for chunked reading
	int16Buf  []int16          // reusable buffer for normalizePCM output
	Info      WAVInfo
}

// NewWAVReader opens a WAV file and reads header.
func NewWAVReader(path string) (*WAVReader, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open WAV file: %w", err)
	}

	dec := wav.NewDecoder(f)
	dec.ReadInfo()
	if !dec.IsValidFile() {
		f.Close()
		return nil, fmt.Errorf("invalid WAV file: %s", path)
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

	numChans := max(int(dec.NumChans), 1)

	// Buffer for chunkSamples frames * numChannels samples
	pcmBuf := &audio.IntBuffer{
		Data: make([]int, chunkSamples*numChans),
		Format: &audio.Format{
			NumChannels: numChans,
			SampleRate:  int(dec.SampleRate),
		},
		SourceBitDepth: int(dec.BitDepth),
	}

	return &WAVReader{
		file:     f,
		decoder:  dec,
		pcmBuf:   pcmBuf,
		int16Buf: make([]int16, chunkSamples*numChans),
		Info:     info,
	}, nil
}

// Close closes the file.
func (r *WAVReader) Close() error {
	return r.file.Close()
}

// ReadSamplesInt16 reads next chunk of PCM data in int16 interleaved format.
// Returns nil, nil on EOF.
func (r *WAVReader) ReadSamplesInt16() ([]int16, error) {
	n, err := r.decoder.PCMBuffer(r.pcmBuf)
	if err != nil {
		return nil, fmt.Errorf("error reading PCM: %w", err)
	}
	if n == 0 {
		return nil, nil // EOF
	}
	normalizePCM(r.int16Buf[:n], r.pcmBuf.Data[:n], r.Info.BitDepth)
	return r.int16Buf[:n], nil
}

// normalizePCM converts PCM samples of various bit depths to int16.
// Writes into dst to avoid per-chunk allocation.
func normalizePCM(dst []int16, samples []int, bitDepth int) {
	for i, s := range samples {
		switch bitDepth {
		case 8:
			// 8-bit unsigned (0–255) → signed int16 (0–32767 after offset)
			dst[i] = int16((s - 128) << 8)
		case 16:
			dst[i] = int16(s)
		case 24:
			// 24-bit → 16-bit: drop 8 least significant bits
			dst[i] = int16(s >> 8)
		case 32:
			// 32-bit → 16-bit: drop 16 least significant bits
			dst[i] = int16(s >> 16)
		default:
			dst[i] = int16(s)
		}
	}
}

// TotalFrames returns total number of frames per channel (for progressbar).
func (r *WAVReader) TotalSamples() int {
	if r.Info.Duration == 0 || r.Info.SampleRate == 0 {
		return 0
	}
	return int(r.Info.Duration.Seconds() * float64(r.Info.SampleRate))
}
