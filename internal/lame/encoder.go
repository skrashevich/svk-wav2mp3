// Package lame provides a pure-Go MP3 encoding API built on LAME 3.100
// source code transpiled from C to Go using modernc.org/ccgo.
package lame

import (
	"errors"
	"fmt"
	"io"
	"unsafe"

	"modernc.org/libc"
)

// VBR mode constants matching LAME's vbr_mode_e enum.
const (
	VBROff     = 0 // vbr_off: CBR mode
	VBRDefault = 4 // vbr_mtrh: VBR default (new VBR)
)

// Encoder wraps the LAME encoder for converting PCM audio to MP3.
type Encoder struct {
	tls         *libc.TLS
	gfp         uintptr
	w           io.Writer
	closed      bool
	initialized bool
}

// NewEncoder creates a new LAME encoder that writes MP3 data to w.
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

// Write encodes PCM data (int16 little-endian interleaved) and writes
// the resulting MP3 bytes to the underlying writer.
// This matches the go-lame Encoder.Write interface.
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

	// Encode using interleaved buffer
	ret := lame_encode_buffer_interleaved(e.tls, e.gfp, pcmBuf, int32(nsamples), mp3buf, int32(mp3bufSize))
	if ret < 0 {
		return 0, fmt.Errorf("lame_encode_buffer_interleaved failed: %d", ret)
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

// GetLametagFrame returns the LAME/Xing VBR header frame that should be
// written at the beginning of the MP3 file after encoding is complete.
// This header contains accurate duration and bitrate information.
// Returns nil if no tag is available (e.g. encoding hasn't started).
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

// Close releases all LAME resources. The encoder cannot be used after Close.
func (e *Encoder) Close() error {
	if e.closed {
		return nil
	}
	e.closed = true
	lame_close(e.tls, e.gfp)
	e.tls.Close()
	return nil
}
