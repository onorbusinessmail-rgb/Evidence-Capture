// internal/imageutil/imageutil.go

package imageutil

import (
	"Evidence-Capture/internal/winapi"
	"fmt"
	"image"
	"unsafe"
)

// ImageToBitmapHandle は Go の image.Image を Windows の HBITMAP に変換します。
func ImageToBitmapHandle(img image.Image) (uintptr, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// BITMAPINFOHEADER の定義
	type BITMAPINFOHEADER struct {
		BiSize          uint32
		BiWidth         int32
		BiHeight        int32
		BiPlanes        uint16
		BiBitCount      uint16
		BiCompression   uint32
		BiSizeImage     uint32
		BiXPelsPerMeter int32
		BiYPelsPerMeter int32
		BiClrUsed       uint32
		BiClrImportant  uint32
	}

	type BITMAPINFO struct {
		Header BITMAPINFOHEADER
		Colors [1]uint32
	}

	bi := BITMAPINFO{}
	bi.Header.BiSize = uint32(unsafe.Sizeof(bi.Header))
	bi.Header.BiWidth = int32(width)
	bi.Header.BiHeight = int32(-height) // 負の値でトップダウン（上から下へ描画）を指定
	bi.Header.BiPlanes = 1
	bi.Header.BiBitCount = 32
	bi.Header.BiCompression = 0 // BI_RGB

	var bits uintptr
	// 1. メモリ領域（空のキャンバス）を確保
	hBitmap, _, _ := winapi.CreateDIBSection.Call(
		0,
		uintptr(unsafe.Pointer(&bi)),
		winapi.DIB_RGB_COLORS,
		uintptr(unsafe.Pointer(&bits)),
		0,
		0,
	)

	if hBitmap == 0 {
		return 0, fmt.Errorf("CreateDIBSection に失敗しました")
	}

	// 2. 欠落していたピクセルデータのコピー処理を実行
	// bits ポインタを Go のスライスとして扱い、安全かつ高速にメモリ転送を行う
	pixelBytes := width * height * 4
	dest := unsafe.Slice((*byte)(unsafe.Pointer(bits)), pixelBytes)

	// image.Image の型によって効率的にピクセルを取り出してBGRA配列に詰める
	switch srcImg := img.(type) {
	case *image.NRGBA:
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				srcIdx := (y * srcImg.Stride) + (x * 4)
				dstIdx := (y * width * 4) + (x * 4)
				dest[dstIdx+0] = srcImg.Pix[srcIdx+2] // B (Windowsは青が先)
				dest[dstIdx+1] = srcImg.Pix[srcIdx+1] // G
				dest[dstIdx+2] = srcImg.Pix[srcIdx+0] // R
				dest[dstIdx+3] = 255                  // A (クリップボード用は不透明指定)
			}
		}
	default:
		// その他の画像形式に対する汎用フォールバック
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				r, g, b, _ := img.At(x, y).RGBA()
				dstIdx := (y * width * 4) + (x * 4)
				dest[dstIdx+0] = byte(b >> 8) // B
				dest[dstIdx+1] = byte(g >> 8) // G
				dest[dstIdx+2] = byte(r >> 8) // R
				dest[dstIdx+3] = 255          // A
			}
		}
	}

	return hBitmap, nil
}
