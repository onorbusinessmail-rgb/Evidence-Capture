// internal/excel/destination.go

package excel

import (
	"Evidence-Capture/internal/clipboard"
	"Evidence-Capture/internal/types"
	"fmt"
	"image"
)

// ExcelDestination は画像を Excel シートに貼り付ける Destination 実装です。
type ExcelDestination struct {
	Wrapper    *ExcelWrapper
	Clipboard  types.ClipboardProvider // 外部から注入されるインターフェース
	TargetCell string
	Scale      float64
}

func (d *ExcelDestination) Store(img image.Image, _ string) error {
	if d.Wrapper == nil {
		return fmt.Errorf("ExcelWrapper not initialized")
	}

	// 編集モードチェックと画像セット
	if err := d.Wrapper.CheckEditMode(); err != nil {
		return err
	}
	if err := d.Clipboard.SetImage(img); err != nil {
		return err
	}

	// 表の通り、PasteImage 一括で貼り付けとリサイズを実行
	return d.Wrapper.PasteImage(d.TargetCell, d.Scale)
}

// 具体的なクリップボード転送ロジックの実装
func (d *ExcelDestination) setClipboardImage(img image.Image) error {
	// クリップボードへの画像転送処理を共通モジュールに完全に委譲する
	if err := clipboard.SetImage(img); err != nil {
		return fmt.Errorf("クリップボードへの画像セットに失敗しました: %v", err)
	}
	return nil
}
