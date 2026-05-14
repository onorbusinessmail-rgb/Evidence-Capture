package excel

import (
	"Evidence-Capture/internal/config"
	"fmt"
	"strings"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// IsIndexSheetExists はインデックスシートが存在するかどうかを確認します。
func IsIndexSheetExists() bool {
	if config.CurrentConfig.ExcelOutputPath == "" {
		return false
	}
	w, err := NewExcelWrapper()
	if err != nil {
		return false
	}
	defer w.Release()

	if err := w.Open(config.CurrentConfig.ExcelOutputPath); err != nil {
		return false
	}

	sheetsVar, err := oleutil.GetProperty(w.Workbook, "Sheets")
	if err != nil {
		return false
	}
	sheets := w.register(sheetsVar.ToIDispatch())

	idxName := config.CurrentConfig.IndexSheetName
	if idxName == "" {
		idxName = "インデックス"
	}

	sheet := GetSheetByName(sheets, idxName)
	return sheet != nil
}

// UpdateIndexStatus はアクティブセルとインデックスシートのステータスを更新します。
func UpdateIndexStatus(statusName string, applyColor bool, hexColor string) error {
	w, err := NewExcelWrapper()
	if err != nil {
		return err
	}
	defer w.Release()

	if err := w.Open(config.CurrentConfig.ExcelOutputPath); err != nil {
		return err
	}

	// Excelが編集モード（セル入力中）の場合は中断
	if err := w.CheckEditMode(); err != nil {
		return err
	}

	// --- A. アクティブセルへのステータス入力 ---
	var activeCellAddr string
	selectionVar, err := oleutil.GetProperty(w.App, "Selection")
	if err == nil {
		selection := w.register(selectionVar.ToIDispatch())
		oleutil.PutProperty(selection, "Value", statusName)
		// ハイパーリンク用のセルアドレス(A1形式)を取得
		if addrVar, err := oleutil.GetProperty(selection, "Address", false, false); err == nil {
			activeCellAddr = addrVar.ToString()
		}
	}

	// --- B. インデックスシートとカレントシートの特定 ---
	idxName := config.CurrentConfig.IndexSheetName
	if idxName == "" {
		idxName = "インデックス"
	}

	sheetsVar, _ := oleutil.GetProperty(w.Workbook, "Sheets")
	sheets := w.register(sheetsVar.ToIDispatch())
	idxSheet := w.register(GetSheetByName(sheets, idxName))

	if idxSheet == nil {
		return fmt.Errorf("インデックスシート「%s」が見つかりません", idxName)
	}

	// 現在操作中のシート名を取得
	activeSheetVar, _ := oleutil.GetProperty(w.App, "ActiveSheet")
	activeSheet := w.register(activeSheetVar.ToIDispatch())
	activeSheetNameVar, _ := oleutil.GetProperty(activeSheet, "Name")
	activeSheetName := activeSheetNameVar.ToString()

	// インデックスシート自体を操作している場合は更新不要
	if activeSheetName == idxName {
		return nil
	}

	// --- C. 挿入・更新位置の特定 (Findメソッドによる高速化) ---
	var finalRow int
	targetRow, err := FindSheetRowByName(idxSheet, activeSheetName)
	if err != nil {
		config.Log("ERROR", "インデックス検索中にエラー: %v", "Index search failed: %v", err)
	}

	if targetRow > 0 {
		// 1. 既存行が見つかった場合：その行の下に新しいステータス行を挿入
		finalRow = targetRow + 1
		rowsVar, _ := oleutil.GetProperty(idxSheet, "Rows", finalRow)
		rows := w.register(rowsVar.ToIDispatch())
		oleutil.CallMethod(rows, "Insert")

		// 挿入された行の背景色をリセット
		if interiorVar, err := oleutil.GetProperty(rows, "Interior"); err == nil {
			interior := w.register(interiorVar.ToIDispatch())
			oleutil.PutProperty(interior, "ColorIndex", -4142)	// xlNone
		}
	} else {
		// 2. 見つからなかった場合：末尾の新しい行を特定
		finalRow, _ = GetNextIndexRow(idxSheet)
	}

	// --- D. データの書き込みと書式設定 ---

	// 行の高さ設定（設定値がある場合）
	if config.CurrentConfig.DefaultRowHeight > 0 {
		if rowToSet, err := oleutil.GetProperty(idxSheet, "Rows", finalRow); err == nil {
			r := w.register(rowToSet.ToIDispatch())
			oleutil.PutProperty(r, "RowHeight", config.CurrentConfig.DefaultRowHeight*0.75)
		}
	}

	// B列: シート名
	if cellBVar, err := oleutil.GetProperty(idxSheet, "Cells", finalRow, 2); err == nil {
		cellB := w.register(cellBVar.ToIDispatch())
		oleutil.PutProperty(cellB, "Value", activeSheetName)
	}

	// C列: ステータス名と背景色
	if cellCVar, err := oleutil.GetProperty(idxSheet, "Cells", finalRow, 3); err == nil {
		cellC := w.register(cellCVar.ToIDispatch())
		oleutil.PutProperty(cellC, "Value", statusName)

		if interiorVar, err := oleutil.GetProperty(cellC, "Interior"); err == nil {
			interior := w.register(interiorVar.ToIDispatch())
			if applyColor && hexColor != "" {
				if colorInt, err := HexToExcelColor(hexColor); err == nil {
					oleutil.PutProperty(interior, "Color", colorInt)
				}
			} else {
				oleutil.PutProperty(interior, "ColorIndex", -4142)
			}
		}
	}

	// D列: エビデンスへのハイパーリンク
	if activeCellAddr != "" {
		if cellDVar, err := oleutil.GetProperty(idxSheet, "Cells", finalRow, 4); err == nil {
			cellD := w.register(cellDVar.ToIDispatch())
			if hlinksVar, err := oleutil.GetProperty(idxSheet, "Hyperlinks"); err == nil {
				hlinks := w.register(hlinksVar.ToIDispatch())
				subAddr := fmt.Sprintf("'%s'!%s", activeSheetName, activeCellAddr)
				textDisp := fmt.Sprintf("→ %s_%s", activeSheetName, activeCellAddr)
				// Hyperlinks.Add(Anchor, Address, SubAddress, ScreenTip, TextToDisplay)
				oleutil.CallMethod(hlinks, "Add", cellD, "", subAddr, "", textDisp)
			}
		}
	}

	return nil
}

// CreateIndexSheet はインデックスシートを新規作成します。
func CreateIndexSheet() error {
	w, err := NewExcelWrapper()
	if err != nil {
		return err
	}
	defer w.Release()

	if err := w.Open(config.CurrentConfig.ExcelOutputPath); err != nil {
		return err
	}

	if err := w.CheckEditMode(); err != nil {
		return err
	}

	oleutil.PutProperty(w.App, "Visible", true)
	sheetsVar, _ := GetPropertyIDispatch(w.App, "Sheets")
	sheets := w.register(sheetsVar)

	var anchorSheet *ole.IDispatch
	var isAfter bool

	tocName := GetTOCName()
	if tocSheet := GetSheetByName(sheets, tocName); tocSheet != nil {
		anchorSheet = w.register(tocSheet)
		isAfter = true
	} else {
		if firstSheet, err := GetPropertyIDispatch(sheets, "Item", 1); err == nil {
			anchorSheet = w.register(firstSheet)
			isAfter = false
		}
	}

	var newSheetVar *ole.VARIANT
	if isAfter {
		newSheetVar, err = oleutil.CallMethod(sheets, "Add", nil, anchorSheet)
	} else {
		newSheetVar, err = oleutil.CallMethod(sheets, "Add", anchorSheet)
	}

	if err != nil {
		return fmt.Errorf("シートの追加に失敗しました: %v", err)
	}
	newSheet := w.register(newSheetVar.ToIDispatch())

	idxName := config.CurrentConfig.IndexSheetName
	if idxName == "" {
		idxName = "インデックス"
	}
	oleutil.PutProperty(newSheet, "Name", idxName)

	if cellsVar, err := GetPropertyIDispatch(newSheet, "Cells"); err == nil {
		cells := w.register(cellsVar)
		if config.CurrentConfig.DefaultRowHeight > 0 {
			oleutil.PutProperty(cells, "RowHeight", config.CurrentConfig.DefaultRowHeight*0.75)
		}
		if config.CurrentConfig.DefaultColWidth > 0 {
			colWidth := (config.CurrentConfig.DefaultColWidth - 5.0) / 7.0
			if colWidth < 0.1 {
				colWidth = 0.1
			}
			oleutil.PutProperty(cells, "ColumnWidth", colWidth)
		}
	}

	headers := []string{"No.", "シート名", "ステータス", "リンク先", "備考"}
	for i, h := range headers {
		if cellVar, err := oleutil.GetProperty(newSheet, "Cells", 3, i+1); err == nil {
			cell := w.register(cellVar.ToIDispatch())
			oleutil.PutProperty(cell, "Value", h)
			oleutil.PutProperty(cell, "HorizontalAlignment", -4108)

			if interiorVar, err := GetPropertyIDispatch(cell, "Interior"); err == nil {
				interior := w.register(interiorVar)
				oleutil.PutProperty(interior, "Color", 13421823)
			}

			if fontVar, err := GetPropertyIDispatch(cell, "Font"); err == nil {
				font := w.register(fontVar)
				oleutil.PutProperty(font, "Bold", true)
			}
		}
	}

	for i := 1; i <= 100; i++ {
		rowAddr := 3 + i
		if cellVar, err := oleutil.GetProperty(newSheet, "Cells", rowAddr, 1); err == nil {
			cell := w.register(cellVar.ToIDispatch())
			oleutil.PutProperty(cell, "Value", i)
			oleutil.PutProperty(cell, "HorizontalAlignment", -4108)
		}
	}

	if rngVar, err := oleutil.GetProperty(newSheet, "Range", "A3:E103"); err == nil {
		rng := w.register(rngVar.ToIDispatch())
		if bordersVar, err := GetPropertyIDispatch(rng, "Borders"); err == nil {
			borders := w.register(bordersVar)
			oleutil.PutProperty(borders, "LineStyle", 1)
		}
	}

	if rngHeaderVar, err := oleutil.GetProperty(newSheet, "Range", "A3:E3"); err == nil {
		rngHeader := w.register(rngHeaderVar.ToIDispatch())
		oleutil.CallMethod(rngHeader, "AutoFilter")
	}

	oleutil.CallMethod(newSheet, "Activate")
	if activeWinVar, err := oleutil.GetProperty(w.App, "ActiveWindow"); err == nil {
		activeWin := w.register(activeWinVar.ToIDispatch())
		if config.CurrentConfig.HideGridlines {
			oleutil.PutProperty(activeWin, "DisplayGridlines", false)
		}
		if config.CurrentConfig.ZoomPercent > 0 {
			oleutil.PutProperty(activeWin, "Zoom", config.CurrentConfig.ZoomPercent)
		}

		if cellA4Var, err := GetPropertyIDispatch(newSheet, "Range", "A4"); err == nil {
			cellA4 := w.register(cellA4Var)
			oleutil.CallMethod(cellA4, "Select")
			oleutil.PutProperty(activeWin, "FreezePanes", true)
		}
	}

	w.BringToForeground()

	if rngVar, err := oleutil.GetProperty(newSheet, "Range", "C4:C104"); err == nil {
		rng := w.register(rngVar.ToIDispatch())
		oleutil.PutProperty(rng, "HorizontalAlignment", -4108)
	}

	return nil
}

// CopyTOCNamesToIndex は目次の内容をインデックスシートにコピーします。
func CopyTOCNamesToIndex(excelApp *ole.IDispatch, indexSheet *ole.IDispatch) {
	tocName := GetTOCName()
	sheetsVar, err := oleutil.GetProperty(excelApp, "Sheets")
	if err != nil {
		return
	}
	sheets := sheetsVar.ToIDispatch()
	defer sheets.Release()

	tocSheet := GetSheetByName(sheets, tocName)
	if tocSheet == nil {
		return
	}
	defer tocSheet.Release()

	row := 4
	idxRow := 4
	for {
		cellCVar, err := oleutil.GetProperty(tocSheet, "Cells", row, 3)
		if err != nil {
			break
		}
		cellC := cellCVar.ToIDispatch()
		valVar, _ := oleutil.GetProperty(cellC, "Value")
		valStr := strings.TrimSpace(fmt.Sprint(valVar.Value()))
		cellC.Release()

		if valStr == "" || valStr == "<nil>" || valStr == "0" {
			break
		}

		if cellBVar, err := oleutil.GetProperty(indexSheet, "Cells", idxRow, 2); err == nil {
			cellB := cellBVar.ToIDispatch()
			oleutil.PutProperty(cellB, "Value", valStr)
			cellB.Release()
		}

		row++
		idxRow++
		if row > 500 {
			break
		}
	}
}
