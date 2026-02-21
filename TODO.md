# TODO: svk-wav2mp3

## Выполнено

- [x] go mod init + go get зависимостей
- [x] internal/config/options.go — ConvertOptions, ID3Tags
- [x] internal/converter/wav_reader.go — WAVReader + bit-depth нормализация
- [x] internal/converter/mp3_writer.go — MP3Writer + lame encoder config
- [x] internal/converter/pipeline.go — потоковый PCM pipeline + progressbar
- [x] internal/tagger/cover.go — LoadCover с magic bytes detection
- [x] internal/tagger/tagger.go — ID3v2Tagger
- [x] internal/converter/converter.go — DefaultConverter (оркестрация)
- [x] internal/validate/validate.go — валидация опций
- [x] internal/cli/root.go — cobra команда с флагами
- [x] cmd/wav2mp3/main.go — main с signal handling
- [x] Makefile — build/test/install/testdata

## Выполнено (продолжение)

- [x] Тесты: unit (validate — 8 тестов, cover detection — 6 тестов, bit-depth — 5 тестов)
- [x] Тесты: integration (5 тестов: stereo 16-bit, mono 8-bit, CBR, auto output path, context cancel)
- [x] Верификация: make build && ffprobe подтвердил корректные теги в MP3

## Следующие шаги (опционально)

## Будущие улучшения

- [ ] Пакетная конвертация (glob-паттерны, рекурсивный обход директорий)
- [ ] Пул горутин для параллельной конвертации
- [ ] Поддержка ReplayGain
- [ ] Вывод версии go.mod через `wav2mp3 --version`
