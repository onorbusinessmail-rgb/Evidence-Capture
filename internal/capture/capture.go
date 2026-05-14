package capture

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/winapi"
	"image"
	"unsafe"
)

// Win32CaptureSource は Win32 API を使用して画面キャプチャを行う Source 実装です。
type Win32CaptureSource struct{}

// Fetch はアクティブウィンドウをキャプチャして image.Image を返します。
func (s *Win32CaptureSource) Fetch() (image.Image, error) {
	// A. DPI Awareness の設定 (変更なし)
	if winapi.SetProcessDpiAwarenessContext.Find() == nil {
		winapi.User32.NewProc("SetProcessDPIAware").Call()
	} else {
		winapi.SetProcessDpiAwarenessContext.Call(winapi.DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2)
	}

	// 1. アクティブウィンドウのハンドルを取得 (変更なし)
	hwnd, _, _ := winapi.GetForegroundWindow.Call()
	if hwnd == 0 {
		return nil, config.NewMultiLangError("アクティブなウィンドウが見つかりません", "Active window not found")
	}

	// 2. ウィンドウの矩形領域（座標）を「影なし」で取得
	var rect winapi.RECT

	// 影を除いた正確な表示領域を取得するためのAPI呼び出し
	// ProcDwmGetWindowAttr は internal/winapi 等で事前に定義しておく必要があります
	ret, _, _ := winapi.ProcDwmGetWindowAttr.Call(
		hwnd,
		uintptr(winapi.DWMWA_EXTENDED_FRAME_BOUNDS),
		uintptr(unsafe.Pointer(&rect)),
		uintptr(unsafe.Sizeof(rect)),
	)

	// DWMでの取得に失敗した場合は、従来の GetWindowRect を使う（保険）
	if ret != 0 {
		winapi.GetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rect)))
	}

	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top

	// 最小化・無効なサイズの検知
	if width <= 0 || height <= 0 {
		return nil, config.NewMultiLangError("ウィンドウのサイズが不正なためキャプチャをスキップしました", "Skipped capture because window size is invalid")
	}
	// 画面外のチェック（BitBlt時のみ適用）
	if !config.CurrentConfig.EnableFullWindowCapture && (rect.Left <= -10000 || rect.Top <= -10000) {
		return nil, config.NewMultiLangError("ウィンドウが画面外にあるためキャプチャをスキップしました", "Skipped capture because window is out of screen")
	}

	// 3. 以降の BitBlt 処理は、取得した rect.Left / rect.Top を使うのでそのままでOK
	hdcSrc, _, _ := winapi.GetDC.Call(0)
	if hdcSrc == 0 {
		return nil, config.NewMultiLangError("デバイスコンテキストの取得に失敗しました", "Failed to get device context")
	}
	defer winapi.ReleaseDC.Call(0, hdcSrc)

	hdcDest, _, _ := winapi.CreateCompatibleDC.Call(hdcSrc)
	if hdcDest == 0 {
		return nil, config.NewMultiLangError("互換DCの作成に失敗しました", "Failed to create compatible DC")
	}
	defer winapi.DeleteDC.Call(hdcDest)

	hBitmap, _, _ := winapi.CreateCompatibleBitmap.Call(hdcSrc, uintptr(width), uintptr(height))
	if hBitmap == 0 {
		return nil, config.NewMultiLangError("互換ビットマップの作成に失敗しました", "Failed to create compatible bitmap")
	}
	defer winapi.DeleteObject.Call(hBitmap)

	oldObj, _, _ := winapi.SelectObject.Call(hdcDest, hBitmap)
	if oldObj == 0 || oldObj == winapi.HGDI_ERROR {
		return nil, config.NewMultiLangError("互換DCへのビットマップ選択に失敗しました (Fetch)", "Failed to select bitmap to compatible DC (Fetch)")
	}

	// 4. キャプチャ実行の分岐
	// 「撮影範囲：デフォルト撮影範囲」が 1 (アクティブウィンドウ) かつ 完全キャプチャが有効な場合のみ PrintWindow を使用
	if config.CurrentConfig.DefaultCaptureRange == 1 && config.CurrentConfig.EnableFullWindowCapture {
		// 完全キャプチャ (PrintWindow)
		// 第3引数の 2 はモダンアプリ向けの PW_RENDERFULLCONTENT フラグです
		ret, _, _ := winapi.PrintWindow.Call(uintptr(hwnd), uintptr(hdcDest), 2)
		if ret == 0 {
			config.Log("WARN", "PrintWindowによる取得に失敗しました。ターゲットが管理者権限のウィンドウである可能性があります。", "PrintWindow failed. Target might be an admin window.")
		}
	} else {
		// 見えている範囲のみ取得 (BitBlt)
		// 影なし座標 (rect.Left, rect.Top) からコピーを開始するため、余白が消える
		winapi.BitBlt.Call(hdcDest, 0, 0, uintptr(width), uintptr(height), hdcSrc, uintptr(rect.Left), uintptr(rect.Top), uintptr(winapi.SRCCOPY))
	}

	// BitmapHandleToImage(GetDIBits) を呼ぶ前に、DCからビットマップを解放する必要がある
	winapi.SelectObject.Call(hdcDest, oldObj)

	return BitmapHandleToImage(hBitmap)
}

// BitmapHandleToImage は Windows の HBITMAP ハンドルを Go の image.Image (*image.NRGBA) に変換します。
func BitmapHandleToImage(hBitmap uintptr) (*image.NRGBA, error) {
	// BITMAP 構造体でビットマップのサイズを取得
	type BITMAP struct {
		Type       int32
		Width      int32
		Height     int32
		WidthBytes int32
		Planes     uint16
		BitsPixel  uint16
		Bits       uintptr
	}
	var bm BITMAP
	winapi.GetObject.Call(hBitmap, uintptr(unsafe.Sizeof(bm)), uintptr(unsafe.Pointer(&bm)))
	if bm.Width == 0 || bm.Height == 0 {
		return nil, config.NewMultiLangError("ビットマップのサイズ取得に失敗しました", "Failed to get bitmap size")
	}

	// BITMAPINFOHEADER: GetDIBits に渡すピクセルフォーマット指定
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
	bih := BITMAPINFOHEADER{
		BiSize:     uint32(unsafe.Sizeof(BITMAPINFOHEADER{})),
		BiWidth:    bm.Width,
		BiHeight:   -bm.Height, // 負 = トップダウン（上から順）
		BiPlanes:   1,
		BiBitCount: 32, // BGRA 各8bit
	}

	pixelSize := int(bm.Width) * int(bm.Height) * 4
	pixels := make([]byte, pixelSize)

	// メモリDC作成 → ピクセルデータ取得
	// (GetDIBitsに渡すビットマップはDCに選択されていてはならないため、SelectObjectは不要)
	hdc, _, _ := winapi.CreateCompatibleDC.Call(0)
	if hdc == 0 {
		return nil, config.NewMultiLangError("メモリDC作成に失敗しました", "Failed to create memory DC")
	}
	defer winapi.DeleteDC.Call(hdc)

	ret, _, _ := winapi.GetDIBits.Call(
		hdc, hBitmap,
		0, uintptr(bm.Height),
		uintptr(unsafe.Pointer(&pixels[0])),
		uintptr(unsafe.Pointer(&bih)),
		winapi.DIB_RGB_COLORS,
	)
	if ret == 0 {
		return nil, config.NewMultiLangError("ピクセルデータ取得に失敗しました", "Failed to get pixel data")
	}

	// BGRA -> RGBA 変換
	img := image.NewNRGBA(image.Rect(0, 0, int(bm.Width), int(bm.Height)))
	for i := 0; i < len(pixels); i += 4 {
		img.Pix[i+0] = pixels[i+2] // R
		img.Pix[i+1] = pixels[i+1] // G
		img.Pix[i+2] = pixels[i+0] // B
		img.Pix[i+3] = 255         // A
	}
	return img, nil
}
