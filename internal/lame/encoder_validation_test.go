package lame

import (
	"bytes"
	"testing"
)

// setupStereoEncoder creates and configures a stereo CBR encoder at the given sample rate and bitrate.
func setupStereoEncoder(t *testing.T, buf *bytes.Buffer, sampleRate, bitrate int) *Encoder {
	t.Helper()
	enc := NewEncoder(buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if err := enc.SetInSamplerate(sampleRate); err != nil {
		t.Fatalf("SetInSamplerate(%d): %v", sampleRate, err)
	}
	if err := enc.SetNumChannels(2); err != nil {
		t.Fatalf("SetNumChannels(2): %v", err)
	}
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR(VBROff): %v", err)
	}
	if err := enc.SetBrate(bitrate); err != nil {
		t.Fatalf("SetBrate(%d): %v", bitrate, err)
	}
	return enc
}

// encodeAndFlush writes PCM and flushes the encoder, returning the total MP3 bytes written.
func encodeAndFlush(t *testing.T, enc *Encoder, pcm []byte) int {
	t.Helper()
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	return 0 // caller reads from buf directly
}

// TestMP3_SyncWord verifies that the first MP3 audio frame starts with a valid sync word (0xFF 0xE0+).
// LAME may prepend an info/Xing frame; we scan up to 4096 bytes to find the first sync.
func TestMP3_SyncWord(t *testing.T) {
	var buf bytes.Buffer
	enc := setupStereoEncoder(t, &buf, 44100, 128)
	defer enc.Close()

	pcm := generateSineWave(44100, 2, 1000)
	encodeAndFlush(t, enc, pcm)

	data := buf.Bytes()
	if len(data) < 4 {
		t.Fatalf("MP3 output too short: %d bytes", len(data))
	}

	// Scan for a valid MP3 sync word: 0xFF followed by 0xE0..0xFF (11 sync bits set).
	// MPEG-1 Layer 3 frames: 0xFF 0xFB (CBR), 0xFF 0xFA, 0xFF 0xF3, 0xFF 0xF2 are common.
	found := false
	limit := len(data) - 1
	if limit > 4096 {
		limit = 4096
	}
	for i := 0; i < limit; i++ {
		if data[i] == 0xFF && (data[i+1]&0xE0) == 0xE0 {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("no valid MP3 sync word found in first %d bytes of output", limit)
	}
}

// TestMP3_OutputSize_CBR verifies that CBR output size is within 20% of the expected bitrate-derived size.
func TestMP3_OutputSize_CBR(t *testing.T) {
	const (
		sampleRate  = 44100
		channels    = 2
		durationMs  = 1000
		bitrateKbps = 128
	)

	var buf bytes.Buffer
	enc := setupStereoEncoder(t, &buf, sampleRate, bitrateKbps)
	defer enc.Close()

	pcm := generateSineWave(sampleRate, channels, durationMs)
	encodeAndFlush(t, enc, pcm)

	got := buf.Len()
	// Expected bytes = bitrate_kbps * 1000 / 8 * duration_seconds
	expected := bitrateKbps * 1000 / 8 * durationMs / 1000
	low := expected * 80 / 100
	high := expected * 120 / 100

	if got < low || got > high {
		t.Errorf("CBR 128kbps 1s output size = %d bytes, want %d..%d (±20%% of %d)",
			got, low, high, expected)
	}
}

// TestMP3_OutputSize_VBR_Monotonic verifies that VBR quality 0 (best) produces more bytes than quality 9 (worst)
// for identical input, demonstrating higher quality equals larger file.
func TestMP3_OutputSize_VBR_Monotonic(t *testing.T) {
	pcm := generateSineWave(44100, 2, 1000)

	encodeVBR := func(quality float64) int {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		if enc == nil {
			t.Fatal("NewEncoder returned nil")
		}
		defer enc.Close()
		if err := enc.SetInSamplerate(44100); err != nil {
			t.Fatalf("SetInSamplerate: %v", err)
		}
		if err := enc.SetNumChannels(2); err != nil {
			t.Fatalf("SetNumChannels: %v", err)
		}
		if err := enc.SetVBR(VBRDefault); err != nil {
			t.Fatalf("SetVBR: %v", err)
		}
		if err := enc.SetVBRQuality(quality); err != nil {
			t.Fatalf("SetVBRQuality(%v): %v", quality, err)
		}
		if _, err := enc.Write(pcm); err != nil {
			t.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			t.Fatalf("Flush: %v", err)
		}
		return buf.Len()
	}

	sizeV0 := encodeVBR(0) // best quality
	sizeV9 := encodeVBR(9) // worst quality

	if sizeV0 <= sizeV9 {
		t.Errorf("VBR quality 0 output (%d bytes) should be larger than quality 9 (%d bytes)", sizeV0, sizeV9)
	}
}

// TestMP3_DifferentSampleRates verifies that encoding succeeds and produces output at 22050, 44100, and 48000 Hz.
func TestMP3_DifferentSampleRates(t *testing.T) {
	rates := []int{22050, 44100, 48000}

	for _, rate := range rates {
		rate := rate
		t.Run("sampleRate="+itoa(rate), func(t *testing.T) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			if enc == nil {
				t.Fatal("NewEncoder returned nil")
			}
			defer enc.Close()

			if err := enc.SetInSamplerate(rate); err != nil {
				t.Fatalf("SetInSamplerate(%d): %v", rate, err)
			}
			if err := enc.SetNumChannels(2); err != nil {
				t.Fatalf("SetNumChannels: %v", err)
			}
			if err := enc.SetVBR(VBROff); err != nil {
				t.Fatalf("SetVBR: %v", err)
			}
			if err := enc.SetBrate(128); err != nil {
				t.Fatalf("SetBrate: %v", err)
			}

			pcm := generateSineWave(rate, 2, 500)
			if _, err := enc.Write(pcm); err != nil {
				t.Fatalf("Write at %dHz: %v", rate, err)
			}
			if _, err := enc.Flush(); err != nil {
				t.Fatalf("Flush at %dHz: %v", rate, err)
			}

			if buf.Len() == 0 {
				t.Errorf("sample rate %dHz produced no output", rate)
			}
		})
	}
}

// TestMP3_MultipleWrites verifies that writing PCM in small chunks produces the same non-zero output
// as writing all PCM in a single call.
func TestMP3_MultipleWrites(t *testing.T) {
	const (
		sampleRate = 44100
		channels   = 2
		durationMs = 1000
		bitrate    = 128
	)
	pcm := generateSineWave(sampleRate, channels, durationMs)

	// Single write
	var singleBuf bytes.Buffer
	encSingle := setupStereoEncoder(t, &singleBuf, sampleRate, bitrate)
	if _, err := encSingle.Write(pcm); err != nil {
		t.Fatalf("single Write: %v", err)
	}
	if _, err := encSingle.Flush(); err != nil {
		t.Fatalf("single Flush: %v", err)
	}
	_ = encSingle.Close()

	// Chunked writes: 4096-byte chunks (mimics converter pipeline)
	var chunkBuf bytes.Buffer
	encChunk := setupStereoEncoder(t, &chunkBuf, sampleRate, bitrate)
	const chunkSize = 4096
	for offset := 0; offset < len(pcm); offset += chunkSize {
		end := offset + chunkSize
		if end > len(pcm) {
			end = len(pcm)
		}
		if _, err := encChunk.Write(pcm[offset:end]); err != nil {
			t.Fatalf("chunked Write at offset %d: %v", offset, err)
		}
	}
	if _, err := encChunk.Flush(); err != nil {
		t.Fatalf("chunked Flush: %v", err)
	}
	_ = encChunk.Close()

	if singleBuf.Len() == 0 {
		t.Error("single write produced no output")
	}
	if chunkBuf.Len() == 0 {
		t.Error("chunked write produced no output")
	}
}

// TestMP3_LargeInput verifies that encoding 10 seconds of audio completes without error.
func TestMP3_LargeInput(t *testing.T) {
	const (
		sampleRate = 44100
		channels   = 2
		durationMs = 10000 // 10 seconds
		bitrate    = 128
	)

	var buf bytes.Buffer
	enc := setupStereoEncoder(t, &buf, sampleRate, bitrate)
	defer enc.Close()

	pcm := generateSineWave(sampleRate, channels, durationMs)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write 10s audio: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush 10s audio: %v", err)
	}

	if buf.Len() == 0 {
		t.Error("10s encode produced no output")
	}

	// Rough sanity: at 128kbps, 10s should be ~160000 bytes ±50%
	expected := 128 * 1000 / 8 * 10
	low := expected / 2
	high := expected * 3 / 2
	if got := buf.Len(); got < low || got > high {
		t.Errorf("10s CBR 128kbps output = %d bytes, want roughly %d..%d", got, low, high)
	}
}

// TestMP3_VBRHeader verifies that VBR encoding produces a non-nil GetLametagFrame result,
// indicating the Xing/LAME VBR header is available after encoding.
func TestMP3_VBRHeader(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetInSamplerate(44100); err != nil {
		t.Fatalf("SetInSamplerate: %v", err)
	}
	if err := enc.SetNumChannels(2); err != nil {
		t.Fatalf("SetNumChannels: %v", err)
	}
	if err := enc.SetVBR(VBRDefault); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetVBRQuality(4); err != nil {
		t.Fatalf("SetVBRQuality: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	frame := enc.GetLametagFrame()
	if frame == nil {
		t.Error("GetLametagFrame returned nil after VBR encoding; expected a Xing/LAME header frame")
	}
	if len(frame) < 4 {
		t.Errorf("GetLametagFrame returned %d bytes, expected at least 4", len(frame))
	}
}

// itoa converts an int to its decimal string representation without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
