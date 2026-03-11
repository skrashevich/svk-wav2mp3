package lame

import (
	"bytes"
	"strings"
	"testing"
)

func TestClose_Basic(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func TestClose_Idempotent(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("first Close() returned error: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("second Close() returned error: %v", err)
	}
}

func TestClose_AfterEncoding(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if err := enc.SetInSamplerate(44100); err != nil {
		t.Fatalf("SetInSamplerate: %v", err)
	}
	if err := enc.SetNumChannels(2); err != nil {
		t.Fatalf("SetNumChannels: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestWriteAfterClose(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if err := enc.SetInSamplerate(44100); err != nil {
		t.Fatalf("SetInSamplerate: %v", err)
	}
	if err := enc.SetNumChannels(2); err != nil {
		t.Fatalf("SetNumChannels: %v", err)
	}
	if err := enc.SetBrate(128); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}

	pcm := generateSineWave(44100, 2, 100)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write before close: %v", err)
	}
	if err := enc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	_, err := enc.Write(pcm)
	if err == nil {
		t.Fatal("expected error writing after Close, got nil")
	}
	if !strings.Contains(err.Error(), "encoder is closed") {
		t.Fatalf("expected 'encoder is closed' error, got: %v", err)
	}
}

func TestFullEncodeCycle(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	if err := enc.SetInSamplerate(44100); err != nil {
		t.Fatalf("SetInSamplerate: %v", err)
	}
	if err := enc.SetNumChannels(2); err != nil {
		t.Fatalf("SetNumChannels: %v", err)
	}
	if err := enc.SetBrate(192); err != nil {
		t.Fatalf("SetBrate: %v", err)
	}
	if err := enc.SetVBR(VBROff); err != nil {
		t.Fatalf("SetVBR: %v", err)
	}

	pcm := generateSineWave(44100, 2, 1000)
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	lametagFrame := enc.GetLametagFrame()
	if lametagFrame == nil {
		t.Log("GetLametagFrame returned nil (acceptable for CBR)")
	} else if len(lametagFrame) == 0 {
		t.Fatal("GetLametagFrame returned empty slice")
	}

	if err := enc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	mp3Size := buf.Len()
	const minSize = 10 * 1024 // 10 KB
	const maxSize = 50 * 1024 // 50 KB
	if mp3Size < minSize || mp3Size > maxSize {
		t.Fatalf("MP3 output size %d bytes out of expected range [%d, %d]", mp3Size, minSize, maxSize)
	}
}

func TestFullEncodeCycle_VBR(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
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
	if _, err := enc.Write(pcm); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if _, err := enc.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	lametagFrame := enc.GetLametagFrame()
	if lametagFrame == nil || len(lametagFrame) == 0 {
		t.Fatal("GetLametagFrame returned nil or empty for VBR — expected valid VBR header")
	}

	if err := enc.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if buf.Len() == 0 {
		t.Fatal("VBR encode produced no output")
	}
}
