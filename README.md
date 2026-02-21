# wav2mp3

High-quality WAV → MP3 converter. Uses [libmp3lame](https://lame.sourceforge.io/) via CGo with VBR support, ID3v2 tags, and album cover embedding.

## Requirements

- macOS (Apple Silicon or Intel)
- Go 1.21+
- libmp3lame:

```bash
brew install lame
```

## Installation

```bash
git clone ...
cd svk-wav2mp3
make install        # builds and copies to /usr/local/bin/wav2mp3
```

Or just build:

```bash
make build          # ./bin/wav2mp3
```

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

## Development

```bash
make test           # all tests
make testdata       # regenerate WAV fixtures
make fmt            # go fmt
```

## License

MIT