package lame

import (
	"io"
	"testing"
)

func TestNewEncoder(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()
}

func TestSetInSamplerate(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	for _, rate := range []int{22050, 44100, 48000} {
		if err := enc.SetInSamplerate(rate); err != nil {
			t.Errorf("SetInSamplerate(%d) error: %v", rate, err)
		}
	}
}

func TestSetNumChannels(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	for _, n := range []int{1, 2} {
		if err := enc.SetNumChannels(n); err != nil {
			t.Errorf("SetNumChannels(%d) error: %v", n, err)
		}
	}
}

func TestSetQuality(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	for q := 0; q <= 9; q++ {
		if err := enc.SetQuality(q); err != nil {
			t.Errorf("SetQuality(%d) error: %v", q, err)
		}
	}
}

func TestSetVBR(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	for _, mode := range []int{VBROff, VBRDefault} {
		if err := enc.SetVBR(mode); err != nil {
			t.Errorf("SetVBR(%d) error: %v", mode, err)
		}
	}
}

func TestSetVBRQuality(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	for _, q := range []float64{0.0, 5.0, 9.0} {
		if err := enc.SetVBRQuality(q); err != nil {
			t.Errorf("SetVBRQuality(%v) error: %v", q, err)
		}
	}
}

func TestSetBrate(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	for _, kbps := range []int{128, 192, 256, 320} {
		if err := enc.SetBrate(kbps); err != nil {
			t.Errorf("SetBrate(%d) error: %v", kbps, err)
		}
	}
}

func TestSetWriteID3TagAutomatic(t *testing.T) {
	enc := NewEncoder(io.Discard)
	if enc == nil {
		t.Fatal("NewEncoder returned nil")
	}
	defer enc.Close()

	enc.SetWriteID3TagAutomatic(true)
	enc.SetWriteID3TagAutomatic(false)
}
