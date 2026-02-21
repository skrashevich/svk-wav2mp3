//go:build ignore

// Generates WAV files in testdata/fixtures/ for integration tests.
// Run: go run testdata/gen_fixtures.go
package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata/fixtures"
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fixtures := []struct {
		name        string
		sampleRate  int
		numChannels int
		bitDepth    int
		duration    float64
	}{
		{"sine_44100_stereo_16.wav", 44100, 2, 16, 2.0},
		{"sine_22050_mono_8.wav", 22050, 1, 8, 1.0},
		{"sine_48000_stereo_16.wav", 48000, 2, 16, 1.0},
		{"sine_44100_stereo_24.wav", 44100, 2, 24, 1.0},
	}

	for _, fx := range fixtures {
		path := filepath.Join(dir, fx.name)
		if err := writeWAV(path, fx.sampleRate, fx.numChannels, fx.bitDepth, fx.duration); err != nil {
			fmt.Fprintf(os.Stderr, "error creating %s: %v\n", path, err)
			os.Exit(1)
		}
		fi, _ := os.Stat(path)
		fmt.Printf("created: %s (%.1f KB)\n", path, float64(fi.Size())/1024)
	}
}

func writeWAV(path string, sampleRate, numChannels, bitDepth int, durationSec float64) error {
	numSamples := int(float64(sampleRate) * durationSec)
	bytesPerSample := (bitDepth + 7) / 8
	dataSize := numSamples * numChannels * bytesPerSample

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// RIFF header
	f.Write([]byte("RIFF"))
	binary.Write(f, binary.LittleEndian, uint32(36+dataSize))
	f.Write([]byte("WAVE"))

	// fmt chunk
	f.Write([]byte("fmt "))
	binary.Write(f, binary.LittleEndian, uint32(16))
	binary.Write(f, binary.LittleEndian, uint16(1)) // PCM
	binary.Write(f, binary.LittleEndian, uint16(numChannels))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate))
	binary.Write(f, binary.LittleEndian, uint32(sampleRate*numChannels*bytesPerSample))
	binary.Write(f, binary.LittleEndian, uint16(numChannels*bytesPerSample))
	binary.Write(f, binary.LittleEndian, uint16(bitDepth))

	// data chunk
	f.Write([]byte("data"))
	binary.Write(f, binary.LittleEndian, uint32(dataSize))

	const freq = 440.0
	buf3 := make([]byte, 3)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		switch bitDepth {
		case 8:
			s := uint8(math.Sin(2*math.Pi*freq*t)*127 + 128)
			for ch := 0; ch < numChannels; ch++ {
				f.Write([]byte{s})
			}
		case 16:
			s := int16(math.Sin(2*math.Pi*freq*t) * 32000)
			for ch := 0; ch < numChannels; ch++ {
				binary.Write(f, binary.LittleEndian, s)
			}
		case 24:
			// 24-bit little-endian signed
			s := int32(math.Sin(2*math.Pi*freq*t) * 8388607)
			buf3[0] = byte(s)
			buf3[1] = byte(s >> 8)
			buf3[2] = byte(s >> 16)
			for ch := 0; ch < numChannels; ch++ {
				f.Write(buf3)
			}
		}
	}
	return nil
}
