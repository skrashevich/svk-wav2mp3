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

// BenchmarkFlush benchmarks the full encode+flush cycle, focusing on flush overhead.
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
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
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

// BenchmarkEncode_CBR128_NoReservoir benchmarks CBR 128 with bit reservoir disabled.
// Compare with BenchmarkEncode_CBR128_Stereo to measure the cost of disabling reservoir.
func BenchmarkEncode_CBR128_NoReservoir(b *testing.B) {
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
		_ = enc.SetDisableReservoir(true)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkEncode_VBR_V2_Clamped benchmarks VBR V2 with min/max bitrate limits.
// Compare with BenchmarkEncode_VBR_V4 to measure the overhead of bitrate clamping.
func BenchmarkEncode_VBR_V2_Clamped(b *testing.B) {
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
		_ = enc.SetVBRQuality(2)
		_ = enc.SetVBRMinBitrateKbps(128)
		_ = enc.SetVBRMaxBitrateKbps(256)
		if _, err := enc.Write(pcm); err != nil {
			b.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			b.Fatalf("Flush: %v", err)
		}
		_ = enc.Close()
	}
}

// BenchmarkFlushNogap benchmarks encode+nogap flush for gapless encoding scenarios.
// Compare with BenchmarkFlush to measure the difference.
func BenchmarkFlushNogap(b *testing.B) {
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
		if _, err := enc.FlushNogap(); err != nil {
			b.Fatalf("FlushNogap: %v", err)
		}
		_ = enc.Close()
	}
}
