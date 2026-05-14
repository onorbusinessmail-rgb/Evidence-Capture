package uimonitor

import (
	"Evidence-Capture/internal/capture"
	"Evidence-Capture/internal/clipboard"
	"Evidence-Capture/internal/config"
	// "errors"
)

// CaptureActiveWindowRobust 商用レベルで安全にアクティブウィンドウをキャプチャします。
func (m *Monitor) CaptureActiveWindowRobust() error {
	// 1. Win32 APIを利用してアクティブウィンドウの画像を直接メモリに取得
	source := &capture.Win32CaptureSource{}
	img, err := source.Fetch()
	if err != nil {
		config.Log("ERROR", "ウィンドウのキャプチャに失敗しました: %v", "Failed to capture window: %v", err)
		return config.NewMultiLangError("ウィンドウのキャプチャに失敗しました", "Failed to capture window")
	}

	// 2. 取得した画像をクリップボードにセット
	err = clipboard.SetImage(img)
	if err != nil {
		config.Log("ERROR", "クリップボードへの画像格納に失敗しました: %v", "Failed to set image to clipboard: %v", err)
		return config.NewMultiLangError("クリップボードへの画像格納に失敗しました", "Failed to set image to clipboard")
	}

	return nil
}
