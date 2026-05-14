package clipboard

import (
	"Evidence-Capture/internal/capture"
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/imageutil"
	"Evidence-Capture/internal/winapi"
	"fmt"
	"image"
	"time"
)

// ClipboardSource はクリップボードから画像を取得する Source 実装です。
type ClipboardSource struct{}

// Fetch はクリップボードから画像を抽出して image.Image を返します。
func (s *ClipboardSource) Fetch() (image.Image, error) {
	ret, _, _ := winapi.OpenClipboard.Call(0)
	if ret == 0 {
		return nil, config.NewMultiLangError("クリップボードを開けません", "Cannot open clipboard")
	}
	defer winapi.CloseClipboard.Call()

	hBitmap, _, _ := winapi.GetClipboardData.Call(uintptr(winapi.CF_BITMAP))
	if hBitmap == 0 {
		return nil, config.NewMultiLangError("クリップボードにビットマップが存在しません", "No bitmap in clipboard")
	}

	return capture.BitmapHandleToImage(hBitmap)
}

func SetImage(img image.Image) error {
	// 画像からHBITMAPへの変換
	hBitmap, err := imageutil.ImageToBitmapHandle(img)
	if err != nil {
		return fmt.Errorf("%v: %w", config.NewMultiLangError("ビットマップ変換エラー", "Bitmap conversion error"), err)
	}

	// 成功フラグ。SetClipboardDataに成功してOSに所有権が移った場合のみ true にする
	ownedBySystem := false
	defer func() {
		if !ownedBySystem && hBitmap != 0 {
			winapi.DeleteObject.Call(hBitmap)
		}
	}()

	ret, _, _ := winapi.OpenClipboard.Call(0)
	if ret == 0 {
		return config.NewMultiLangError("クリップボードを開けません", "Cannot open clipboard")
	}
	defer winapi.CloseClipboard.Call()

	winapi.EmptyClipboard.Call()
	res, _, _ := winapi.SetClipboardData.Call(uintptr(winapi.CF_BITMAP), hBitmap)
	if res == 0 {
		return config.NewMultiLangError("SetClipboardData失敗", "SetClipboardData failed")
	}

	// システムに所有権が移ったため、以降 DeleteObject は行わない
	ownedBySystem = true

	// Excelがクリップボードを認識するまでの猶予を与える
	time.Sleep(100 * time.Millisecond)

	return nil
}

// GetClipboardImage は互換性のために残します。
func GetClipboardImage() (image.Image, error) {
	s := &ClipboardSource{}
	return s.Fetch()
}
