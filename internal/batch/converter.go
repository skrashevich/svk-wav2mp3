package batch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/schollz/progressbar/v3"
	"github.com/svk/wav2mp3/internal/config"
	"github.com/svk/wav2mp3/internal/converter"
)

// BatchOptions содержит параметры batch конвертации.
type BatchOptions struct {
	Patterns   []string
	Recursive  bool
	OutputDir  string
	Tags       config.ID3Tags
	Quiet      bool
	ReplaceExt bool // заменить .wav на .mp3
}

// ConvertBatch конвертирует несколько WAV файлов с glob-паттернами.
func ConvertBatch(ctx context.Context, opts BatchOptions) (*BatchStats, error) {
	stats := &BatchStats{}
	startTime := time.Now()

	if len(opts.Patterns) == 0 {
		return nil, fmt.Errorf("no patterns provided")
	}

	// Собираем все WAV файлы по glob-паттернам
	matches, err := gatherFiles(ctx, opts)
	if err != nil {
		return nil, err
	}
	stats.TotalFiles = len(matches)

	if stats.TotalFiles == 0 {
		if !opts.Quiet {
			fmt.Fprintln(os.Stderr, "No matching WAV files found.")
		}
		return stats, nil
	}

	// Обновляем outputDir если указан
	if opts.OutputDir != "" {
		opts.OutputDir = filepath.Clean(opts.OutputDir)
	}

	var bar *progressbar.ProgressBar
	if !opts.Quiet && stats.TotalFiles > 0 {
		bar = progressbar.NewOptions(stats.TotalFiles,
			progressbar.OptionSetDescription("Batch conversion"),
			progressbar.OptionSetWriter(os.Stderr),
			progressbar.OptionShowBytes(false),
			progressbar.OptionSetWidth(60),
			progressbar.OptionThrottle(0),
			progressbar.OptionShowCount(),
			progressbar.OptionClearOnFinish(),
			progressbar.OptionSetRenderBlankState(true),
		)
	}

	// Обрабатываем каждый файл
	successes := 0
	failed := 0
	var outputPaths []string

	for _, inputPath := range matches {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("batch operation cancelled: %w", err)
		}

		// Определяем выходной путь
		outputPath := determineOutputPath(inputPath, opts)

		// Создаем опции конвертации
		convertOpts := config.ConvertOptions{
			InputPath:  inputPath,
			OutputPath: outputPath,
			Tags:       opts.Tags,
			Quiet:      true, // Внутренний конвертер использует тихий режим
		}

		// Конвертируем
		_, err := converter.Convert(ctx, convertOpts)
		if err != nil {
			failed++
			stats.ErrorMessages = append(stats.ErrorMessages,
				fmt.Sprintf("%s -> %s: %v", inputPath, outputPath, err))
			if !opts.Quiet {
				fmt.Fprintf(os.Stderr, "  [FAIL] %s\n", filepath.Base(inputPath))
			}
			continue
		}

		successes++
		outputPaths = append(outputPaths, outputPath)

		if bar != nil {
			_ = bar.Add(1)
		}
	}

	if bar != nil {
		_ = bar.Finish()
	}

	stats.Successful = successes
	stats.Failed = failed
	stats.OutputFiles = outputPaths
	stats.ElapsedTime = int64(time.Since(startTime).Seconds())

	// Выводим статистику
	if !opts.Quiet {
		printStats(stats)
	}

	return stats, nil
}

// gatherFiles собирает WAV файлы по glob-паттернам (с учетом рекурсивности).
func gatherFiles(ctx context.Context, opts BatchOptions) ([]string, error) {
	var allFiles []string

	for _, pattern := range opts.Patterns {
		// Добавляем текущую директорию как базовую
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(".", pattern)
		}

		matches, err := doublestar.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("glob error for pattern %q: %w", pattern, err)
		}

		if opts.Recursive {
			// Рекурсивный поиск
			matches = findFilesRecursive(pattern, matches)
		}

		allFiles = append(allFiles, matches...)
	}

	// Удаляем дубликаты
	uniqueFiles := make(map[string]struct{})
	for _, file := range allFiles {
		if _, exists := uniqueFiles[file]; exists {
			continue
		}
		uniqueFiles[file] = struct{}{}
	}

	files := make([]string, 0, len(uniqueFiles))
	for file := range uniqueFiles {
		files = append(files, file)
	}

	return files, nil
}

// findFilesRecursive находит файлы рекурсивно.
func findFilesRecursive(basePattern string, baseMatches []string) []string {
	var results []string

	// Получаем директорию из паттерна
	baseDir := filepath.Dir(basePattern)
	if baseDir == "." {
		baseDir = "."
	}

	// Рекурсивно сканируем
	walkFn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && (strings.HasSuffix(strings.ToLower(path), ".wav") ||
			strings.HasSuffix(strings.ToLower(path), ".wave")) {
			// Преобразуем базовый паттерн в glob формат
			globPattern := patternToGlob(basePattern)
			// Проверяем соответствие
			if _, err := doublestar.Match(globPattern, path); err == nil {
				results = append(results, path)
			}
		}
		return nil
	}

	filepath.Walk(baseDir, walkFn)

	return results
}

// determineOutputPath определяет выходной путь для конвертации.
func determineOutputPath(inputPath string, opts BatchOptions) string {
	if opts.OutputDir != "" {
		// Сохраняем в специфицированную директорию
		filename := filepath.Base(inputPath)

		// Заменяем .wave на .mp3 если ReplaceExt=true
		if strings.HasSuffix(strings.ToLower(filename), ".wave") && opts.ReplaceExt {
			filename = filename[:len(filename)-5] + ".mp3"
			return filepath.Join(opts.OutputDir, filename)
		}

		ext := filepath.Ext(filename)
		if opts.ReplaceExt {
			ext = ".mp3"
			// Добавляем .mp3 если нет расширения
			if ext == "" {
				return filepath.Join(opts.OutputDir, filename+".mp3")
			}
		}

		return filepath.Join(opts.OutputDir, filename[:len(filename)-len(ext)]+ext)
	}

	// Используем путь рядом с входным
	if strings.HasSuffix(strings.ToLower(inputPath), ".wave") {
		return inputPath[:len(inputPath)-5] + ".mp3"
	}
	return inputPath[:len(inputPath)-4] + ".mp3"
}

// printStats выводит статистику batch конвертации.
func printStats(stats *BatchStats) {
	total := stats.Successful + stats.Failed + stats.Skipped
	if total == 0 {
		return
	}

	percentComplete := 0.0
	if stats.Successful > 0 {
		percentComplete = float64(stats.Successful) / float64(total) * 100
	}

	fmt.Fprintf(os.Stderr, "\n=== Batch Conversion Complete ===\n")
	fmt.Fprintf(os.Stderr, "Total files:    %d\n", stats.TotalFiles)
	fmt.Fprintf(os.Stderr, "Successful:     %d (%.1f%%)\n", stats.Successful, percentComplete)
	if stats.Failed > 0 {
		fmt.Fprintf(os.Stderr, "Failed:         %d\n", stats.Failed)
	}
	if stats.Skipped > 0 {
		fmt.Fprintf(os.Stderr, "Skipped:        %d\n", stats.Skipped)
	}
	fmt.Fprintf(os.Stderr, "Output files:   %d\n", stats.Successful)
	if stats.ElapsedTime > 0 {
		fmt.Fprintf(os.Stderr, "Elapsed time:   %.1fs\n", float64(stats.ElapsedTime))
	}

	if len(stats.ErrorMessages) > 0 {
		fmt.Fprintf(os.Stderr, "\nFailed files:\n")
		for _, msg := range stats.ErrorMessages {
			fmt.Fprintf(os.Stderr, "  - %s\n", msg)
		}
	}
	fmt.Fprintf(os.Stderr, "==============================\n")
}

// patternToGlob конвертирует паттерн filepath в glob-паттерн для рекурсивного поиска.
// *.wav -> **/*.wav, test/*.wav -> test/**/*.wav, test -> test/*.wav
func patternToGlob(pattern string) string {
	if !strings.Contains(pattern, "*") {
		// No wildcard — add /*.wav
		return filepath.Join(pattern, "*.wav")
	}
	// Replace single * with **/* for recursive doublestar matching
	dir := filepath.Dir(pattern)
	base := filepath.Base(pattern)
	if dir == "." {
		return "**/" + base
	}
	return dir + "/**/" + base
}

// globWAVFiles собирает WAV файлы по паттерну.
func globWAVFiles(ctx context.Context, pattern string) ([]string, error) {
	if !filepath.IsAbs(pattern) {
		pattern = filepath.Join(".", pattern)
	}

	matches, err := doublestar.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, m := range matches {
		m = filepath.Clean(m)
		if strings.HasSuffix(strings.ToLower(m), ".wav") ||
			strings.HasSuffix(strings.ToLower(m), ".wave") {
			result = append(result, m)
		}
	}

	// Если не найдено, пробуем с явным .wav расширением
	if len(result) == 0 {
		pattern = strings.TrimSuffix(pattern, "*")
		if !filepath.IsAbs(pattern) {
			pattern = filepath.Join(".", pattern)
		}
		pattern += "*.wav"
		matches, err = doublestar.Glob(pattern)
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			result = append(result, m)
		}
	}

	// Сортируем для детерминированного вывода
	sort.Strings(result)

	return result, nil
}
