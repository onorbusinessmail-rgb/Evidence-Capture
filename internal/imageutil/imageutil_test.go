package imageutil

import (
	"image"
	"image/color"
	"testing"

	"Evidence-Capture/internal/winapi"
)

func TestImageToBitmapHandle(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 0x10 * uint8(x), G: 0x20 * uint8(y), B: 0x30, A: 0xFF})
		}
	}

	hBitmap, err := ImageToBitmapHandle(img)
	if err != nil {
		t.Fatalf("ImageToBitmapHandle failed: %v", err)
	}
	if hBitmap == 0 {
		t.Fatal("ImageToBitmapHandle returned invalid HBITMAP")
	}

	ret, _, _ := winapi.DeleteObject.Call(hBitmap)
	if ret == 0 {
		t.Fatal("DeleteObject failed for HBITMAP")
	}
}
