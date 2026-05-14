package capture

import (
	"Evidence-Capture/internal/config"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
)

const (
	ImgFormatJPEG = 2 // 設定値：JPEG
	ImgFormatBMP  = 3 // 設定値：BMP
)

var DefaultJpegQuality = 90

// FileDestination は画像をローカルファイルとして保存する Destination 実装です。
type FileDestination struct {
	Format int
}

// Store は画像をファイルに書き出します。
func (d *FileDestination) Store(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		config.Log("ERROR", "画像ファイル作成失敗: %v", "Failed to create file: %v", err)
		return fmt.Errorf("画像ファイル作成失敗: %v", err)
	}
	defer f.Close()

	switch d.Format {
	case ImgFormatJPEG:
		return jpeg.Encode(f, img, &jpeg.Options{Quality: DefaultJpegQuality})
	case ImgFormatBMP:
		config.Log("INFO", "BMP形式は未対応のためPNGで保存します: %s", "BMP format fallback to PNG: %s", path)
		return png.Encode(f, img)
	default:
		return png.Encode(f, img)
	}
}

// SaveImageToFile は互換性のために残すヘルパー関数です。
func SaveImageToFile(img image.Image, destPath string, format int) error {
	d := &FileDestination{Format: format}
	return d.Store(img, destPath)
}
