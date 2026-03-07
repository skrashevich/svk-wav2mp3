# TODO: svk-wav2mp3

## Completed

- [x] go mod init + go get dependencies
- [x] internal/config/options.go — ConvertOptions, ID3Tags
- [x] internal/converter/wav_reader.go — WAVReader + bit-depth normalization
- [x] internal/converter/mp3_writer.go — MP3Writer + lame encoder config
- [x] internal/converter/pipeline.go — streaming PCM pipeline + progressbar
- [x] internal/tagger/cover.go — LoadCover with magic bytes detection
- [x] internal/tagger/tagger.go — ID3v2Tagger
- [x] internal/converter/converter.go — DefaultConverter (orchestration)
- [x] internal/validate/validate.go — validation of options
- [x] internal/cli/root.go — cobra command with flags
- [x] cmd/wav2mp3/main.go — main with signal handling
- [x] Makefile — build/test/install/testdata

## Completed (continued)

- [x] Tests: unit (validate — 8 tests, cover detection — 6 tests, bit-depth — 5 tests)
- [x] Tests: integration (5 tests: stereo 16-bit, mono 8-bit, CBR, auto output path, context cancel)
- [x] Verification: make build && ffprobe confirmed correct tags in MP3

## Next Steps (optional)

## Future Improvements

- [ ] Batch conversion (glob patterns, recursive directory traversal)
- [ ] Goroutine pool for parallel conversion
- [ ] ReplayGain support
- [ ] Version output via `wav2mp3 --version`