package tagger

import (
	"os"
	"testing"
)

func TestLoadCover_JPEG(t *testing.T) {
	// JPEG magic bytes: FF D8 FF E0
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}
	f, err := os.CreateTemp("", "cover*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	cover, err := LoadCover(f.Name())
	if err != nil {
		t.Fatalf("не ожидалась ошибка: %v", err)
	}
	if cover.MIMEType != "image/jpeg" {
		t.Errorf("ожидался image/jpeg, получен %s", cover.MIMEType)
	}
}

func TestLoadCover_PNG(t *testing.T) {
	// PNG magic bytes
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00}
	f, err := os.CreateTemp("", "cover*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	cover, err := LoadCover(f.Name())
	if err != nil {
		t.Fatalf("не ожидалась ошибка: %v", err)
	}
	if cover.MIMEType != "image/png" {
		t.Errorf("ожидался image/png, получен %s", cover.MIMEType)
	}
}

func TestLoadCover_GIF(t *testing.T) {
	// GIF magic bytes: GIF89a
	data := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00}
	f, err := os.CreateTemp("", "cover*.gif")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	cover, err := LoadCover(f.Name())
	if err != nil {
		t.Fatalf("не ожидалась ошибка: %v", err)
	}
	if cover.MIMEType != "image/gif" {
		t.Errorf("ожидался image/gif, получен %s", cover.MIMEType)
	}
}

func TestLoadCover_WebP(t *testing.T) {
	// WebP: RIFF????WEBP
	data := []byte{
		0x52, 0x49, 0x46, 0x46, // RIFF
		0x24, 0x00, 0x00, 0x00, // размер
		0x57, 0x45, 0x42, 0x50, // WEBP
		0x56, 0x50, 0x38, 0x20, // VP8
	}
	f, err := os.CreateTemp("", "cover*.webp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	cover, err := LoadCover(f.Name())
	if err != nil {
		t.Fatalf("не ожидалась ошибка: %v", err)
	}
	if cover.MIMEType != "image/webp" {
		t.Errorf("ожидался image/webp, получен %s", cover.MIMEType)
	}
}

func TestLoadCover_UnknownFormat(t *testing.T) {
	data := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}
	f, err := os.CreateTemp("", "cover*.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Write(data)
	f.Close()

	if _, err := LoadCover(f.Name()); err == nil {
		t.Error("ожидалась ошибка для неизвестного формата")
	}
}

func TestLoadCover_NonExistent(t *testing.T) {
	if _, err := LoadCover("/nonexistent/cover.jpg"); err == nil {
		t.Error("ожидалась ошибка для несуществующего файла")
	}
}
