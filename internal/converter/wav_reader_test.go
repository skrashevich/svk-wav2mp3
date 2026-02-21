package converter

import "testing"

func TestNormalizePCM_8bit(t *testing.T) {
	// 8-bit: value 128 (silence) → 0
	samples := []int{128, 0, 255}
	result := normalizePCM(samples, 8)

	if result[0] != 0 {
		t.Errorf("128 (8-bit) should normalize to 0, got %d", result[0])
	}
	// 0 → (0 - 128) << 8 = -128 << 8 = -32768
	if result[1] != -32768 {
		t.Errorf("0 (8-bit) should normalize to -32768, got %d", result[1])
	}
	// 255 → (255 - 128) << 8 = 127 << 8 = 32512
	if result[2] != 32512 {
		t.Errorf("255 (8-bit) should normalize to 32512, got %d", result[2])
	}
}

func TestNormalizePCM_16bit(t *testing.T) {
	samples := []int{0, 32767, -32768, 1000}
	result := normalizePCM(samples, 16)

	for i, s := range samples {
		if result[i] != int16(s) {
			t.Errorf("16-bit sample[%d]: expected %d, got %d", i, int16(s), result[i])
		}
	}
}

func TestNormalizePCM_24bit(t *testing.T) {
	// 24-bit: >> 8
	samples := []int{8388607, -8388608, 0, 256}
	result := normalizePCM(samples, 24)

	expected := []int16{
		int16(8388607 >> 8),
		int16(-8388608 >> 8),
		0,
		int16(256 >> 8),
	}
	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("24-bit sample[%d]: expected %d, got %d", i, exp, result[i])
		}
	}
}

func TestNormalizePCM_32bit(t *testing.T) {
	// 32-bit: >> 16
	samples := []int{2147483647, -2147483648, 0, 65536}
	result := normalizePCM(samples, 32)

	expected := []int16{
		int16(2147483647 >> 16),
		int16(-2147483648 >> 16),
		0,
		int16(65536 >> 16),
	}
	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("32-bit sample[%d]: expected %d, got %d", i, exp, result[i])
		}
	}
}

func TestNormalizePCM_EmptyInput(t *testing.T) {
	result := normalizePCM(nil, 16)
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d elements", len(result))
	}
}
