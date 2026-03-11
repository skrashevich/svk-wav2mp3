package lame

import (
	"io"
	"math"
	"testing"
)

// --- Getter tests (after init / after encoding) ---

func TestGetEncoderDelay(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	delay := enc.GetEncoderDelay()
	if delay <= 0 {
		t.Errorf("GetEncoderDelay() = %d; want > 0 after encoding", delay)
	}
}

func TestGetEncoderPadding(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	padding := enc.GetEncoderPadding()
	if padding < 0 {
		t.Errorf("GetEncoderPadding() = %d; want >= 0", padding)
	}
}

func TestGetFrameSize(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	fs := enc.GetFrameSize()
	if fs <= 0 {
		t.Errorf("GetFrameSize() = %d; want > 0 after init", fs)
	}
	// MPEG1 = 1152, MPEG2/2.5 = 576
	if fs != 576 && fs != 1152 {
		t.Errorf("GetFrameSize() = %d; want 576 or 1152", fs)
	}
}

func TestGetVersion(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	v := enc.GetVersion()
	if v != 0 && v != 1 && v != 2 {
		t.Errorf("GetVersion() = %d; want 0, 1, or 2", v)
	}
}

func TestGetOutSamplerate(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	rate := enc.GetOutSamplerate()
	if rate <= 0 {
		t.Errorf("GetOutSamplerate() = %d; want > 0", rate)
	}
}

func TestGetNumChannels_mono(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetNumChannels(1); err != nil {
		t.Fatalf("SetNumChannels(1): %v", err)
	}

	if got := enc.GetNumChannels(); got != 1 {
		t.Errorf("GetNumChannels() = %d; want 1", got)
	}
}

func TestGetNumChannels_stereo(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetNumChannels(2); err != nil {
		t.Fatalf("SetNumChannels(2): %v", err)
	}

	if got := enc.GetNumChannels(); got != 2 {
		t.Errorf("GetNumChannels() = %d; want 2", got)
	}
}

func TestGetInSamplerate(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetInSamplerate(48000); err != nil {
		t.Fatalf("SetInSamplerate(48000): %v", err)
	}

	if got := enc.GetInSamplerate(); got != 48000 {
		t.Errorf("GetInSamplerate() = %d; want 48000", got)
	}
}

func TestGetBrate(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetBrate(320); err != nil {
		t.Fatalf("SetBrate(320): %v", err)
	}

	if got := enc.GetBrate(); got != 320 {
		t.Errorf("GetBrate() = %d; want 320", got)
	}
}

func TestGetVBR_off(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR(VBROff): %v", err)
	}

	if got := enc.GetVBR(); got != VBROff {
		t.Errorf("GetVBR() = %d; want VBROff (%d)", got, VBROff)
	}
}

func TestGetVBR_default(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetVBR(VBRDefault); err != nil {
		t.Fatalf("SetVBR(VBRDefault): %v", err)
	}

	if got := enc.GetVBR(); got != VBRDefault {
		t.Errorf("GetVBR() = %d; want VBRDefault (%d)", got, VBRDefault)
	}
}

func TestGetVBRQuality(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetVBRQuality(2.0); err != nil {
		t.Fatalf("SetVBRQuality(2.0): %v", err)
	}

	got := enc.GetVBRQuality()
	if math.Abs(got-2.0) > 0.01 {
		t.Errorf("GetVBRQuality() = %f; want ~2.0", got)
	}
}

func TestGetQuality(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetQuality(3); err != nil {
		t.Fatalf("SetQuality(3): %v", err)
	}

	if got := enc.GetQuality(); got != 3 {
		t.Errorf("GetQuality() = %d; want 3", got)
	}
}

// --- Setter tests ---

func TestSetOutSamplerate_explicit(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetOutSamplerate(22050); err != nil {
		t.Errorf("SetOutSamplerate(22050) error: %v", err)
	}
}

func TestSetOutSamplerate_auto(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetOutSamplerate(0); err != nil {
		t.Errorf("SetOutSamplerate(0) error: %v", err)
	}
}

func TestSetMode_stereo(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetMode(ModeStereo); err != nil {
		t.Errorf("SetMode(ModeStereo) error: %v", err)
	}
}

func TestSetMode_jointStereo(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetMode(ModeJointStereo); err != nil {
		t.Errorf("SetMode(ModeJointStereo) error: %v", err)
	}
}

func TestSetMode_dualChannel(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetMode(ModeDualChannel); err != nil {
		t.Errorf("SetMode(ModeDualChannel) error: %v", err)
	}
}

func TestSetMode_mono(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetMode(ModeMono); err != nil {
		t.Errorf("SetMode(ModeMono) error: %v", err)
	}
}

func TestSetScale_attenuate(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetScale(0.5); err != nil {
		t.Errorf("SetScale(0.5) error: %v", err)
	}
}

func TestSetScale_amplify(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetScale(2.0); err != nil {
		t.Errorf("SetScale(2.0) error: %v", err)
	}
}

func TestSetLowpassFreq_explicit(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetLowpassFreq(18000); err != nil {
		t.Errorf("SetLowpassFreq(18000) error: %v", err)
	}
}

func TestSetLowpassFreq_auto(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetLowpassFreq(0); err != nil {
		t.Errorf("SetLowpassFreq(0) error: %v", err)
	}
}

func TestSetLowpassFreq_disabled(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetLowpassFreq(-1); err != nil {
		t.Errorf("SetLowpassFreq(-1) error: %v", err)
	}
}

func TestSetHighpassFreq_explicit(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetHighpassFreq(100); err != nil {
		t.Errorf("SetHighpassFreq(100) error: %v", err)
	}
}

func TestSetPreset_medium(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetPreset(PresetMedium); err != nil {
		t.Errorf("SetPreset(PresetMedium) error: %v", err)
	}
}

func TestSetPreset_standard(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetPreset(PresetStandard); err != nil {
		t.Errorf("SetPreset(PresetStandard) error: %v", err)
	}
}

func TestSetPreset_extreme(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetPreset(PresetExtreme); err != nil {
		t.Errorf("SetPreset(PresetExtreme) error: %v", err)
	}
}

func TestSetPreset_insane(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetPreset(PresetInsane); err != nil {
		t.Errorf("SetPreset(PresetInsane) error: %v", err)
	}
}

// --- Integration tests ---

func TestPreset_Standard_Encoding(t *testing.T) {
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
	if err := enc.SetPreset(PresetStandard); err != nil {
		t.Fatalf("SetPreset(PresetStandard): %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	n, err := enc.Write(pcm)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(pcm) {
		t.Errorf("Write() = %d bytes consumed; want %d", n, len(pcm))
	}

	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

func TestMode_Mono_Encoding(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetInSamplerate(44100); err != nil {
		t.Fatalf("SetInSamplerate: %v", err)
	}
	if err := enc.SetNumChannels(1); err != nil {
		t.Fatalf("SetNumChannels(1): %v", err)
	}
	if err := enc.SetMode(ModeMono); err != nil {
		t.Fatalf("SetMode(ModeMono): %v", err)
	}

	pcm := generateSineWave(44100, 1, 500)
	n, err := enc.Write(pcm)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(pcm) {
		t.Errorf("Write() = %d bytes consumed; want %d", n, len(pcm))
	}

	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

func TestFiltering_Lowpass(t *testing.T) {
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
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetLowpassFreq(8000); err != nil {
		t.Fatalf("SetLowpassFreq(8000): %v", err)
	}

	pcm := generateSineWave(44100, 2, 500)
	n, err := enc.Write(pcm)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n == 0 {
		t.Error("Write() consumed 0 bytes; want > 0")
	}

	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
}

func TestGetters_AfterEncoding(t *testing.T) {
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
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(192); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetQuality(5); err != nil {
		t.Fatalf("SetQuality: %v", err)
	}

	pcm := generateSineWave(44100, 2, 200)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if got := enc.GetInSamplerate(); got != 44100 {
		t.Errorf("GetInSamplerate() = %d; want 44100", got)
	}
	if got := enc.GetNumChannels(); got != 2 {
		t.Errorf("GetNumChannels() = %d; want 2", got)
	}
	if got := enc.GetVBR(); got != VBROff {
		t.Errorf("GetVBR() = %d; want VBROff (%d)", got, VBROff)
	}
	if got := enc.GetBrate(); got != 192 {
		t.Errorf("GetBrate() = %d; want 192", got)
	}
	if got := enc.GetQuality(); got != 5 {
		t.Errorf("GetQuality() = %d; want 5", got)
	}
	if got := enc.GetVersion(); got != 0 && got != 1 && got != 2 {
		t.Errorf("GetVersion() = %d; want 0, 1, or 2", got)
	}
	if got := enc.GetFrameSize(); got != 576 && got != 1152 {
		t.Errorf("GetFrameSize() = %d; want 576 or 1152", got)
	}
	if got := enc.GetEncoderDelay(); got <= 0 {
		t.Errorf("GetEncoderDelay() = %d; want > 0", got)
	}
	if got := enc.GetEncoderPadding(); got < 0 {
		t.Errorf("GetEncoderPadding() = %d; want >= 0", got)
	}
	if got := enc.GetOutSamplerate(); got <= 0 {
		t.Errorf("GetOutSamplerate() = %d; want > 0", got)
	}
}
