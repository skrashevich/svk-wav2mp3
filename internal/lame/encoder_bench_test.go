package lame

import (
	"bytes"
	"testing"
)

// BenchmarkEncode_CBR128_Stereo benchmarks encoding 1s stereo 44100Hz at CBR 128kbps.
func BenchmarkEncode_CBR128_Stereo(b *testing.B) {
	pcm := generateSineWave(44100, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(44100)
		_ = enc.SetNumChannels(2)
		_ = enc.SetVBR(VBROff)
		_ = enc.SetBrate(128)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkEncode_CBR320_Stereo benchmarks encoding 1s stereo 44100Hz at CBR 320kbps.
func BenchmarkEncode_CBR320_Stereo(b *testing.B) {
	pcm := generateSineWave(44100, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(44100)
		_ = enc.SetNumChannels(2)
		_ = enc.SetVBR(VBROff)
		_ = enc.SetBrate(320)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkEncode_VBR_V0 benchmarks encoding 1s stereo 44100Hz at VBR quality 0 (best).
func BenchmarkEncode_VBR_V0(b *testing.B) {
	pcm := generateSineWave(44100, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(44100)
		_ = enc.SetNumChannels(2)
		_ = enc.SetVBR(VBRDefault)
		_ = enc.SetVBRQuality(0)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkEncode_VBR_V4 benchmarks encoding 1s stereo 44100Hz at VBR quality 4 (middle).
func BenchmarkEncode_VBR_V4(b *testing.B) {
	pcm := generateSineWave(44100, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(44100)
		_ = enc.SetNumChannels(2)
		_ = enc.SetVBR(VBRDefault)
		_ = enc.SetVBRQuality(4)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkEncode_Mono benchmarks encoding 1s mono 44100Hz at CBR 128kbps.
func BenchmarkEncode_Mono(b *testing.B) {
	pcm := generateSineWave(44100, 1, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(44100)
		_ = enc.SetNumChannels(1)
		_ = enc.SetVBR(VBROff)
		_ = enc.SetBrate(128)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkEncode_48kHz benchmarks encoding 1s stereo 48000Hz at CBR 256kbps.
func BenchmarkEncode_48kHz(b *testing.B) {
	pcm := generateSineWave(48000, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(48000)
		_ = enc.SetNumChannels(2)
		_ = enc.SetVBR(VBROff)
		_ = enc.SetBrate(256)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkFlush benchmarks the flush operation after encoding 1s of audio.
func BenchmarkFlush(b *testing.B) {
	pcm := generateSineWave(44100, 2, 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.SetInSamplerate(44100)
		_ = enc.SetNumChannels(2)
		_ = enc.SetVBR(VBROff)
		_ = enc.SetBrate(128)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		b.StartTimer()
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		b.StopTimer()
		_ = enc.Close()
	}
}

// BenchmarkNewEncoder benchmarks creating and closing an encoder without encoding.
func BenchmarkNewEncoder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			b.Fatal("NewEncoder returned nil")
		}
		_ = enc.Close()
	}
}
