package validate

import (
	"os"
	"testing"

	"github.com/svk/wav2mp3/internal/config"
)

func TestConvertOptions_MissingInput(t *testing.T) {
	opts := config.DefaultConvertOptions()
	if err := ConvertOptions(&opts); err == nil {
		t.Error("ожидалась ошибка при отсутствии входного файла")
	}
}

func TestConvertOptions_NonExistentInput(t *testing.T) {
	opts := config.DefaultConvertOptions()
	opts.InputPath = "/nonexistent/path/file.wav"
	if err := ConvertOptions(&opts); err == nil {
		t.Error("ожидалась ошибка для несуществующего файла")
	}
}

func TestConvertOptions_WrongExtension(t *testing.T) {
	// Создаём временный файл с неверным расширением
	f, err := os.CreateTemp("", "test*.mp3")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.Close()

	opts := config.DefaultConvertOptions()
	opts.InputPath = f.Name()
	if err := ConvertOptions(&opts); err == nil {
		t.Error("ожидалась ошибка для файла с расширением .mp3")
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
		t.Errorf("не ожидалась ошибка для валидных VBR параметров: %v", err)
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
	opts.VBRQuality = 10.5 // вне диапазона
	if err := ConvertOptions(&opts); err == nil {
		t.Error("ожидалась ошибка для VBR quality > 9.9")
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
	opts.Bitrate = 400 // вне диапазона
	if err := ConvertOptions(&opts); err == nil {
		t.Error("ожидалась ошибка для bitrate > 320")
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
	opts.Quality = 10 // вне диапазона
	if err := ConvertOptions(&opts); err == nil {
		t.Error("ожидалась ошибка для quality > 9")
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
		t.Error("ожидалась ошибка для несуществующей обложки")
	}
}
