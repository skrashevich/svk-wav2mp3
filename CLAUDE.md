# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Pure-Go Build

No system dependencies required. LAME 3.100 is transpiled from C to Go using `modernc.org/ccgo` and lives in `internal/lame/`. The build uses `CGO_ENABLED=0` (no C compiler needed). The transpiled code uses `modernc.org/libc` as its runtime.

## Commands

```bash
make build       # → ./bin/wav2mp3 (CGO_ENABLED=0, pure Go)
make test        # CGO_ENABLED=0 go test ./... -v -count=1
make install     # copies to /usr/local/bin/wav2mp3
make testdata    # regenerates testdata/fixtures/*.wav via go run testdata/gen_fixtures.go
make fmt         # go fmt ./...
make lint        # golangci-lint run ./...

# Single package test
CGO_ENABLED=0 go test ./internal/converter/... -run TestConvert_Stereo24bit -v

# Run binary directly
./bin/wav2mp3 -i input.wav --bitrate 320        # CBR
./bin/wav2mp3 -i input.wav --vbr-quality 0      # VBR V0 (best)
./bin/wav2mp3 -i input.wav --artist "X" --cover cover.jpg
```

## Architecture

### Data flow
```
WAV file → WAVReader → normalizePCM → MP3Writer (LAME) → MP3 file → id3v2.Open → tags
```

### Package responsibilities

- **`internal/lame`** — pure-Go LAME 3.100 encoder, transpiled from C via ccgo. Provides `Encoder` with `Write`/`Flush`/`Close` and VBR/CBR configuration.
- **`internal/config`** — `ConvertOptions` and `ID3Tags` structs. No logic.
- **`internal/validate`** — validates `ConvertOptions` before conversion starts (file existence, extension, bitrate/quality ranges).
- **`internal/converter`** — the conversion core:
  - `WAVReader` reads PCM in 4096-frame chunks via `decoder.PCMBuffer()` (streaming, not `FullPCMBuffer`).
  - `normalizePCM` converts 8/16/24/32-bit `[]int` → `[]int16` (LAME input format).
  - `MP3Writer` wraps LAME encoder; `WriteSamples` converts `[]int16` → `[]byte` (LE) before passing to `encoder.Write`.
  - `converter.Convert` orchestrates: open WAV → create MP3Writer → `RunPipeline` → close → write tags.
- **`internal/tagger`** — `LoadCover` detects MIME type via magic bytes (JPEG/PNG/GIF/WebP); `Apply` writes ID3v2.4 tags via bogem/id3v2 **after** LAME encoder is closed.
- **`internal/cli`** — cobra root command. VBR/CBR mode is derived from `cmd.Flags().Changed("bitrate")`: if `--bitrate` was explicitly provided → CBR; otherwise → VBR. The `--vbr` boolean flag does not exist.

### Critical ordering constraint
LAME ID3 tag writing is disabled (`SetWriteID3TagAutomatic(false)`). Tags are written **after** `MP3Writer.Close()` using `id3v2.Open()`. Reversing this order corrupts the file.

### Cleanup on error
`converter.Convert` uses a `defer` that calls `os.Remove(outputPath)` unless `success = true` is reached. This prevents partial MP3 files on error or context cancellation.

## internal/lame package (pure-Go LAME)
- Transpiled from LAME 3.100 C source using `modernc.org/ccgo` v4
- Platform-specific generated files: `lame_darwin_arm64.go`, `lame_linux_arm64.go`
- `encoder.go` provides the Go wrapper API: `NewEncoder`, `SetBrate`, `SetVBR`, `Write`, `Flush`, `Close`
- `SetBrate(kbps int)` — CBR bitrate (not `SetBitrate`)
- `initParams` is called automatically on first `Write`
- `encoder.Write([]byte)` expects int16 LE interleaved bytes
- `encoder.Flush()` returns `(int, error)`
- To regenerate for new platforms, use ccgo inside Docker (see `build/Dockerfile.ccgo`)

## Testing
- Unit tests: `normalizePCM` (wav_reader_test.go), magic bytes (cover_test.go), validation (validate_test.go)
- Integration tests in `integration_test.go` generate WAV in-process; fixture files in `testdata/fixtures/` are pre-generated WAVs committed to the repo
- **In zsh, avoid `if ! command` for exit code checks** — use `command; rc=$?; test $rc -ne 0` instead (zsh `!` history expansion interference after many commands)
