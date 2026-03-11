# wav2mp3

High-quality WAV → MP3 converter. Pure Go, no CGO, no system dependencies. LAME 3.100 is [transpiled from C to Go](lame/) and linked statically.

Supports VBR/CBR encoding, ID3v2 tags, and album cover embedding.

## Installation

```bash
go install github.com/svk/wav2mp3/cmd/wav2mp3@latest
```

Or build from source:

```bash
make build          # ./bin/wav2mp3
make install        # copies to /usr/local/bin/wav2mp3
```

### Platforms

linux/amd64, linux/arm64, darwin/amd64, darwin/arm64

## Usage

```
wav2mp3 -i INPUT [flags]
```

### Examples

```bash
# VBR V2 by default — best quality/size balance
wav2mp3 -i song.wav

# Specify output file
wav2mp3 -i song.wav -o song_hq.mp3

# With tags and cover
wav2mp3 -i song.wav \
  --title "Song Title" \
  --artist "Artist Name" \
  --album "Album" \
  --year "2026" \
  --genre "Electronic" \
  --track "3" \
  --cover cover.jpg

# CBR 320 kbps
wav2mp3 -i song.wav --bitrate 320

# VBR maximum quality (V0)
wav2mp3 -i song.wav --vbr-quality 0

# Quiet mode (no progress bar and stats)
wav2mp3 -i song.wav -q

# Verbose output (encoder parameters)
wav2mp3 -i song.wav -v
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-i, --input` | — | Input WAV file **(required)** |
| `-o, --output` | next to input | Output MP3 file |
| `--title` | | Track title |
| `--artist` | | Artist |
| `--album` | | Album |
| `--year` | | Year |
| `--genre` | | Genre |
| `--track` | | Track number |
| `--comment` | | Comment |
| `--cover` | | Cover (JPEG, PNG, GIF, WebP) |
| `--vbr-quality` | `2.0` | VBR quality 0.0 (better) – 9.9; incompatible with `--bitrate` |
| `--bitrate` | — | CBR bitrate kbps (32–320); specifying disables VBR |
| `--quality` | `2` | LAME algorithmic quality 0 (better) – 9 |
| `-v, --verbose` | | Verbose output |
| `-q, --quiet` | | No progress bar and stats |
| `--version` | | Version |

### Supported WAV Formats

| Format | Channels |
|--------|----------|
| 8-bit PCM | Mono, Stereo |
| 16-bit PCM | Mono, Stereo |
| 24-bit PCM | Mono, Stereo |
| 32-bit PCM | Mono, Stereo |

## Output

```
Input:  song.wav (44100 Hz, Stereo, 16-bit, 3m 42s, 39.1 MB)
Output: song.mp3 (VBR V2, elapsed 12.3s, 8.4 MB, compression 4.65x)
Tags:   Title="Song Title", Artist="Artist Name", Cover=cover.jpg
```

## Using the LAME encoder as a library

The [`lame`](lame/) package is a standalone pure-Go MP3 encoder that can be used in any Go project:

```bash
go get github.com/svk/wav2mp3/lame
```

```go
import "github.com/svk/wav2mp3/lame"

f, _ := os.Create("output.mp3")
enc := lame.NewEncoder(f)
defer enc.Close()

enc.SetInSamplerate(44100)
enc.SetNumChannels(2)
enc.SetVBR(lame.VBRDefault)
enc.SetVBRQuality(2)

enc.Write(pcmBytes) // int16 little-endian interleaved
enc.Flush()
```

No CGO or system libraries required. See [`lame/README.md`](lame/README.md) for the full API reference.

## Development

```bash
make test           # all tests
make testdata       # regenerate WAV fixtures
make fmt            # go fmt
```

## License

MIT