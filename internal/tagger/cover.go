package tagger

import (
	"fmt"
	"os"
)

// CoverData содержит данные обложки и её MIME-тип.
type CoverData struct {
	Data     []byte
	MIMEType string
}

// LoadCover загружает файл обложки и определяет MIME-тип по magic bytes.
func LoadCover(path string) (*CoverData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать обложку %s: %w", path, err)
	}
	if len(data) < 4 {
		return nil, fmt.Errorf("файл обложки слишком маленький: %s", path)
	}

	mimeType, err := detectMIME(data)
	if err != nil {
		return nil, fmt.Errorf("неподдерживаемый формат обложки %s: %w", path, err)
	}

	return &CoverData{Data: data, MIMEType: mimeType}, nil
}

// detectMIME определяет MIME-тип изображения по magic bytes.
func detectMIME(data []byte) (string, error) {
	// JPEG: FF D8 FF
	if len(data) >= 3 && data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg", nil
	}
	// PNG: 89 50 4E 47 0D 0A 1A 0A
	if len(data) >= 8 &&
		data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 &&
		data[4] == 0x0D && data[5] == 0x0A && data[6] == 0x1A && data[7] == 0x0A {
		return "image/png", nil
	}
	// GIF: 47 49 46 38
	if len(data) >= 4 && data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 {
		return "image/gif", nil
	}
	// WebP: 52 49 46 46 ?? ?? ?? ?? 57 45 42 50
	if len(data) >= 12 &&
		data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 &&
		data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50 {
		return "image/webp", nil
	}

	return "", fmt.Errorf("неизвестный формат (поддерживаются JPEG, PNG, GIF, WebP)")
}
