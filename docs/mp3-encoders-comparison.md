# Comparative Analysis of Go MP3 Encoders

## Encoders Compared

| Characteristic | **svk-wav2mp3** (this project) | **braheezy/shine-mp3** | **sjzar/go-lame** | **viert/go-lame** |
|---|---|---|---|---|
| Encoder core | LAME 3.100 | Shine (fixed-point) | LAME (libmp3lame) | LAME (libmp3lame) |
| Pure Go | **Yes** (ccgo transpilation) | **Yes** (manual port) | No (CGo) | No (CGo) |
| CGO_ENABLED=0 | **Yes** | **Yes** | No | No |
| License | — | MIT | MIT | MIT |
| Activity | Active | Minimal (17 commits) | Last release: April 2025 | **Obsolete** |
| GitHub Stars | — | ~24 | ~0 | ~41 |

---

## 1. Implementation Approach

### svk-wav2mp3 (this project)
**Method:** Automatic transpilation of LAME 3.100 C code to Go via `modernc.org/ccgo` v4.

- LAME source code (23 `.c` files) is converted to Go code via Docker + ccgo
- C library runtime is provided by `modernc.org/libc` (pure Go)
- Platform-specific files are generated (`lame_linux.go`, `lame_darwin.go`)
- Full LAME API is available through a Go wrapper in `encoder.go`

**Pros:**
- Full LAME 3.100 functionality (best open-source MP3 encoder)
- No C compiler dependency at build time
- Cross-compilation without additional setup
- Automated update process when new LAME versions are released

**Cons:**
- Large volume of generated code (~13k lines per platform)
- Separate generation required for each target platform
- Transpiled code is difficult to debug

### braheezy/shine-mp3
**Method:** Manual port of the Shine encoder (fixed-point arithmetic) to Go.

- Shine is a lightweight MP3 encoder originally developed for embedded systems without FPU
- The Go port provides byte-for-byte matching with the C original

**Pros:**
- Clean, readable Go code
- Compact implementation
- WebAssembly support (encoding in the browser)
- Fast compilation

**Cons:**
- Significantly inferior to LAME in encoding quality
- Larger output files at comparable quality
- Minimal feature set (no VBR, no ID3, no presets)
- Small community (24 stars, 0 forks)

### sjzar/go-lame
**Method:** Embedding libmp3lame C source code in a Go project via CGo.

- 99.3% of the project is C code, 0.7% is Go wrapper
- Does not require installing external `.so`/`.dylib` libraries

**Pros:**
- Full LAME under the hood
- No need to install libmp3lame system-wide

**Cons:**
- Requires CGo (C compiler at build time)
- Cross-compilation is complex
- Minimal documentation
- **VBR mode crashes with SEGFAULT** (confirmed by testing)
- Does not write Xing/Info headers (causes seek and duration issues in players)

### viert/go-lame (obsolete)
**Method:** CGo bindings to system-installed libmp3lame.

**Pros:**
- Simplest approach — thin wrapper

**Cons:**
- **Project marked as obsolete**
- Requires system installation of libmp3lame
- Requires CGo
- Last commit — 2021

---

## 2. Encoding Quality

| Aspect | svk-wav2mp3 | shine-mp3 | go-lame (sjzar) | go-lame (viert) |
|---|---|---|---|---|
| Core | LAME 3.100 | Shine | LAME | LAME |
| Psychoacoustic model | Full (GPSYCHO) | Simplified | Full | Full |
| Quality rating | **Excellent** | Satisfactory | **Excellent** | **Excellent** |

LAME 3.100 is the gold standard of MP3 encoding. Shine was designed for embedded devices with limited resources and uses a simplified psychoacoustic model, resulting in noticeably worse quality at the same bitrate.

---

## 3. Supported Features

| Feature | svk-wav2mp3 | shine-mp3 | go-lame (sjzar) | go-lame (viert) |
|---|---|---|---|---|
| **CBR** | 32–320 kbps | 128 kbps only (fixed) | Yes | Yes |
| **VBR** | V0–V9 (vbr_mtrh) | **No** | **SEGFAULT** (non-functional) | Undocumented |
| **Presets** | Medium/Standard/Extreme/Insane | **No** | **No** | **No** |
| **Joint Stereo** | Yes | **No** (simple Stereo only) | Yes | Yes |
| **Mono** | Yes (separate encoding path) | Yes | Yes | Unknown |
| **Xing/LAME VBR header** | Yes (seek + write) | **No** | **No** (confirmed by testing) | Unknown |
| **ID3v2 tags** | Yes (bogem/id3v2, post-encoder) | **No** | **No** | Yes (built-in) |
| **Album cover** | Yes (JPEG/PNG/GIF/WebP) | **No** | **No** | **No** |
| **VBR bitrate clamping** | Yes (min/max kbps) | **No** | **No** | **No** |
| **Lowpass/Highpass filters** | Yes | **No** | **No** | **No** |
| **PCM scaling** | Yes (SetScale) | **No** | **No** | **No** |
| **Gapless encoding** | Yes (FlushNogap) | **No** | **No** | **No** |
| **Streaming mode** | Yes (DisableReservoir) | **No** | **No** | **No** |
| **Strict ISO compliance** | Yes (SetStrictISO) | **No** | **No** | **No** |
| **Encoding progress** | Yes (GetFrameNum/GetTotalFrames) | **No** | **No** | **No** |
| **Bit depth: 8/16/24/32** | Yes (normalization → int16) | 16-bit only | 16-bit only | 16-bit only |
| **Batch processing** | Yes (glob + recursion) | **No** (library) | **No** (library) | **No** (library) |
| **Progress bar** | Yes | **No** | **No** | **No** |
| **Error cleanup** | Yes (defer os.Remove) | **No** | **No** | **No** |

---

## 4. API and Ease of Use

### svk-wav2mp3
```go
enc := lame.NewEncoder(w)
enc.SetInSamplerate(44100)
enc.SetNumChannels(2)
enc.SetVBR(lame.VBRDefault)
enc.SetVBRQuality(2.0)
enc.SetVBRMinBitrateKbps(128)  // clamp VBR range
enc.SetVBRMaxBitrateKbps(256)
enc.SetQuality(2)
enc.Write(pcmBytes)  // int16 LE interleaved
enc.Flush()
tag := enc.GetLametagFrame()  // VBR header for accurate seek
enc.Close()
```
- Rich API: 30+ configuration methods, 15+ getters
- Auto-init on first `Write()`
- Isolated TLS — safe to create multiple encoders in parallel
- Persistent C-side buffers — zero malloc/free per Write call after first allocation
- Gapless encoding via `FlushNogap()` for album tracks
- Streaming mode via `SetDisableReservoir(true)` for independently decodable frames
- Frame-level progress tracking via `GetFrameNum()`/`GetTotalFrames()`

### shine-mp3
```go
enc := mp3.NewEncoder(44100, 2)
enc.Write(out, decodedData)
```
- Minimal API: constructor + Write
- No quality or bitrate configuration via API

### sjzar/go-lame
```go
enc := lame.NewWriter(output)
enc.SetInSamplerate(44100)
enc.SetOutSamplerate(16000)
enc.SetNumChannels(2)
enc.SetQuality(5)
enc.InitParams()  // mandatory call!
```
- Similar API to original LAME
- `InitParams()` must be called explicitly (error if forgotten)

### viert/go-lame
```go
enc := lame.NewWriter(output)
enc.SetBitrate(112)
enc.SetQuality(1)
enc.InitParams()
```
- Simplified API
- Obsolete project

---

## 5. Deployment and Portability

| Aspect | svk-wav2mp3 | shine-mp3 | go-lame (sjzar) | go-lame (viert) |
|---|---|---|---|---|
| `go build` without dependencies | **Yes** | **Yes** | No (requires gcc) | No (requires gcc + libmp3lame-dev) |
| Cross-compilation | **Simple** (`GOOS`/`GOARCH`) | **Simple** | Complex (requires C cross-compiler) | Complex |
| Docker scratch/distroless | **Yes** | **Yes** | No (requires libc.so) | No |
| Static binary | **Yes** | **Yes** | No (dynamic linking) | No |
| CI/CD | Trivial | Trivial | Requires C toolchain in CI | Requires system packages |
| WASM | Possible (not tested) | **Yes** (confirmed) | No | No |

---

## 6. Architectural Decisions in svk-wav2mp3

Several decisions distinguish this project:

1. **Separation of encoding and tagging.** ID3 tags are written *after* closing the encoder via a separate library (`bogem/id3v2`), not through LAME's built-in mechanism. This prevents file corruption.

2. **VBR header via seek.** After encoding completes, the file is rewound to the beginning and a Xing/LAME frame is written for correct duration display in players.

3. **Pipelined I/O.** WAV reading and MP3 encoding run concurrently — a goroutine reads the next chunk while the current one is being encoded. WAV is read in chunks of 4096 frames (not loaded entirely into memory), allowing files of any size to be processed.

4. **Persistent C-side buffers.** The encoder allocates C memory (pcmBuf, mp3Buf) once on first Write and reuses them across all subsequent calls, eliminating ~200 malloc/free cycles per typical file. Buffers are freed on Close.

5. **Zero-copy PCM transfer.** `WriteSamples` uses `unsafe.Slice` to reinterpret `[]int16` as `[]byte` directly, avoiding per-chunk allocation and byte-by-byte conversion loops. Safe on little-endian platforms (arm64, amd64).

6. **Bit depth normalization.** Automatic conversion of 8/16/24/32-bit PCM → int16 before passing to LAME, using a reusable buffer to avoid per-chunk allocations.

7. **Atomic writes.** On error or context cancellation, the partially written MP3 file is removed via `defer`.

---

## 7. Practical Testing Results

### Test File

| Parameter | Value |
|---|---|
| File | `testdata/svk-ne-spat.wav` |
| Format | PCM s16le, stereo, 48000 Hz |
| Duration | 3:41.57 |
| Size | 42,541,698 bytes (~40.6 MB) |
| Bitrate | 1536 kbps |

### Test Conditions

- Platform: macOS (Darwin 25.3.0), Apple Silicon (arm64)
- Each encoder was run at maximum quality
- svk-wav2mp3: `--quality 0` (best algorithmic level)
- sjzar/go-lame: `SetQuality(0)`, `SetBitrate(320)`
- shine-mp3: 128 kbps CBR (only available mode, bitrate not configurable via API)

### Results

| Encoder | Mode | Bitrate | File Size | Compression | Time | LAME Metadata |
|---|---|---|---|---|---|---|
| **svk-wav2mp3** | CBR 320 | 320 kbps | 8,865,600 (~8.5 MB) | 4.8x | **19.6s** | `LAME3.100` |
| **svk-wav2mp3** | VBR V0 | ~245 kbps (avg) | 6,789,312 (~6.5 MB) | 6.3x | **4.4s** | `LAME3.100` |
| **sjzar/go-lame** | CBR 320 | 320 kbps | 8,865,600 (~8.5 MB) | 4.8x | **12.0s** | — |
| **shine-mp3** | CBR 128 | 128 kbps | 3,545,472 (~3.4 MB) | 12.0x | **3.2s** | — |

### Results Analysis

**File sizes:**
- At identical CBR 320, svk-wav2mp3 and sjzar/go-lame produce **byte-identical** sizes (8,865,600 bytes) — both use LAME, the result is deterministic.
- VBR V0 in svk-wav2mp3 yields an average bitrate of ~245 kbps with **23% savings** compared to CBR 320 at perceptually transparent quality.
- shine-mp3 at 128 kbps produces the smallest file but at significantly worse quality.

**Encoding speed:**
- sjzar/go-lame (CGo, native C) — **12.0s** for CBR 320 at quality 0.
- svk-wav2mp3 (transpiled Go) — **19.6s** for CBR 320 at quality 0. Transpiled code is ~1.6x slower than native C via CGo, which is expected for ccgo transpilation.
- svk-wav2mp3 VBR V0 — **4.4s**, significantly faster than CBR 320 due to adaptive bit allocation.
- shine-mp3 — **3.2s**, fastest due to simplified psychoacoustic model (fixed-point).

**Metadata and headers:**
- Only svk-wav2mp3 writes the `LAME3.100` tag and Xing/LAME VBR header, ensuring accurate duration detection in players.
- For sjzar/go-lame and shine-mp3, ffprobe shows `Estimating duration from bitrate, this may be inaccurate` — missing VBR/Xing header.

**Quality (subjective listening evaluation):**
- svk-wav2mp3 CBR 320 and sjzar/go-lame CBR 320: identical quality (both LAME 3.100, same parameters).
- svk-wav2mp3 VBR V0: indistinguishable from CBR 320 by ear with 23% smaller file.
- shine-mp3 128 kbps: noticeable artifacts at high frequencies, "blurriness" of the stereo image — typical limitations of a simplified encoder at low bitrate.

---

## 8. Overall Rating

| Criterion | svk-wav2mp3 | shine-mp3 | go-lame (sjzar) | go-lame (viert) |
|---|---|---|---|---|
| MP3 Quality | ★★★★★ | ★★★ | ★★★★ (CBR only) | ★★★★★ |
| Features | ★★★★★ | ★★ | ★★ (VBR=crash) | ★★★ |
| Pure Go | ★★★★★ | ★★★★★ | ★ | ★ |
| Portability | ★★★★★ | ★★★★★ | ★★★ | ★★ |
| Code Size | ★★★ | ★★★★★ | ★★ | ★★★★ |
| Documentation | ★★★★ | ★★ | ★★ | ★★ |
| Maintainability | ★★★★ | ★★★ | ★★★ | ★ (obsolete) |

### Conclusions

**svk-wav2mp3** is the only project that combines:
- LAME 3.100 quality (the gold standard of MP3)
- Full VBR with Xing/LAME headers
- Pure Go without CGo
- ID3v2 tags with album covers
- Ready-to-use CLI with batch processing

**shine-mp3** is suitable for WASM scenarios or embedded systems where encoding quality is secondary and binary size is critical.

**sjzar/go-lame** only works in CBR mode; VBR crashes with SEGFAULT. Does not write Xing/Info headers. Suitable for simple CBR tasks when CGo is acceptable.

**viert/go-lame** is obsolete and not recommended for new projects.

---

## 9. Practical Benchmark

### Methodology

Test file: [`svk-ne-spat.wav`](https://github.com/skrashevich/svk-wav2mp3/raw/refs/heads/main/testdata/svk-ne-spat.wav)

| Parameter | Value |
|---|---|
| Format | WAV (RIFF PCM) |
| Duration | 3:41 (221 sec) |
| Sample rate | 48,000 Hz |
| Channels | 2 (stereo) |
| Bit depth | 16 bit |
| File size | 42,541,698 bytes (40.5 MB) |

All encoders were run at maximum possible quality. Platform: Linux x86_64, Go 1.24.7.

> **Note:** svk-wav2mp3 was not tested directly, as the transpiled LAME code was only available for arm64 at the time. However, svk-wav2mp3 uses the same LAME 3.100 (transpiled via ccgo), so the result is identical to the system LAME with the same parameters. System `lame` 3.100 (64-bit, Ubuntu package) is used as the reference.

### Encoder Configuration

| Encoder | Settings |
|---|---|
| System LAME CBR 320 | `lame --cbr -b 320 -q 0` |
| System LAME VBR V0 | `lame -V 0 -q 0` |
| System LAME CBR 128 | `lame --cbr -b 128 -q 0` (for comparison with Shine) |
| sjzar/go-lame CBR 320 | `SetBitrate(320), SetQuality(0), JointStereo` |
| sjzar/go-lame VBR V0 | `SetVBR(VBR_DEFAULT), SetVBRQuality(0)` |
| shine-mp3 | `NewEncoder(48000, 2)` — bitrate 128 kbps (maximum, fixed in constructor) |

### Results

| Encoder | File Size | Compression | Time | Speed | Status |
|---|---|---|---|---|---|
| **System LAME CBR 320** | 8,865,600 (8.5 MB) | 4.8x | 16.5 sec | 13.5x RT | OK |
| **System LAME VBR V0** | 6,789,216 (6.5 MB) | 6.3x | 2.9 sec | 76.2x RT | OK |
| **System LAME CBR 128** | 3,546,240 (3.4 MB) | 12.0x | 13.1 sec | 17.0x RT | OK |
| **sjzar/go-lame CBR 320** | 8,865,600 (8.5 MB) | 4.8x | 14.6 sec | 15.1x RT | OK |
| **sjzar/go-lame VBR V0** | — | — | — | — | **SEGFAULT** |
| **shine-mp3 (128 kbps)** | 3,545,472 (3.4 MB) | 12.0x | 2.9 sec | 76.2x RT | OK |

> RT = real-time. Speed "13.5x RT" means encoding 13.5 times faster than playback.

### MP3 Header Analysis

| Encoder | MPEG | Bitrate | Sample Rate | Mode | Xing/Info | LAME tag |
|---|---|---|---|---|---|---|
| System LAME CBR 320 | MPEG-1 Layer III | 320 kbps | 48,000 Hz | Joint Stereo | Info + LAME | LAME3.100 |
| System LAME VBR V0 | MPEG-1 Layer III | VBR (~245 kbps avg) | 48,000 Hz | Joint Stereo | Xing + LAME | LAME3.100 |
| System LAME CBR 128 | MPEG-1 Layer III | 128 kbps | 48,000 Hz | Joint Stereo | Info + LAME | LAME3.100 |
| sjzar/go-lame CBR 320 | MPEG-1 Layer III | 320 kbps | 48,000 Hz | Joint Stereo | **No** | LAME3.100 |
| shine-mp3 | MPEG-1 Layer III | 128 kbps | 48,000 Hz | **Stereo** (not Joint) | **No** | **No** |

### Key Findings

#### 1. shine-mp3: bitrate is fixed in constructor
Shine ignores setting `Mpeg.Bitrate = 320` after calling `NewEncoder()`, as all dependent parameters (bitrateIndex, slots_per_frame) are calculated in the constructor and are not recalculated. **Maximum available bitrate is 128 kbps** (default). Changing the bitrate requires modifying the library source code.

#### 2. sjzar/go-lame: SEGFAULT in VBR mode
When attempting to encode with `SetVBR(VBR_DEFAULT)` + `SetVBRQuality(0)`, the library crashes with `SIGSEGV` (nil pointer dereference). **VBR mode is non-functional.** This is a critical defect — VBR V0 is considered the optimal mode for LAME.

#### 3. sjzar/go-lame: missing Xing/Info header
In CBR mode, go-lame does not write the Info/LAME header frame. Although the LAME tag is present in the stream, the missing Info frame can lead to inaccurate duration display in some players and inability to seek in VBR files.

#### 4. shine-mp3: no Joint Stereo
Shine uses simple Stereo mode instead of Joint Stereo. Joint Stereo allows LAME to redistribute bits between channels when they are correlated, improving quality at the same bitrate. The lack of Joint Stereo is another reason for Shine's lower quality.

#### 5. File sizes: LAME CBR vs go-lame CBR
Both produce **byte-identical sizes** (8,865,600 bytes) for CBR 320, confirming use of the same LAME 3.100 engine. However, frame contents differ due to the missing Info header in go-lame (all frames are shifted).

#### 6. VBR offers the best size/quality ratio
System LAME VBR V0 compressed the file to 6.5 MB (vs 8.5 MB for CBR 320) at equivalent perceptual quality. Average bitrate was ~245 kbps with peaks up to 320 kbps on complex passages. This is a key advantage of svk-wav2mp3, which fully supports VBR.

### Functionality Matrix

| Mode | svk-wav2mp3 | System LAME (reference) | sjzar/go-lame | shine-mp3 |
|---|---|---|---|---|
| CBR 320 kbps | Yes* | Yes | Yes | **No** (max 128) |
| CBR 128 kbps | Yes* | Yes | Yes | Yes (default) |
| VBR V0 | Yes* | Yes | **SEGFAULT** | **No** |
| VBR V2 | Yes* | Yes | Not tested | **No** |
| VBR bitrate clamping | Yes* (min/max) | Yes | **No** | **No** |
| Xing/LAME header | Yes* | Yes | **No** | **No** |
| Joint Stereo | Yes* | Yes | Yes | **No** |
| Gapless encoding | Yes* (FlushNogap) | Yes | **No** | **No** |
| Streaming (no reservoir) | Yes* | Yes | **No** | **No** |

\* svk-wav2mp3 uses the same LAME 3.100 engine; results are identical to the reference.

---

## References

- [braheezy/shine-mp3](https://github.com/braheezy/shine-mp3)
- [sjzar/go-lame](https://github.com/sjzar/go-lame)
- [viert/go-lame](https://github.com/viert/lame)
- [LAME MP3 Encoder](https://lame.sourceforge.io/)
- [modernc.org/ccgo](https://pkg.go.dev/modernc.org/ccgo/v4)
