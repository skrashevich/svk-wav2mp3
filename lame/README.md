# lame — pure-Go MP3 encoder

Pure-Go LAME 3.100 MP3 encoder. No CGO, no system libraries — works with `CGO_ENABLED=0`.

The LAME C source is transpiled to Go using [modernc.org/ccgo](https://pkg.go.dev/modernc.org/ccgo/v4) and uses [modernc.org/libc](https://pkg.go.dev/modernc.org/libc) as its runtime.

## Install

```bash
go get github.com/svk/wav2mp3/lame
```

## Platforms

- `darwin/arm64`
- `darwin/amd64`
- `linux/arm64`
- `linux/amd64`

amd64 files are generated in CI via [ccgo](build/Dockerfile.ccgo).

## Usage

### VBR encoding (recommended)

```go
package main

import (
	"os"
	"github.com/svk/wav2mp3/lame"
)

func main() {
	f, _ := os.Create("output.mp3")
	defer f.Close()

	enc := lame.NewEncoder(f)
	defer enc.Close()

	enc.SetInSamplerate(44100)
	enc.SetNumChannels(2)
	enc.SetQuality(2)              // algorithm quality: 0=best, 9=fast
	enc.SetVBR(lame.VBRDefault)
	enc.SetVBRQuality(2)           // VBR quality: 0=best, 9.9=worst

	// Feed PCM data (int16 little-endian, interleaved L R L R …)
	enc.Write(pcmBytes)

	// Finalize
	enc.Flush()

	// Write VBR header for accurate duration reporting
	if tag := enc.GetLametagFrame(); len(tag) > 0 {
		f.Seek(0, 0)
		f.Write(tag)
	}
}
```

### CBR encoding

```go
enc := lame.NewEncoder(f)
defer enc.Close()

enc.SetInSamplerate(44100)
enc.SetNumChannels(2)
enc.SetQuality(2)
enc.SetVBR(lame.VBROff)
enc.SetBrate(320)    // bitrate in kbps

enc.Write(pcmBytes)
enc.Flush()
```

### Presets

Presets configure VBR quality, filters, and psychoacoustic tuning in a single call:

```go
enc.SetPreset(lame.PresetStandard) // ~170–210 kbps VBR
```

| Constant | Bitrate | Description |
|---|---|---|
| `PresetMedium` | ~150–180 kbps | Good quality |
| `PresetStandard` | ~170–210 kbps | Transparent for most listeners |
| `PresetExtreme` | ~210–260 kbps | Near-transparent |
| `PresetInsane` | 320 kbps CBR | Maximum quality |

### Channel modes

```go
enc.SetMode(lame.ModeJointStereo) // default — best compression
enc.SetMode(lame.ModeStereo)      // independent L/R
enc.SetMode(lame.ModeMono)        // mono
```

## API reference

### Constructor

| Function | Description |
|---|---|
| `NewEncoder(w io.Writer) *Encoder` | Creates encoder writing MP3 to `w`. Returns `nil` on failure. |

### Configuration (call before first Write)

| Method | Description |
|---|---|
| `SetInSamplerate(hz int)` | Input sample rate (e.g. 44100, 48000) |
| `SetOutSamplerate(hz int)` | Output sample rate (0 = auto) |
| `SetNumChannels(n int)` | 1 = mono, 2 = stereo |
| `SetQuality(q int)` | Algorithm quality: 0 = best/slow, 9 = fast |
| `SetVBR(mode int)` | `VBRDefault` or `VBROff` |
| `SetVBRQuality(q float64)` | VBR quality: 0.0 = best, 9.9 = worst |
| `SetBrate(kbps int)` | CBR bitrate (128, 192, 256, 320…) |
| `SetMode(mode int)` | Channel mode (see constants) |
| `SetPreset(preset int)` | LAME preset (see constants) |
| `SetScale(factor float64)` | PCM scaling factor |
| `SetLowpassFreq(hz int)` | Lowpass cutoff (0 = auto, -1 = disabled) |
| `SetHighpassFreq(hz int)` | Highpass cutoff (0 = auto, -1 = disabled) |
| `SetWriteID3TagAutomatic(v bool)` | Control LAME's built-in ID3 writing |

### Encoding

| Method | Description |
|---|---|
| `Write(pcm []byte) (int, error)` | Encode PCM (int16 LE interleaved). Auto-initializes on first call. |
| `Flush() (int, error)` | Flush remaining MP3 data |
| `GetLametagFrame() []byte` | Get VBR/Xing header (call after Flush, before Close) |
| `Close() error` | Release resources. Idempotent. |

### Getters

| Method | Returns |
|---|---|
| `GetInSamplerate()` | Input sample rate |
| `GetOutSamplerate()` | Output sample rate |
| `GetNumChannels()` | Number of channels |
| `GetBrate()` | CBR bitrate |
| `GetVBR()` | VBR mode |
| `GetVBRQuality()` | VBR quality |
| `GetQuality()` | Algorithm quality |
| `GetVersion()` | MPEG version (1/2/0) |
| `GetFrameSize()` | Samples per MP3 frame |
| `GetEncoderDelay()` | Encoder delay samples |
| `GetEncoderPadding()` | Encoder padding samples |

## Concurrency

Each `Encoder` owns an isolated TLS context. Different encoders can be used from different goroutines concurrently. A single `Encoder` is **not** safe for concurrent use.

## Important notes

- **Always call `Close()`** — failing to close leaks memory
- **VBR header**: after `Flush()`, call `GetLametagFrame()` and write the result at file offset 0. Without this, players estimate duration from bitrate (inaccurate for VBR)
- **Input format**: `Write()` expects `[]byte` containing int16 little-endian samples, interleaved for stereo (L R L R…)
- **ID3 tags**: if you need ID3v2 tags, disable LAME's built-in writer (`SetWriteID3TagAutomatic(false)`) and write tags after closing the encoder using a library like [bogem/id3v2](https://github.com/bogem/id3v2)
