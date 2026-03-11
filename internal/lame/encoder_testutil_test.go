package lame

import "math"

// generateSineWave produces int16 LE interleaved PCM data.
// Returns []byte of length: durationMs/1000 * sampleRate * channels * 2
func generateSineWave(sampleRate, channels, durationMs int) []byte {
	// Generate a 440Hz sine wave
	totalSamples := sampleRate * durationMs / 1000
	buf := make([]byte, totalSamples*channels*2)
	for i := range totalSamples {
		sample := int16(math.Sin(2*math.Pi*440*float64(i)/float64(sampleRate)) * 32000)
		for ch := range channels {
			offset := (i*channels + ch) * 2
			buf[offset] = byte(sample)
			buf[offset+1] = byte(sample >> 8)
		}
	}
	return buf
}
