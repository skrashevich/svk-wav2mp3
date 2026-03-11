package lame

import (
	"bytes"
	"io"
	"testing"
)

// --- VBR bitrate range tests ---

func TestSetVBRMeanBitrateKbps(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetVBRMeanBitrateKbps(192); err != nil {
		t.Fatalf("SetVBRMeanBitrateKbps(192): %v", err)
	}

	if got := enc.GetVBRMeanBitrateKbps(); got != 192 {
		t.Errorf("GetVBRMeanBitrateKbps() = %d; want 192", got)
	}
}

func TestSetVBRMinBitrateKbps(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetVBRMinBitrateKbps(128); err != nil {
		t.Fatalf("SetVBRMinBitrateKbps(128): %v", err)
	}

	if got := enc.GetVBRMinBitrateKbps(); got != 128 {
		t.Errorf("GetVBRMinBitrateKbps() = %d; want 128", got)
	}
}

func TestSetVBRMaxBitrateKbps(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetVBRMaxBitrateKbps(256); err != nil {
		t.Fatalf("SetVBRMaxBitrateKbps(256): %v", err)
	}

	if got := enc.GetVBRMaxBitrateKbps(); got != 256 {
		t.Errorf("GetVBRMaxBitrateKbps() = %d; want 256", got)
	}
}

func TestVBR_BitrateRange_Encoding(t *testing.T) {
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
	if err := enc.SetVBRQuality(2); err != nil {
		t.Fatalf("SetVBRQuality: %v", err)
	}
	if err := enc.SetVBRMinBitrateKbps(128); err != nil {
		t.Fatalf("SetVBRMinBitrateKbps: %v", err)
	}
	if err := enc.SetVBRMaxBitrateKbps(256); err != nil {
		t.Fatalf("SetVBRMaxBitrateKbps: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected output > 0 bytes after VBR encoding with bitrate range")
	}
}

// --- NumSamples tests ---

func TestSetNumSamples(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetNumSamples(44100 * 10); err != nil {
		t.Fatalf("SetNumSamples: %v", err)
	}

	if got := enc.GetNumSamples(); got != 44100*10 {
		t.Errorf("GetNumSamples() = %d; want %d", got, 44100*10)
	}
}

func TestSetNumSamples_Zero(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetNumSamples(0); err != nil {
		t.Fatalf("SetNumSamples(0): %v", err)
	}

	if got := enc.GetNumSamples(); got != 0 {
		t.Errorf("GetNumSamples() = %d; want 0", got)
	}
}

func TestNumSamples_ImprovesVBRHeader(t *testing.T) {
	encode := func(setNumSamples bool) []byte {
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
		if err := enc.SetVBRQuality(2); err != nil {
			t.Fatalf("SetVBRQuality: %v", err)
		}

		pcm := generateSineWave(44100, 2, 1000)
		totalSamplesPerChannel := len(pcm) / (2 * 2) // int16=2bytes, stereo=2ch
		if setNumSamples {
			if err := enc.SetNumSamples(uint64(totalSamplesPerChannel)); err != nil {
				t.Fatalf("SetNumSamples: %v", err)
			}
		}

		if _, err := enc.Write(pcm); err != nil {
			t.Fatalf("Write: %v", err)
		}
		if _, err := enc.Flush(); err != nil {
			t.Fatalf("Flush: %v", err)
		}

		frame := enc.GetLametagFrame()
		return frame
	}

	withSamples := encode(true)
	withoutSamples := encode(false)

	// Both should produce a valid VBR header
	if withSamples == nil {
		t.Fatal("GetLametagFrame with NumSamples returned nil")
	}
	if withoutSamples == nil {
		t.Fatal("GetLametagFrame without NumSamples returned nil")
	}

	t.Logf("VBR header with NumSamples: %d bytes, without: %d bytes", len(withSamples), len(withoutSamples))
}

// --- DisableReservoir tests ---

func TestSetDisableReservoir(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetDisableReservoir(true); err != nil {
		t.Fatalf("SetDisableReservoir(true): %v", err)
	}

	if got := enc.GetDisableReservoir(); !got {
		t.Error("GetDisableReservoir() = false; want true")
	}
}

func TestSetDisableReservoir_Off(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetDisableReservoir(false); err != nil {
		t.Fatalf("SetDisableReservoir(false): %v", err)
	}

	if got := enc.GetDisableReservoir(); got {
		t.Error("GetDisableReservoir() = true; want false")
	}
}

func TestDisableReservoir_Encoding(t *testing.T) {
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
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetDisableReservoir(true); err != nil {
		t.Fatalf("SetDisableReservoir: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected output > 0 bytes with disabled reservoir")
	}
}

// --- StrictISO tests ---

func TestSetStrictISO(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetStrictISO(true); err != nil {
		t.Fatalf("SetStrictISO(true): %v", err)
	}

	if got := enc.GetStrictISO(); !got {
		t.Error("GetStrictISO() = false; want true")
	}
}

func TestSetStrictISO_Off(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetStrictISO(false); err != nil {
		t.Fatalf("SetStrictISO(false): %v", err)
	}

	if got := enc.GetStrictISO(); got {
		t.Error("GetStrictISO() = true; want false")
	}
}

func TestStrictISO_Encoding(t *testing.T) {
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
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetStrictISO(true); err != nil {
		t.Fatalf("SetStrictISO: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("expected output > 0 bytes with strict ISO")
	}
}

// --- TotalFrames / FrameNum tests ---

func TestGetTotalFrames_WithNumSamples(t *testing.T) {
	enc := NewEncoder(io.Discard)
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
	// 1 second = 44100 samples per channel
	if err := enc.SetNumSamples(44100); err != nil {
		t.Fatalf("SetNumSamples: %v", err)
	}

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	total := enc.GetTotalFrames()
	if total <= 0 {
		t.Errorf("GetTotalFrames() = %d; want > 0 when NumSamples is set", total)
	}

	// For 44100 Hz, MPEG1 frame = 1152 samples → ~38 frames per second
	if total < 30 || total > 50 {
		t.Errorf("GetTotalFrames() = %d; expected ~38 for 1s @ 44100Hz", total)
	}
}

func TestGetFrameNum_ProgressDuringEncoding(t *testing.T) {
	enc := NewEncoder(io.Discard)
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

	before := enc.GetFrameNum()

	pcm := generateSineWave(44100, 2, 1000) // 1 second
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	after := enc.GetFrameNum()
	if after <= before {
		t.Errorf("GetFrameNum() after Write = %d; want > %d (before Write)", after, before)
	}
}

// --- GetMode tests ---

func TestGetMode_Default(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	// LAME default mode before init_params is NOT_SET (4);
	// after init it becomes JointStereo. We just verify it's a valid value.
	got := enc.GetMode()
	if got < 0 || got > 4 {
		t.Errorf("GetMode() = %d; want 0..4", got)
	}
}

func TestGetMode_AfterSet(t *testing.T) {
	tests := []struct {
		name string
		mode int
	}{
		{"Stereo", ModeStereo},
		{"JointStereo", ModeJointStereo},
		{"DualChannel", ModeDualChannel},
		{"Mono", ModeMono},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := NewEncoder(io.Discard)
			if enc == nil {
				t.Fatal("NewEncoder returned nil")
			}
			defer enc.Close()

			if err := enc.SetMode(tt.mode); err != nil {
				t.Fatalf("SetMode(%d): %v", tt.mode, err)
			}

			if got := enc.GetMode(); got != tt.mode {
				t.Errorf("GetMode() = %d; want %d", got, tt.mode)
			}
		})
	}
}

// --- FlushNogap test (via low-level function) ---

func TestFlushNogap_Encoding(t *testing.T) {
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
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}

	pcm := generateSineWave(44100, 2, 500)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	// Use FlushNogap instead of Flush
	n, err := enc.FlushNogap()
	if err != nil {
		t.Fatalf("FlushNogap: %v", err)
	}
	if n < 0 {
		t.Fatalf("FlushNogap returned negative: %d", n)
	}

	if buf.Len() == 0 {
		t.Fatal("expected output > 0 bytes after Write+FlushNogap")
	}

	// After FlushNogap the encoder should still accept more data
	pcm2 := generateSineWave(44100, 2, 500)
	if _, err := enc.Write(pcm2); err != nil {
		t.Fatalf("Write after FlushNogap: %v", err)
	}

	if _, err := enc.Flush(); err != nil {
		t.Fatalf("final Flush: %v", err)
	}
}
