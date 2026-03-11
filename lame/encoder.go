// Package lame provides a pure-Go MP3 encoder built on LAME 3.100,
// transpiled from C to Go using modernc.org/ccgo. No CGO or system C
// libraries are required — the transpiled code uses modernc.org/libc
// as its runtime.
//
// Each [Encoder] owns an isolated TLS (thread-local storage) context,
// so separate encoders may be used concurrently from different goroutines.
// A single Encoder is NOT safe for concurrent use.
//
// [Encoder.Close] must be called to release C-side resources. Failing to
// close an encoder leaks memory.
//
// Supported platforms are determined by the transpiled lame_<OS>_<ARCH>.go
// files: currently darwin/arm64 and linux/arm64. See build/Dockerfile.ccgo
// to generate additional platforms.
//
// # Encoding workflow
//
// The typical encoding sequence is:
//
//	enc := lame.NewEncoder(w)
//	defer enc.Close()
//	enc.SetInSamplerate(44100)
//	enc.SetNumChannels(2)
//	enc.SetVBR(lame.VBRDefault)
//	enc.SetVBRQuality(2)
//	enc.Write(pcmData)  // int16 LE interleaved
//	enc.Flush()
//	frame := enc.GetLametagFrame()  // VBR header — write to file offset 0
//
// # VBR header (lametag frame)
//
// After Flush, call [Encoder.GetLametagFrame] to obtain the Xing/LAME header.
// Seek to the beginning of the output file and write this frame before calling
// Close. Without this header players estimate duration from bitrate, which is
// inaccurate for VBR files.
package lame

import (
	"errors"
	"fmt"
	"io"
	"unsafe"

	"modernc.org/libc"
)

// VBR mode constants for use with [Encoder.SetVBR].
const (
	VBROff     = 0 // Constant bitrate (CBR)
	VBRDefault = 4 // Variable bitrate, LAME's "new VBR" algorithm (vbr_mtrh)
)

// Encoder wraps the LAME encoder for converting PCM audio to MP3.
// Create with [NewEncoder], configure with Set* methods, then feed PCM
// data via [Encoder.Write]. Call [Encoder.Flush] when done, optionally
// retrieve the VBR header with [Encoder.GetLametagFrame], then [Encoder.Close].
type Encoder struct {
	tls         *libc.TLS
	gfp         uintptr
	w           io.Writer
	closed      bool
	initialized bool
}

// NewEncoder creates a new LAME encoder that writes MP3 data to w.
// Returns nil if LAME initialization fails (out of memory).
func NewEncoder(w io.Writer) *Encoder {
	tls := libc.NewTLS()
	gfp := lame_init(tls)
	if gfp == 0 {
		return nil
	}
	return &Encoder{
		tls: tls,
		gfp: gfp,
		w:   w,
	}
}

// SetWriteID3TagAutomatic controls whether LAME writes ID3 tags automatically.
func (e *Encoder) SetWriteID3TagAutomatic(v bool) {
	val := int32(0)
	if v {
		val = 1
	}
	lame_set_write_id3tag_automatic(e.tls, e.gfp, val)
}

// SetInSamplerate sets the input sample rate in Hz.
func (e *Encoder) SetInSamplerate(rate int) error {
	if rc := lame_set_in_samplerate(e.tls, e.gfp, int32(rate)); rc < 0 {
		return fmt.Errorf("lame_set_in_samplerate(%d) failed: %d", rate, rc)
	}
	return nil
}

// SetNumChannels sets the number of input channels (1=mono, 2=stereo).
func (e *Encoder) SetNumChannels(n int) error {
	if rc := lame_set_num_channels(e.tls, e.gfp, int32(n)); rc < 0 {
		return fmt.Errorf("lame_set_num_channels(%d) failed: %d", n, rc)
	}
	return nil
}

// SetQuality sets the algorithm quality (0=best/slow, 9=worst/fast).
func (e *Encoder) SetQuality(q int) error {
	if rc := lame_set_quality(e.tls, e.gfp, int32(q)); rc < 0 {
		return fmt.Errorf("lame_set_quality(%d) failed: %d", q, rc)
	}
	return nil
}

// SetVBR sets the VBR mode. Use VBRDefault or VBROff.
func (e *Encoder) SetVBR(mode int) error {
	if rc := lame_set_VBR(e.tls, e.gfp, int32(mode)); rc < 0 {
		return fmt.Errorf("lame_set_VBR(%d) failed: %d", mode, rc)
	}
	return nil
}

// SetVBRQuality sets the VBR quality (0.0=best, 9.9=worst).
func (e *Encoder) SetVBRQuality(q float64) error {
	if rc := lame_set_VBR_quality(e.tls, e.gfp, float32(q)); rc < 0 {
		return fmt.Errorf("lame_set_VBR_quality(%f) failed: %d", q, rc)
	}
	return nil
}

// SetBrate sets the CBR bitrate in kbps.
func (e *Encoder) SetBrate(kbps int) error {
	if rc := lame_set_brate(e.tls, e.gfp, int32(kbps)); rc < 0 {
		return fmt.Errorf("lame_set_brate(%d) failed: %d", kbps, rc)
	}
	return nil
}

// initParams initializes LAME internal parameters. Called automatically
// on first Write if not called explicitly.
func (e *Encoder) initParams() error {
	if rc := lame_init_params(e.tls, e.gfp); rc < 0 {
		return fmt.Errorf("lame_init_params failed: %d", rc)
	}
	return nil
}

// Write encodes PCM audio and writes MP3 bytes to the underlying writer.
// pcm must contain int16 little-endian samples, interleaved for stereo
// (L R L R …). On first call, LAME parameters are initialized automatically.
// Returns len(pcm) on success to satisfy io.Writer semantics.
func (e *Encoder) Write(pcm []byte) (int, error) {
	if e.closed {
		return 0, errors.New("encoder is closed")
	}

	// Initialize params on first write (like go-lame does)
	if !e.initialized {
		if err := e.initParams(); err != nil {
			return 0, err
		}
		e.initialized = true
	}

	numChannels := int(lame_get_num_channels(e.tls, e.gfp))
	if numChannels < 1 {
		numChannels = 2
	}

	// pcm is int16 LE interleaved, so each sample is 2 bytes
	// nsamples = number of samples per channel
	bytesPerSample := 2
	totalSamples := len(pcm) / bytesPerSample
	nsamples := totalSamples / numChannels

	if nsamples == 0 {
		return len(pcm), nil
	}

	// MP3 output buffer: worst case is 1.25 * nsamples + 7200
	mp3bufSize := int(float64(nsamples)*1.25) + 7200
	mp3buf := libc.Xmalloc(e.tls, uint64(mp3bufSize))
	if mp3buf == 0 {
		return 0, errors.New("failed to allocate MP3 buffer")
	}
	defer libc.Xfree(e.tls, mp3buf)

	// Allocate C memory for PCM data
	pcmBuf := libc.Xmalloc(e.tls, uint64(len(pcm)))
	if pcmBuf == 0 {
		return 0, errors.New("failed to allocate PCM buffer")
	}
	defer libc.Xfree(e.tls, pcmBuf)

	// Copy PCM data to C memory
	copy(unsafe.Slice((*byte)(unsafe.Pointer(pcmBuf)), len(pcm)), pcm)

	// Encode PCM data. For mono, use lame_encode_buffer (non-interleaved) because
	// lame_encode_buffer_interleaved hardcodes jump=2 and reads out of bounds for mono.
	// For stereo, use lame_encode_buffer_interleaved with the interleaved buffer.
	var ret int32
	if numChannels == 1 {
		ret = lame_encode_buffer(e.tls, e.gfp, pcmBuf, pcmBuf, int32(nsamples), mp3buf, int32(mp3bufSize))
	} else {
		ret = lame_encode_buffer_interleaved(e.tls, e.gfp, pcmBuf, int32(nsamples), mp3buf, int32(mp3bufSize))
	}
	if ret < 0 {
		return 0, fmt.Errorf("lame_encode_buffer failed: %d", ret)
	}

	if ret > 0 {
		// Copy MP3 data from C memory and write to output
		mp3data := unsafe.Slice((*byte)(unsafe.Pointer(mp3buf)), int(ret))
		if _, err := e.w.Write(mp3data); err != nil {
			return 0, err
		}
	}

	return len(pcm), nil
}

// Flush flushes the LAME encoder buffers, writing any remaining MP3 data.
func (e *Encoder) Flush() (int, error) {
	if e.closed {
		return 0, errors.New("encoder is closed")
	}

	// Allocate buffer for flush output
	mp3bufSize := 7200
	mp3buf := libc.Xmalloc(e.tls, uint64(mp3bufSize))
	if mp3buf == 0 {
		return 0, errors.New("failed to allocate flush buffer")
	}
	defer libc.Xfree(e.tls, mp3buf)

	ret := lame_encode_flush(e.tls, e.gfp, mp3buf, int32(mp3bufSize))
	if ret < 0 {
		return 0, fmt.Errorf("lame_encode_flush failed: %d", ret)
	}

	if ret > 0 {
		mp3data := unsafe.Slice((*byte)(unsafe.Pointer(mp3buf)), int(ret))
		if _, err := e.w.Write(mp3data); err != nil {
			return int(ret), err
		}
	}

	return int(ret), nil
}

// GetLametagFrame returns the Xing/LAME VBR header frame containing
// accurate duration and bitrate metadata. Write this frame at byte
// offset 0 of the MP3 file for correct playback duration reporting.
//
// Must be called after [Encoder.Flush] and before [Encoder.Close].
// Returns nil if no tag is available (CBR without info tag, or
// encoding hasn't started).
func (e *Encoder) GetLametagFrame() []byte {
	// First call with nil buffer to get required size
	needed := lame_get_lametag_frame(e.tls, e.gfp, 0, 0)
	if needed == 0 {
		return nil
	}

	buf := libc.Xmalloc(e.tls, uint64(needed))
	if buf == 0 {
		return nil
	}
	defer libc.Xfree(e.tls, buf)

	written := lame_get_lametag_frame(e.tls, e.gfp, buf, needed)
	if written == 0 || written > needed {
		return nil
	}

	result := make([]byte, written)
	copy(result, unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(written)))
	return result
}

// Close releases the LAME encoder state and TLS context.
// Safe to call multiple times. The encoder cannot be used after Close.
func (e *Encoder) Close() error {
	if e.closed {
		return nil
	}
	e.closed = true
	lame_close(e.tls, e.gfp)
	e.tls.Close()
	return nil
}

// MPEG channel mode constants for use with [Encoder.SetMode].
// The default is ModeJointStereo which lets LAME switch between
// stereo and mid/side coding per-frame for better compression.
const (
	ModeStereo      = 0 // L/R stereo, independent channels
	ModeJointStereo = 1 // Joint stereo — mid/side adaptive (default)
	ModeDualChannel = 2 // Dual channel — independent, equal bitrate
	ModeMono        = 3 // Mono
)

// Preset constants for use with [Encoder.SetPreset].
// Presets configure VBR quality, filters, and psychoacoustic tuning
// in a single call. They override individual Set* calls.
const (
	PresetMedium   = 1001 // ~150-180 kbps VBR, good quality
	PresetStandard = 1002 // ~170-210 kbps VBR, transparent for most listeners
	PresetExtreme  = 1003 // ~210-260 kbps VBR, near-transparent
	PresetInsane   = 1004 // 320 kbps CBR, maximum quality
)

// GetEncoderDelay returns the number of samples added at the start by the encoder.
func (e *Encoder) GetEncoderDelay() int {
	return int(lame_get_encoder_delay(e.tls, e.gfp))
}

// GetEncoderPadding returns the number of samples added at the end by the encoder.
func (e *Encoder) GetEncoderPadding() int {
	return int(lame_get_encoder_padding(e.tls, e.gfp))
}

// GetFrameSize returns the number of samples per MP3 frame.
func (e *Encoder) GetFrameSize() int {
	return int(lame_get_framesize(e.tls, e.gfp))
}

// GetVersion returns the MPEG version (1=MPEG1, 2=MPEG2, 0=MPEG2.5).
func (e *Encoder) GetVersion() int {
	return int(lame_get_version(e.tls, e.gfp))
}

// GetOutSamplerate returns the output sample rate in Hz.
func (e *Encoder) GetOutSamplerate() int {
	return int(lame_get_out_samplerate(e.tls, e.gfp))
}

// GetNumChannels returns the number of input channels.
func (e *Encoder) GetNumChannels() int {
	return int(lame_get_num_channels(e.tls, e.gfp))
}

// GetInSamplerate returns the input sample rate in Hz.
func (e *Encoder) GetInSamplerate() int {
	return int(lame_get_in_samplerate(e.tls, e.gfp))
}

// GetBrate returns the CBR bitrate in kbps.
func (e *Encoder) GetBrate() int {
	return int(lame_get_brate(e.tls, e.gfp))
}

// GetVBR returns the current VBR mode.
func (e *Encoder) GetVBR() int {
	return int(lame_get_VBR(e.tls, e.gfp))
}

// GetVBRQuality returns the VBR quality setting (0.0=best, 9.9=worst).
func (e *Encoder) GetVBRQuality() float64 {
	return float64(lame_get_VBR_quality(e.tls, e.gfp))
}

// GetQuality returns the algorithm quality setting (0=best/slow, 9=worst/fast).
func (e *Encoder) GetQuality() int {
	return int(lame_get_quality(e.tls, e.gfp))
}

// SetOutSamplerate sets the output sample rate in Hz (0=auto-select).
func (e *Encoder) SetOutSamplerate(rate int) error {
	if rc := lame_set_out_samplerate(e.tls, e.gfp, int32(rate)); rc < 0 {
		return fmt.Errorf("lame_set_out_samplerate(%d) failed: %d", rate, rc)
	}
	return nil
}

// SetMode sets the MPEG channel mode. Use ModeStereo, ModeJointStereo, ModeDualChannel, or ModeMono.
func (e *Encoder) SetMode(mode int) error {
	if rc := lame_set_mode(e.tls, e.gfp, MPEG_mode(mode)); rc < 0 {
		return fmt.Errorf("lame_set_mode(%d) failed: %d", mode, rc)
	}
	return nil
}

// SetScale sets the input PCM scaling factor applied before encoding.
func (e *Encoder) SetScale(scale float64) error {
	if rc := lame_set_scale(e.tls, e.gfp, float32(scale)); rc < 0 {
		return fmt.Errorf("lame_set_scale(%f) failed: %d", scale, rc)
	}
	return nil
}

// SetLowpassFreq sets the lowpass filter cutoff frequency in Hz (0=auto, -1=disabled).
func (e *Encoder) SetLowpassFreq(freq int) error {
	if rc := lame_set_lowpassfreq(e.tls, e.gfp, int32(freq)); rc < 0 {
		return fmt.Errorf("lame_set_lowpassfreq(%d) failed: %d", freq, rc)
	}
	return nil
}

// SetHighpassFreq sets the highpass filter cutoff frequency in Hz (0=auto, -1=disabled).
func (e *Encoder) SetHighpassFreq(freq int) error {
	if rc := lame_set_highpassfreq(e.tls, e.gfp, int32(freq)); rc < 0 {
		return fmt.Errorf("lame_set_highpassfreq(%d) failed: %d", freq, rc)
	}
	return nil
}

// SetPreset configures the encoder using a LAME preset. Use PresetMedium, PresetStandard,
// PresetExtreme, or PresetInsane.
func (e *Encoder) SetPreset(preset int) error {
	if rc := lame_set_preset(e.tls, e.gfp, int32(preset)); rc < 0 {
		return fmt.Errorf("lame_set_preset(%d) failed: %d", preset, rc)
	}
	return nil
}

// SetVBRMeanBitrateKbps sets the target average bitrate for ABR/VBR modes in kbps.
func (e *Encoder) SetVBRMeanBitrateKbps(kbps int) error {
	if rc := lame_set_VBR_mean_bitrate_kbps(e.tls, e.gfp, int32(kbps)); rc < 0 {
		return fmt.Errorf("lame_set_VBR_mean_bitrate_kbps(%d) failed: %d", kbps, rc)
	}
	return nil
}

// GetVBRMeanBitrateKbps returns the target average bitrate for ABR/VBR modes in kbps.
func (e *Encoder) GetVBRMeanBitrateKbps() int {
	return int(lame_get_VBR_mean_bitrate_kbps(e.tls, e.gfp))
}

// SetVBRMinBitrateKbps sets the minimum allowed bitrate in VBR mode in kbps.
func (e *Encoder) SetVBRMinBitrateKbps(kbps int) error {
	if rc := lame_set_VBR_min_bitrate_kbps(e.tls, e.gfp, int32(kbps)); rc < 0 {
		return fmt.Errorf("lame_set_VBR_min_bitrate_kbps(%d) failed: %d", kbps, rc)
	}
	return nil
}

// GetVBRMinBitrateKbps returns the minimum allowed bitrate in VBR mode in kbps.
func (e *Encoder) GetVBRMinBitrateKbps() int {
	return int(lame_get_VBR_min_bitrate_kbps(e.tls, e.gfp))
}

// SetVBRMaxBitrateKbps sets the maximum allowed bitrate in VBR mode in kbps.
func (e *Encoder) SetVBRMaxBitrateKbps(kbps int) error {
	if rc := lame_set_VBR_max_bitrate_kbps(e.tls, e.gfp, int32(kbps)); rc < 0 {
		return fmt.Errorf("lame_set_VBR_max_bitrate_kbps(%d) failed: %d", kbps, rc)
	}
	return nil
}

// GetVBRMaxBitrateKbps returns the maximum allowed bitrate in VBR mode in kbps.
func (e *Encoder) GetVBRMaxBitrateKbps() int {
	return int(lame_get_VBR_max_bitrate_kbps(e.tls, e.gfp))
}

// SetNumSamples sets the total number of input PCM samples (per channel).
// Setting this before encoding enables LAME to produce a more accurate VBR header.
func (e *Encoder) SetNumSamples(n uint64) error {
	if rc := lame_set_num_samples(e.tls, e.gfp, n); rc < 0 {
		return fmt.Errorf("lame_set_num_samples(%d) failed: %d", n, rc)
	}
	return nil
}

// GetNumSamples returns the total number of input PCM samples (per channel).
func (e *Encoder) GetNumSamples() uint64 {
	return lame_get_num_samples(e.tls, e.gfp)
}

// SetDisableReservoir disables the bit reservoir. When disabled, each MP3 frame
// is independently decodable, which is useful for streaming applications.
func (e *Encoder) SetDisableReservoir(disable bool) error {
	val := int32(0)
	if disable {
		val = 1
	}
	if rc := lame_set_disable_reservoir(e.tls, e.gfp, val); rc < 0 {
		return fmt.Errorf("lame_set_disable_reservoir(%d) failed: %d", val, rc)
	}
	return nil
}

// GetDisableReservoir returns whether the bit reservoir is disabled.
func (e *Encoder) GetDisableReservoir() bool {
	return lame_get_disable_reservoir(e.tls, e.gfp) != 0
}

// SetStrictISO enables strict ISO compliance mode.
func (e *Encoder) SetStrictISO(strict bool) error {
	val := int32(0)
	if strict {
		val = 1
	}
	if rc := lame_set_strict_ISO(e.tls, e.gfp, val); rc < 0 {
		return fmt.Errorf("lame_set_strict_ISO(%d) failed: %d", val, rc)
	}
	return nil
}

// GetStrictISO returns whether strict ISO compliance mode is enabled.
func (e *Encoder) GetStrictISO() bool {
	return lame_get_strict_ISO(e.tls, e.gfp) != 0
}

// GetTotalFrames returns the estimated total number of MP3 frames for the input.
// Requires SetNumSamples to have been called for an accurate estimate.
func (e *Encoder) GetTotalFrames() int {
	return int(lame_get_totalframes(e.tls, e.gfp))
}

// GetFrameNum returns the number of frames encoded so far.
func (e *Encoder) GetFrameNum() int {
	return int(lame_get_frameNum(e.tls, e.gfp))
}

// GetMode returns the current MPEG channel mode.
func (e *Encoder) GetMode() int {
	return int(lame_get_mode(e.tls, e.gfp))
}

// FlushNogap flushes the encoder without adding the end-of-stream padding.
// This allows gapless encoding of consecutive tracks — after FlushNogap,
// the encoder can accept more PCM data via Write for the next track.
func (e *Encoder) FlushNogap() (int, error) {
	if e.closed {
		return 0, errors.New("encoder is closed")
	}

	mp3bufSize := 7200
	mp3buf := libc.Xmalloc(e.tls, uint64(mp3bufSize))
	if mp3buf == 0 {
		return 0, errors.New("failed to allocate flush buffer")
	}
	defer libc.Xfree(e.tls, mp3buf)

	ret := lame_encode_flush_nogap(e.tls, e.gfp, mp3buf, int32(mp3bufSize))
	if ret < 0 {
		return 0, fmt.Errorf("lame_encode_flush_nogap failed: %d", ret)
	}

	if ret > 0 {
		mp3data := unsafe.Slice((*byte)(unsafe.Pointer(mp3buf)), int(ret))
		if _, err := e.w.Write(mp3data); err != nil {
			return int(ret), err
		}
	}

	return int(ret), nil
}
