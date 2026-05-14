package excel

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/types"

	"github.com/xuri/excelize/v2"
)

// ExtractImagesFromExcel は、指定されたExcelファイルから画像を抽出し、一時フォルダに保存します。
// 返り値: 抽出された情報のリスト, 一時フォルダのパス, エラー
func ExtractImagesFromExcel(excelPath string) ([]types.SheetInfo, string, error) {
	// 1. 一時フォルダを作成 (ダイアログ表示中のみ保持)
	tempDir, err := os.MkdirTemp(config.AppTempDir, "ec_images_*")
	if err != nil {
		return nil, "", fmt.Errorf("一時フォルダの作成に失敗しました: %v", err)
	}

	// 2. ファイルロック回避のため、元ファイルを別のテンポラリにコピーしてから開く
	tmpExcel, err := os.CreateTemp(config.AppTempDir, "ec_read_*.xlsx")
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("作業用ファイルの作成に失敗しました: %v", err)
	}
	tmpExcelPath := tmpExcel.Name()
	defer func() {
		tmpExcel.Close()
		os.Remove(tmpExcelPath)
	}()

	src, err := os.Open(excelPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("元ファイルを開けません: %v", err)
	}
	defer src.Close()

	if _, err = io.Copy(tmpExcel, src); err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("作業用ファイルへのコピーに失敗しました: %v", err)
	}
	tmpExcel.Close()

	// 3. excelize で展開
	f, err := excelize.OpenFile(tmpExcelPath)
	if err != nil {
		os.RemoveAll(tempDir)
		return nil, "", fmt.Errorf("Excelファイルの解析に失敗しました: %v", err)
	}
	defer f.Close()

	var result []types.SheetInfo
	for _, sheetName := range f.GetSheetList() {
		visible, err := f.GetSheetVisible(sheetName)
		if err != nil {
			visible = true
		}

		var existingImages []string
		cells, err := f.GetPictureCells(sheetName)
		if err == nil {
			for _, cell := range cells {
				pics, err := f.GetPictures(sheetName, cell)
				if err == nil {
					for i, pic := range pics {
						// 画像データをファイルとして書き出す
						ext := pic.Extension
						if ext == "" {
							ext = ".png"
						}
						fileName := fmt.Sprintf("%s_%s_%d%s", sheetName, cell, i+1, ext)
						tempPath := filepath.Join(tempDir, fileName)

						if err := os.WriteFile(tempPath, pic.File, 0644); err == nil {
							existingImages = append(existingImages, tempPath)
						}
					}
				}
			}
		}

		result = append(result, types.SheetInfo{
			Name:		sheetName,
			IsVisible:	visible,
			ImageCount:	len(existingImages),
			ExistingImages:	existingImages,
		})
	}

	config.Log("INFO", "Excelから%dシートの画像を抽出しました。一時パス: %s", "Images extracted from %d sheets. Temp path: %s", len(result), tempDir)
	return result, tempDir, nil
}

/*

// ExtractSheetImages はシート内の全画像を一時保存し、そのパスのリストを返します
func ExtractSheetImages(f *excelize.File, sheetName string) ([]string, error) {
	pics, err := f.GetPictures(sheetName)
	if err != nil {
		return nil, err
	}

	tempDir := filepath.Join(config.AppTempDir, "evidence_preview")
	os.MkdirAll(tempDir, 0755)

	var paths []string
	for i, pic := range pics {
		// セル位置を含む一時ファイル名を作成
		tempPath := filepath.Join(tempDir, sheetName+"_"+pic.Cell+"_"+string(i)+".png")
		if err := os.WriteFile(tempPath, pic.File, 0644); err == nil {
			paths = append(paths, tempPath)
		}
	}
	return paths, nil
}


*/
