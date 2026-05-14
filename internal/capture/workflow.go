package capture

import (
	"Evidence-Capture/internal/types"
	"fmt"
)

// ExecuteWorkflow は Source から画像を取得し、Destination に保存する一連の流れを実行します。
// この設計により、取得元や保存先を自由に入れ替えることが可能になります。
func ExecuteWorkflow(source types.ImageSource, destination types.ImageDestination, path string) error {
	img, err := source.Fetch()
	if err != nil {
		return fmt.Errorf("ソースからの画像取得に失敗しました: %w", err)
	}

	if err := destination.Store(img, path); err != nil {
		return fmt.Errorf("宛先への保存に失敗しました: %w", err)
	}

	return nil
}

/*
使用例：

// 1. 画面キャプチャして直接ローカルファイルに保存する場合
func ExampleCaptureToFile() {
    source := &capture.Win32CaptureSource{}
    dest := &capture.FileDestination{Format: capture.ImgFormatJPEG}
    capture.ExecuteWorkflow(source, dest, "C:\\Temp\\evidence.jpg")
}

// 2. 画面キャプチャして Excel に貼り付ける場合
func ExampleCaptureToExcel(wrapper *excel.ExcelWrapper) {
    source := &capture.Win32CaptureSource{}
    dest := &excel.ExcelDestination{
        Wrapper:    wrapper,
        TargetCell: "B5",
        Scale:      0.8,
    }
    capture.ExecuteWorkflow(source, dest, "")
}

// 3. クリップボードにある画像をローカルファイルに保存する場合（既存の動作に近い）
func ExampleClipboardToFile() {
    source := &clipboard.ClipboardSource{}
    dest := &capture.FileDestination{Format: capture.ImgFormatPNG}
    capture.ExecuteWorkflow(source, dest, "C:\\Temp\\from_clipboard.png")
}
*/
