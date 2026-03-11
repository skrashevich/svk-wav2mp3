package batch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGatherFiles(t *testing.T) {
	// Создаем тестовые файлы
	tmpDir := t.TempDir()

	files := []string{
		filepath.Join(tmpDir, "test1.wav"),
		filepath.Join(tmpDir, "test2.wav"),
		filepath.Join(tmpDir, "test3.wav"),
	}

	for _, f := range files {
		if err := os.WriteFile(f, []byte{0x00}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Тестируем glob паттерн
	opts := BatchOptions{
		Patterns:  []string{filepath.Join(tmpDir, "*.wav")},
		Recursive: false,
	}

	files, err := gatherFiles(context.Background(), opts)
	if err != nil {
		t.Fatalf("gatherFiles error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("expected 3 files, got %d", len(files))
	}

	// Тестируем несуществующий паттерн
	opts.Patterns = []string{filepath.Join(tmpDir, "nonexistent*.wav")}
	files, err = gatherFiles(context.Background(), opts)
	if err != nil {
		t.Fatalf("gatherFiles with non-existent pattern: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("expected 0 files, got %d", len(files))
	}
}

func TestDetermineOutputPath(t *testing.T) {
	tests := []struct {
		name           string
		inputPath      string
		outputDir      string
		replaceExt     bool
		expectedResult string
	}{
		{
			name:           "Output dir specified",
			inputPath:      "/tmp/input.wav",
			outputDir:      "/tmp/output",
			replaceExt:     false,
			expectedResult: "/tmp/output/input.wav",
		},
		{
			name:           "Output dir with replace extension",
			inputPath:      "/tmp/input.wav",
			outputDir:      "/tmp/output",
			replaceExt:     true,
			expectedResult: "/tmp/output/input.mp3",
		},
		{
			name:           "No output dir, default extension",
			inputPath:      "/tmp/input.wav",
			outputDir:      "",
			replaceExt:     false,
			expectedResult: "/tmp/input.mp3",
		},
		{
			name:           "Input .wave file",
			inputPath:      "/tmp/input.wave",
			outputDir:      "",
			replaceExt:     false,
			expectedResult: "/tmp/input.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineOutputPath(tt.inputPath, BatchOptions{
				OutputDir:  tt.outputDir,
				ReplaceExt: tt.replaceExt,
			})

			if result != tt.expectedResult {
				t.Errorf("expected %s, got %s", tt.expectedResult, result)
			}
		})
	}
}

func TestPatternToGlob(t *testing.T) {
	tests := []struct {
		input       string
		output      string
		description string
	}{
		{
			input:       "*.wav",
			output:      "**/*.wav",
			description: "*.wav should become **/*.wav",
		},
		{
			input:       "test/*.wav",
			output:      "test/**/*.wav",
			description: "test/*.wav should become test/**/*.wav",
		},
		{
			input:       "/absolute/path/*.wav",
			output:      "/absolute/path/**/*.wav",
			description: "Absolute path should work correctly",
		},
		{
			input:       "test",
			output:      "test/*.wav",
			description: "Pattern without * should add *.wav",
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			result := patternToGlob(tt.input)
			if result != tt.output {
				t.Errorf("expected %s, got %s", tt.output, result)
			}
		})
	}
}

func TestGlobWAVFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Создаем тестовые файлы
	testFiles := []struct {
		name     string
		contains string
	}{
		{"file1.wav", ".wav"},
		{"file2.wav", ".wav"},
		{"file1.wave", ".wave"},
		{"file3.txt", ".txt"}, // Этот файл должен быть пропущен
	}

	for _, tf := range testFiles {
		fpath := filepath.Join(tmpDir, tf.name)
		if err := os.WriteFile(fpath, []byte{0x00}, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Тестируем поиск
	pattern := filepath.Join(tmpDir, "*.wav")
	files, err := globWAVFiles(context.Background(), pattern)
	if err != nil {
		t.Fatalf("globWAVFiles error: %v", err)
	}

	expectedCount := 2 // только .wav файлы
	if len(files) != expectedCount {
		t.Errorf("expected %d files, got %d", expectedCount, len(files))
	}

	// Проверяем, что результаты сортированы
	for i := 1; i < len(files); i++ {
		if files[i-1] > files[i] {
			t.Error("results should be sorted")
		}
	}
}

func TestConvertBatchWithNoFiles(t *testing.T) {
	opts := BatchOptions{
		Patterns: []string{"/nonexistent/*.wav"},
		Quiet:    true,
	}

	stats, err := ConvertBatch(context.Background(), opts)
	if err != nil {
		t.Fatalf("ConvertBatch should not return error: %v", err)
	}

	if stats.TotalFiles != 0 {
		t.Errorf("expected 0 total files, got %d", stats.TotalFiles)
	}
}

func TestConvertBatchWithEmptyPatterns(t *testing.T) {
	opts := BatchOptions{
		Patterns: []string{},
		Quiet:    true,
	}

	_, err := ConvertBatch(context.Background(), opts)
	if err == nil {
		t.Error("expected error for empty patterns")
	}
}

func TestConvertBatchWithSingleFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "src.wav")

	// Создаем невалидный WAV (1 байт) — конвертация должна завершиться ошибкой
	if err := os.WriteFile(srcFile, []byte{0x00}, 0644); err != nil {
		t.Fatal(err)
	}

	opts := BatchOptions{
		Patterns:   []string{srcFile},
		Quiet:      true,
		ReplaceExt: true,
	}

	stats, err := ConvertBatch(context.Background(), opts)
	if err != nil {
		t.Fatalf("ConvertBatch error: %v", err)
	}

	if stats.TotalFiles != 1 {
		t.Errorf("expected 1 total file, got %d", stats.TotalFiles)
	}

	if stats.Failed != 1 {
		t.Errorf("expected 1 failed conversion for invalid WAV, got %d", stats.Failed)
	}
}
