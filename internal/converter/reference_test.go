package converter

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/bogem/id3v2/v2"
	"github.com/svk/wav2mp3/internal/config"
)

const (
	realWAV  = "../../testdata/svk-ne-spat.wav"
	coverPNG = "../../testdata/cover.png"
)

// probeResult holds ffprobe JSON output for an audio stream.
type probeResult struct {
	Duration   float64 // seconds
	Bitrate    int     // kb/s
	SampleRate int
	Channels   int
	Codec      string
}

// ffprobe runs ffprobe on path and returns parsed audio stream info.
func ffprobeAudio(t *testing.T, path string) probeResult {
	t.Helper()

	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		"-select_streams", "a:0",
		path,
	)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ffprobe %s: %v", path, err)
	}

	var raw struct {
		Streams []struct {
			CodecName  string `json:"codec_name"`
			SampleRate string `json:"sample_rate"`
			Channels   int    `json:"channels"`
			BitRate    string `json:"bit_rate"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &raw); err != nil {
		t.Fatalf("ffprobe JSON parse: %v", err)
	}

	var r probeResult
	if len(raw.Streams) > 0 {
		s := raw.Streams[0]
		r.Codec = s.CodecName
		fmt.Sscanf(s.SampleRate, "%d", &r.SampleRate)
		r.Channels = s.Channels
		fmt.Sscanf(s.BitRate, "%d", &r.Bitrate)
		r.Bitrate /= 1000 // bps → kbps
	}
	fmt.Sscanf(raw.Format.Duration, "%f", &r.Duration)
	return r
}

// ffmpegConvert converts WAV to MP3 using ffmpeg with given parameters.
func ffmpegConvert(t *testing.T, wavPath, mp3Path string, vbr bool, vbrQuality float64, bitrate int) {
	t.Helper()
	args := []string{"-y", "-i", wavPath}
	if vbr {
		// ffmpeg uses -q:a for VBR, scale 0-9 (0=best) matching LAME
		args = append(args, "-codec:a", "libmp3lame", "-q:a", fmt.Sprintf("%.0f", vbrQuality))
	} else {
		args = append(args, "-codec:a", "libmp3lame", "-b:a", fmt.Sprintf("%dk", bitrate))
	}
	args = append(args, mp3Path)

	cmd := exec.Command("ffmpeg", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg convert: %v\n%s", err, out)
	}
}

func skipIfNoFixture(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(realWAV); os.IsNotExist(err) {
		t.Skip("fixture missing: testdata/svk-ne-spat.wav")
	}
}

func skipIfNoFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skip("ffmpeg not found in PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		t.Skip("ffprobe not found in PATH")
	}
}

// assertDurationClose checks that two durations are within tolerance seconds.
func assertDurationClose(t *testing.T, name string, got, want, tolerance float64) {
	t.Helper()
	diff := math.Abs(got - want)
	if diff > tolerance {
		t.Errorf("%s: duration %.2fs, want %.2fs (±%.1fs), diff=%.2fs",
			name, got, want, tolerance, diff)
	}
}

// assertBitrateClose checks that bitrates are within tolerance percent.
func assertBitrateClose(t *testing.T, name string, got, want int, tolerancePct float64) {
	t.Helper()
	if want == 0 {
		return
	}
	diff := math.Abs(float64(got-want)) / float64(want) * 100
	if diff > tolerancePct {
		t.Errorf("%s: bitrate %d kbps, want %d kbps (±%.0f%%), diff=%.1f%%",
			name, got, want, tolerancePct, diff)
	}
}

func TestReferenceVBR(t *testing.T) {
	skipIfNoFixture(t)
	skipIfNoFFmpeg(t)

	tmpDir := t.TempDir()
	ourMP3 := filepath.Join(tmpDir, "ours_vbr.mp3")
	refMP3 := filepath.Join(tmpDir, "ref_vbr.mp3")

	vbrQuality := 0.0

	// Convert with our encoder
	opts := config.DefaultConvertOptions()
	opts.InputPath = realWAV
	opts.OutputPath = ourMP3
	opts.VBR = true
	opts.VBRQuality = vbrQuality
	opts.Quiet = true

	_, err := Convert(context.Background(), opts)
	if err != nil {
		t.Fatalf("our conversion failed: %v", err)
	}

	// Convert with ffmpeg (reference)
	ffmpegConvert(t, realWAV, refMP3, true, vbrQuality, 0)

	// Probe the source WAV for ground truth duration
	wavProbe := ffprobeAudio(t, realWAV)
	ourProbe := ffprobeAudio(t, ourMP3)
	refProbe := ffprobeAudio(t, refMP3)

	t.Logf("WAV duration: %.2fs", wavProbe.Duration)
	t.Logf("Our MP3: duration=%.2fs bitrate=%d kbps sr=%d ch=%d",
		ourProbe.Duration, ourProbe.Bitrate, ourProbe.SampleRate, ourProbe.Channels)
	t.Logf("Ref MP3: duration=%.2fs bitrate=%d kbps sr=%d ch=%d",
		refProbe.Duration, refProbe.Bitrate, refProbe.SampleRate, refProbe.Channels)

	// Duration must match WAV source within 0.5s
	assertDurationClose(t, "ours vs WAV", ourProbe.Duration, wavProbe.Duration, 0.5)
	assertDurationClose(t, "ref vs WAV", refProbe.Duration, wavProbe.Duration, 0.5)

	// Our duration must be close to ffmpeg reference within 0.5s
	assertDurationClose(t, "ours vs ref", ourProbe.Duration, refProbe.Duration, 0.5)

	// Bitrate should be in the same ballpark (within 25% of reference)
	assertBitrateClose(t, "VBR bitrate", ourProbe.Bitrate, refProbe.Bitrate, 25)

	// Sample rate and channels must match exactly
	if ourProbe.SampleRate != refProbe.SampleRate {
		t.Errorf("sample rate: ours=%d, ref=%d", ourProbe.SampleRate, refProbe.SampleRate)
	}
	if ourProbe.Channels != refProbe.Channels {
		t.Errorf("channels: ours=%d, ref=%d", ourProbe.Channels, refProbe.Channels)
	}
}

func TestReferenceCBR(t *testing.T) {
	skipIfNoFixture(t)
	skipIfNoFFmpeg(t)

	for _, bitrate := range []int{128, 192, 320} {
		t.Run(fmt.Sprintf("%dkbps", bitrate), func(t *testing.T) {
			tmpDir := t.TempDir()
			ourMP3 := filepath.Join(tmpDir, "ours_cbr.mp3")
			refMP3 := filepath.Join(tmpDir, "ref_cbr.mp3")

			// Convert with our encoder
			opts := config.ConvertOptions{
				InputPath:  realWAV,
				OutputPath: ourMP3,
				VBR:        false,
				Bitrate:    bitrate,
				Quality:    5,
				Quiet:      true,
			}
			_, err := Convert(context.Background(), opts)
			if err != nil {
				t.Fatalf("our conversion failed: %v", err)
			}

			// Convert with ffmpeg (reference)
			ffmpegConvert(t, realWAV, refMP3, false, 0, bitrate)

			wavProbe := ffprobeAudio(t, realWAV)
			ourProbe := ffprobeAudio(t, ourMP3)
			refProbe := ffprobeAudio(t, refMP3)

			t.Logf("WAV duration: %.2fs", wavProbe.Duration)
			t.Logf("Our MP3: duration=%.2fs bitrate=%d kbps sr=%d ch=%d",
				ourProbe.Duration, ourProbe.Bitrate, ourProbe.SampleRate, ourProbe.Channels)
			t.Logf("Ref MP3: duration=%.2fs bitrate=%d kbps sr=%d ch=%d",
				refProbe.Duration, refProbe.Bitrate, refProbe.SampleRate, refProbe.Channels)

			// Duration must match WAV source within 0.5s
			assertDurationClose(t, "ours vs WAV", ourProbe.Duration, wavProbe.Duration, 0.5)

			// Our duration must be close to ffmpeg reference within 0.5s
			assertDurationClose(t, "ours vs ref", ourProbe.Duration, refProbe.Duration, 0.5)

			// CBR bitrate should match target closely (within 5%)
			assertBitrateClose(t, "CBR bitrate vs target", ourProbe.Bitrate, bitrate, 5)

			// Bitrate should be close to ffmpeg reference (within 5%)
			assertBitrateClose(t, "CBR bitrate vs ref", ourProbe.Bitrate, refProbe.Bitrate, 5)

			// Sample rate and channels must match
			if ourProbe.SampleRate != refProbe.SampleRate {
				t.Errorf("sample rate: ours=%d, ref=%d", ourProbe.SampleRate, refProbe.SampleRate)
			}
			if ourProbe.Channels != refProbe.Channels {
				t.Errorf("channels: ours=%d, ref=%d", ourProbe.Channels, refProbe.Channels)
			}
		})
	}
}

func TestReferenceVBRQualities(t *testing.T) {
	skipIfNoFixture(t)
	skipIfNoFFmpeg(t)

	for _, q := range []float64{0, 2, 4, 6} {
		t.Run(fmt.Sprintf("V%.0f", q), func(t *testing.T) {
			tmpDir := t.TempDir()
			ourMP3 := filepath.Join(tmpDir, "ours.mp3")
			refMP3 := filepath.Join(tmpDir, "ref.mp3")

			opts := config.DefaultConvertOptions()
			opts.InputPath = realWAV
			opts.OutputPath = ourMP3
			opts.VBR = true
			opts.VBRQuality = q
			opts.Quiet = true

			_, err := Convert(context.Background(), opts)
			if err != nil {
				t.Fatalf("our conversion failed: %v", err)
			}

			ffmpegConvert(t, realWAV, refMP3, true, q, 0)

			wavProbe := ffprobeAudio(t, realWAV)
			ourProbe := ffprobeAudio(t, ourMP3)
			refProbe := ffprobeAudio(t, refMP3)

			t.Logf("V%.0f — Our: duration=%.2fs bitrate=%d, Ref: duration=%.2fs bitrate=%d",
				q, ourProbe.Duration, ourProbe.Bitrate, refProbe.Duration, refProbe.Bitrate)

			// Duration must match WAV
			assertDurationClose(t, "ours vs WAV", ourProbe.Duration, wavProbe.Duration, 0.5)
			assertDurationClose(t, "ours vs ref", ourProbe.Duration, refProbe.Duration, 0.5)

			// VBR bitrate within 30% of reference (VBR varies more)
			assertBitrateClose(t, "VBR bitrate", ourProbe.Bitrate, refProbe.Bitrate, 30)
		})
	}
}

func TestReferenceTags(t *testing.T) {
	skipIfNoFixture(t)
	if _, err := os.Stat(coverPNG); os.IsNotExist(err) {
		t.Skip("fixture missing: testdata/cover.png")
	}

	tmpDir := t.TempDir()
	mp3Path := filepath.Join(tmpDir, "tagged.mp3")

	opts := config.DefaultConvertOptions()
	opts.InputPath = realWAV
	opts.OutputPath = mp3Path
	opts.Quiet = true
	opts.Tags = config.ID3Tags{
		Title:  "Не спать!",
		Artist: "svk",
		Album:  "Мы вместе",
		Year:   "2026",
		Genre:  "Alternative",
		Track:  "3",
		Cover:  coverPNG,
	}

	_, err := Convert(context.Background(), opts)
	if err != nil {
		t.Fatalf("conversion with tags failed: %v", err)
	}

	// Read back ID3 tags
	tag, err := id3v2.Open(mp3Path, id3v2.Options{Parse: true})
	if err != nil {
		t.Fatalf("failed to open MP3 for reading tags: %v", err)
	}
	defer tag.Close()

	checks := []struct {
		name string
		got  string
		want string
	}{
		{"title", tag.Title(), "Не спать!"},
		{"artist", tag.Artist(), "svk"},
		{"album", tag.Album(), "Мы вместе"},
		{"year", tag.Year(), "2026"},
		{"genre", tag.Genre(), "Alternative"},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("tag %s: got %q, want %q", c.name, c.got, c.want)
		}
	}

	// Check track number
	trackFrames := tag.GetFrames(tag.CommonID("Track number/Position in set"))
	if len(trackFrames) == 0 {
		t.Error("track number tag missing")
	} else {
		tf, ok := trackFrames[0].(id3v2.TextFrame)
		if !ok {
			t.Error("track frame is not TextFrame")
		} else if tf.Text != "3" {
			t.Errorf("track: got %q, want %q", tf.Text, "3")
		}
	}

	// Check cover art exists
	pics := tag.GetFrames(tag.CommonID("Attached picture"))
	if len(pics) == 0 {
		t.Error("cover art missing")
	} else {
		pf, ok := pics[0].(id3v2.PictureFrame)
		if !ok {
			t.Error("picture frame is not PictureFrame")
		} else {
			if pf.MimeType != "image/png" {
				t.Errorf("cover MIME: got %q, want %q", pf.MimeType, "image/png")
			}
			if len(pf.Picture) == 0 {
				t.Error("cover data is empty")
			}
			t.Logf("cover: %s, %d bytes", pf.MimeType, len(pf.Picture))
		}
	}

	// Verify duration is still correct after tagging
	probe := ffprobeAudio(t, mp3Path)
	wavProbe := ffprobeAudio(t, realWAV)
	assertDurationClose(t, "tagged duration vs WAV", probe.Duration, wavProbe.Duration, 0.5)
	t.Logf("tagged MP3: duration=%.2fs bitrate=%d kbps", probe.Duration, probe.Bitrate)
}
