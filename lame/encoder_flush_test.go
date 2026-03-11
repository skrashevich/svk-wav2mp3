package lame

import (
	"bytes"
	"testing"
)

// TestFlush_AfterEncoding verifies that Flush returns valid data after encoding PCM.
func TestFlush_AfterEncoding(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 1000) // 1 second stereo
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}

	n, err := enc.Flush()
	if err != nil {
		t.Fatalf("Flush returned error: %v", err)
	}
	if n < 0 {
		t.Fatalf("Flush returned negative n: %d", n)
	}

	if buf.Len() == 0 {
		t.Fatal("expected total output > 0 bytes after Write+Flush")
	}
}

// TestFlush_WithoutWrite verifies Flush doesn't panic when called without any Write.
func TestFlush_WithoutWrite(t *testing.T) {
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

	// Flush without any prior Write — should not panic, returns gracefully.
	// Note: LAME may return -1 (LAME_NOGAP_NOEND) when flushing without writing;
	// the encoder wrapper checks for ret < 0 and returns an error in that case.
	// We only verify no panic occurs and the signature is correct.
	_, _ = enc.Flush()
}

// TestGetLametagFrame_AfterEncoding verifies that GetLametagFrame returns a valid
// VBR header frame after encoding in VBR mode.
func TestGetLametagFrame_AfterEncoding(t *testing.T) {
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
	if err := enc.SetQuality(2); err != nil {
		t.Fatalf("SetQuality: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000) // 1 second stereo
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	frame := enc.GetLametagFrame()
	if frame == nil {
		t.Fatal("GetLametagFrame returned nil; expected non-nil VBR header")
	}
	if len(frame) == 0 {
		t.Fatal("GetLametagFrame returned empty slice; expected length > 0")
	}
}

// TestGetLametagFrame_CBR verifies GetLametagFrame behavior in CBR mode.
// LAME may or may not produce a Xing/Info frame in CBR mode depending on configuration.
// This test documents the actual behavior: the frame is either nil (no CBR tag written)
// or a valid non-empty byte slice (Info tag present).
func TestGetLametagFrame_CBR(t *testing.T) {
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

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	frame := enc.GetLametagFrame()
	// In CBR mode, LAME typically returns nil from lame_get_lametag_frame
	// because there is no VBR seek table to write. If a non-nil frame is
	// returned it must have length > 0.
	if frame != nil && len(frame) == 0 {
		t.Fatal("GetLametagFrame returned non-nil but empty slice in CBR mode")
	}
	t.Logf("CBR GetLametagFrame result: %v (len=%d)", frame != nil, len(frame))
}

// TestFlush_AfterClose verifies that calling Flush after Close returns an error.
func TestFlush_AfterClose(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}

	if err := enc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := enc.Flush()
	if err == nil {
		t.Fatal("expected error from Flush after Close, got nil")
	}
	if err.Error() != "encoder is closed" {
		t.Fatalf("expected 'encoder is closed', got %q", err.Error())
	}
}
