package lame

import (
	"bytes"
	"testing"
)

// TestWrite_BasicEncoding verifies that writing 1 second of stereo PCM produces MP3 output.
func TestWrite_BasicEncoding(t *testing.T) {
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
	if err := enc.SetQuality(5); err != nil {
		t.Fatalf("SetQuality: %v", err)
	}
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	n, err := enc.Write(pcm)
	if err != nil {
		t.Fatalf("Write: %v", err)
	}
	if n != len(pcm) {
		t.Errorf("Write returned %d, want %d", n, len(pcm))
	}

	if buf.Len() == 0 {
		t.Error("expected MP3 output bytes, got 0")
	}
}

// TestWrite_AutoInitialization verifies that initParams is called automatically on first Write.
func TestWrite_AutoInitialization(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	pcm := generateSineWave(44100, 2, 100)
	_, err := enc.Write(pcm)
	if err != nil {
		t.Fatalf("Write with auto-init failed: %v", err)
	}
}

// TestWrite_EmptyInput verifies that writing empty PCM produces no error.
func TestWrite_EmptyInput(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	_, err := enc.Write([]byte{})
	if err != nil {
		t.Fatalf("Write with empty input: %v", err)
	}
}

// TestWrite_AfterClose verifies that writing after Close returns "encoder is closed" error.
func TestWrite_AfterClose(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}

	if err := enc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := enc.Write([]byte{0, 1, 2, 3})
	if err == nil {
		t.Fatal("expected error after Close, got nil")
	}
	if err.Error() != "encoder is closed" {
		t.Errorf("expected 'encoder is closed', got %q", err.Error())
	}
}

// TestWrite_MonoEncoding verifies that mono encoding works and produces MP3 output.
func TestWrite_MonoEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	if err := enc.SetNumChannels(1); err != nil {
		t.Fatalf("SetNumChannels: %v", err)
	}
	if err := enc.SetInSamplerate(44100); err != nil {
		t.Fatalf("SetInSamplerate: %v", err)
	}
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}

	pcm := generateSineWave(44100, 1, 1000)
	n, err := enc.Write(pcm)
	if err != nil {
		t.Fatalf("Write mono: %v", err)
	}
	if n != len(pcm) {
		t.Errorf("Write returned %d, want %d", n, len(pcm))
	}

	if buf.Len() == 0 {
		t.Error("expected MP3 output bytes for mono encoding, got 0")
	}
}
