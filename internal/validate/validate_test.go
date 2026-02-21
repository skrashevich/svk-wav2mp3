package validate

import (
	"os"
	"testing"

	"github.com/svk/wav2mp3/internal/config"
)

func TestConvertOptions_MissingInput(t *testing.T) {
	opts := config.DefaultConvertOptions()
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error when input file is missing")
	}
}

func TestConvertOptions_NonExistentInput(t *testing.T) {
	opts := config.DefaultConvertOptions()
	opts.InputPath = "/nonexistent/path/file.wav"
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestConvertOptions_WrongExtension(t *testing.T) {
	// Create temporary file with wrong extension
	f, err := os.CreateTemp("", "test*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error for file with .mp3 extension")
	}
}

func TestConvertOptions_ValidVBR(t *testing.T) {
	f, err := os.CreateTemp("", "test*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	opts.VBR = true
	opts.VBRQuality = 2.0
	opts.Quality = 2
	if err := ConvertOptions(&opts); err != nil {
		t.Errorf("unexpected error for valid VBR parameters: %v", err)
	}
}

func TestConvertOptions_InvalidVBRQuality(t *testing.T) {
	f, err := os.CreateTemp("", "test*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	opts.VBR = true
	opts.VBRQuality = 10.5 // out of range
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error for VBR quality > 9.9")
	}
}

func TestConvertOptions_InvalidBitrate(t *testing.T) {
	f, err := os.CreateTemp("", "test*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	opts.VBR = false
	opts.Bitrate = 400 // out of range
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error for bitrate > 320")
	}
}

func TestConvertOptions_InvalidQuality(t *testing.T) {
	f, err := os.CreateTemp("", "test*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	opts.Quality = 10 // out of range
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error for quality > 9")
	}
}

func TestConvertOptions_NonExistentCover(t *testing.T) {
	f, err := os.CreateTemp("", "test*.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	opts.Tags.Cover = "/nonexistent/cover.jpg"
	if err := ConvertOptions(&opts); err == nil {
		t.Error("expected error for non-existent cover")
	}
}
