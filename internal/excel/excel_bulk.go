package excel

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"Evidence-Capture/internal/config"

	"github.com/xuri/excelize/v2"
)

// SheetImages は1シート分の画像挿入データを保持します。
type SheetImages struct {
	SheetName	string
	Images		[]string
}

// BulkInsertImages は抽出されたデータに基づき Excelize を使って一括で画像を挿入します。
func BulkInsertImages(data []SheetImages) (int, error) {
	// ファイルを開く
	// すでにExcel等で開かれてロックされている場合はエラーになります
	f, err := excelize.OpenFile(config.CurrentConfig.ExcelOutputPath)
	if err != nil {
		return 0, fmt.Errorf("ファイルを開けません（すでにExcelで開かれている可能性があります）: %v", err)
	}
	defer f.Close()

	totalInserted := 0
	scale := config.CurrentConfig.ImageScale
	if scale <= 0 {
		scale = 0.8
	}

	for _, sheetData := range data {
		if len(sheetData.Images) == 0 {
			continue
		}

		// シートの存在確認・作成
		sheetIndex, _ := f.GetSheetIndex(sheetData.SheetName)
		if sheetIndex == -1 {
			f.NewSheet(sheetData.SheetName)
		}

		// テキストデータのある最終行を取得
		rows, _ := f.GetRows(sheetData.SheetName)
		lastDataRow := len(rows)

		// 挿入開始列と開始行
		startCol, initRow, _ := ParseCellAddress(config.CurrentConfig.ImageInsertStartCell)
		currentRow := lastDataRow + config.CurrentConfig.Margin
		if currentRow < initRow {
			currentRow = initRow
		}

		for _, imgPath := range sheetData.Images {
			cell := fmt.Sprintf("%s%d", startCol, currentRow)

			// 画像挿入
			err := f.AddPicture(sheetData.SheetName, cell, imgPath, &excelize.GraphicOptions{
				ScaleX:	scale,
				ScaleY:	scale,
			})
			if err != nil {
				config.Log("ERROR", "画像挿入エラー: %v", "Failed to insert picture via excelize: %v", err)
				continue
			}

			// 画像の高さを取得して、次の画像の挿入行（currentRow）を計算する
			file, err := os.Open(imgPath)
			if err == nil {
				imgConfig, _, err := image.DecodeConfig(file)
				file.Close()
				if err == nil {
					// 高さをポイントに変換し、行数に換算（1行約15ポイント/20ピクセルと仮定）
					heightPx := float64(imgConfig.Height) * scale
					rowsSpanned := int(math.Ceil(heightPx / 20.0))
					currentRow += rowsSpanned + config.CurrentConfig.Margin
				} else {
					currentRow += 20	// 取得失敗時のデフォルト加算
				}
			} else {
				currentRow += 20
			}

			totalInserted++
		}
	}

	// 上書き保存
	if err := f.Save(); err != nil {
		return totalInserted, fmt.Errorf("保存に失敗しました: %v", err)
	}

	return totalInserted, nil
}
