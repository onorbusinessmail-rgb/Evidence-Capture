// Package excel の TOC（目次）シート操作を提供します。
package excel

import (
	"fmt"
	"os"
	"strings"

	"Evidence-Capture/internal/config"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// EnsureTOCSheet は目次シートの存在とデータ有無を確認します。
func EnsureTOCSheet() (exists bool, hasData bool, err error) {
	path := config.CurrentConfig.ExcelOutputPath
	if path == "" || (func() bool { _, e := os.Stat(path); return os.IsNotExist(e) }()) {
		config.Log("ERROR", "Excelファイルのパスが正しくありません", "Excel file path is incorrect")
		return false, false, fmt.Errorf("Excelファイルのパスが正しくありません")
	}

	w, err := NewExcelWrapper()
	if err != nil {
		return false, false, err
	}
	defer w.Release()

	if err := w.Open(path); err != nil {
		return false, false, err
	}

	sheetsVar, _ := oleutil.GetProperty(w.App, "Sheets")
	sheets := w.register(sheetsVar.ToIDispatch())

	tocSheet := GetSheetByName(sheets, GetTOCName())
	if tocSheet == nil {
		return false, false, nil
	}
	w.register(tocSheet)

	exists = true
	cellVar, err := oleutil.GetProperty(tocSheet, "Range", "C4")
	if err == nil {
		cell := w.register(cellVar.ToIDispatch())
		valVar, _ := oleutil.GetProperty(cell, "Value")
		val := valVar.ToString()
		hasData = (val != "" && val != "<nil>")
	}
	return exists, hasData, nil
}

// TOC_CreateTOCSheet は目次シートを新規作成します。
func TOC_CreateTOCSheet() error {
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

	firstSheetVar, _ := GetPropertyIDispatch(sheets, "Item", 1)
	firstSheet := w.register(firstSheetVar)

	newSheetVar, err := oleutil.CallMethod(sheets, "Add", firstSheet)
	if err != nil {
		return fmt.Errorf("シート作成失敗: %v", err)
	}
	newSheet := w.register(newSheetVar.ToIDispatch())

	tocName := GetTOCName()
	oleutil.PutProperty(newSheet, "Name", tocName)

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

	setupTOCHeader(w, newSheet)

	samples := []string{"sample1", "sample2", "sample3"}
	for i, s := range samples {
		if cellVar, err := oleutil.GetProperty(newSheet, "Range", fmt.Sprintf("C%d", 4+i)); err == nil {
			cell := w.register(cellVar.ToIDispatch())
			oleutil.PutProperty(cell, "Value", s)
		}
		if cellVar, err := oleutil.GetProperty(newSheet, "Range", fmt.Sprintf("A%d", 4+i)); err == nil {
			cell := w.register(cellVar.ToIDispatch())
			oleutil.PutProperty(cell, "Value", i+1)
		}
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

	if colsVar, err := oleutil.GetProperty(newSheet, "Columns", "A:H"); err == nil {
		cols := w.register(colsVar.ToIDispatch())
		oleutil.CallMethod(cols, "AutoFit")
	}

	w.BringToForeground()
	return nil
}

func setupTOCHeader(w *ExcelWrapper, ws *ole.IDispatch) {
	headers := []string{"No", "カテゴリ", "シート名", "リネーム案", "リンク", "実施日", "担当者", "判定"}
	for i, h := range headers {
		cellVar, _ := oleutil.GetProperty(ws, "Cells", 3, i+1)
		cell := w.register(cellVar.ToIDispatch())
		oleutil.PutProperty(cell, "Value", h)
		oleutil.PutProperty(cell, "HorizontalAlignment", -4108)
		interiorVar, _ := GetPropertyIDispatch(cell, "Interior")
		interior := w.register(interiorVar)
		oleutil.PutProperty(interior, "Color", 13434879)
		fontVar, _ := GetPropertyIDispatch(cell, "Font")
		font := w.register(fontVar)
		oleutil.PutProperty(font, "Bold", true)
	}
	rngVar, _ := GetPropertyIDispatch(ws, "Range", "A3:H100")
	rng := w.register(rngVar)
	bordersVar, _ := GetPropertyIDispatch(rng, "Borders")
	borders := w.register(bordersVar)
	oleutil.PutProperty(borders, "LineStyle", 1)
	colsVar, _ := GetPropertyIDispatch(ws, "Columns", "A:H")
	cols := w.register(colsVar)
	oleutil.CallMethod(cols, "AutoFit")
}

// TOC_SyncSheetsFromList は目次リストに基づいてシートを同期します。
func TOC_SyncSheetsFromList() error {
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

	sheetsVar, _ := oleutil.GetProperty(w.App, "Sheets")
	sheets := w.register(sheetsVar.ToIDispatch())

	existingSheetsMap := make(map[string]bool)
	var allSheetNames []string

	countVar, _ := oleutil.GetProperty(sheets, "Count")
	sCount := int(countVar.Val)

	for i := 1; i <= sCount; i++ {
		itemVar, err := oleutil.GetProperty(sheets, "Item", i)
		if err == nil {
			s := itemVar.ToIDispatch()
			nVar, _ := oleutil.GetProperty(s, "Name")
			name := nVar.ToString()
			existingSheetsMap[name] = true
			allSheetNames = append(allSheetNames, name)
			s.Release()
		}
	}

	tocName := GetTOCName()
	tocSheet := GetSheetByName(sheets, tocName)
	if tocSheet == nil {
		return fmt.Errorf("目次シートが見つかりません")
	}
	w.register(tocSheet)

	row := 4
	listedSheets := make(map[string]bool)
	for {
		cellVar, err := oleutil.GetProperty(tocSheet, "Range", fmt.Sprintf("C%d", row))
		if err != nil {
			break
		}
		cell := w.register(cellVar.ToIDispatch())
		valVar, _ := oleutil.GetProperty(cell, "Value")
		valStr := strings.TrimSpace(fmt.Sprint(valVar.Value()))

		if valStr == "" || valStr == "<nil>" || valStr == "0" {
			break
		}

		if valStr != tocName {
			listedSheets[valStr] = true
			if !existingSheetsMap[valStr] {
				cVar, _ := oleutil.GetProperty(sheets, "Count")
				lastSheetVar, _ := oleutil.GetProperty(sheets, "Item", cVar.Value())
				lastSheet := w.register(lastSheetVar.ToIDispatch())
				newWsVar, err := oleutil.CallMethod(sheets, "Add", nil, lastSheet)
				if err == nil {
					newWs := w.register(newWsVar.ToIDispatch())
					oleutil.PutProperty(newWs, "Name", valStr)
					oleutil.CallMethod(newWs, "Activate")
					if cellsVar, err := GetPropertyIDispatch(newWs, "Cells"); err == nil {
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
					if activeWinVar, err := oleutil.GetProperty(w.App, "ActiveWindow"); err == nil {
						activeWin := w.register(activeWinVar.ToIDispatch())
						if config.CurrentConfig.HideGridlines {
							oleutil.PutProperty(activeWin, "DisplayGridlines", false)
						}
						if config.CurrentConfig.ZoomPercent > 0 {
							oleutil.PutProperty(activeWin, "Zoom", config.CurrentConfig.ZoomPercent)
						}
						if config.CurrentConfig.EnableWindowFreeze && config.CurrentConfig.FreezePaneCell != "" {
							if cellRangeVar, err := GetPropertyIDispatch(newWs, "Range", config.CurrentConfig.FreezePaneCell); err == nil {
								cellRange := w.register(cellRangeVar)
								oleutil.CallMethod(cellRange, "Select")
								oleutil.PutProperty(activeWin, "FreezePanes", true)
							}
						}
					}
					existingSheetsMap[valStr] = true
					SetupBidirectionalLinks(w, newWs, tocSheet, row, valStr, tocName)
				}
			} else {
				if targetWs := GetSheetByName(sheets, valStr); targetWs != nil {
					w.register(targetWs)
					SetupBidirectionalLinks(w, targetWs, tocSheet, row, valStr, tocName)
				}
			}
		}
		row++
	}

	lastRow := row
	for _, sName := range allSheetNames {
		if sName == tocName || listedSheets[sName] {
			continue
		}
		cellVar, _ := oleutil.GetProperty(tocSheet, "Range", fmt.Sprintf("C%d", lastRow))
		cell := w.register(cellVar.ToIDispatch())
		oleutil.PutProperty(cell, "Value", sName)
		if targetWs := GetSheetByName(sheets, sName); targetWs != nil {
			w.register(targetWs)
			SetupBidirectionalLinks(w, targetWs, tocSheet, lastRow, sName, tocName)
		}
		lastRow++
	}

	oleutil.CallMethod(tocSheet, "Activate")
	w.BringToForeground()
	return nil
}

// TOC_SortSheets は目次リストの順番に従ってシートを並び替えます。
func TOC_SortSheets() error {
	w, err := NewExcelWrapper()
	if err != nil {
		return err
	}
	defer w.Release()

	if err := w.Open(config.CurrentConfig.ExcelOutputPath); err != nil {
		return err
	}

	oleutil.PutProperty(w.App, "ScreenUpdating", false)
	defer oleutil.PutProperty(w.App, "ScreenUpdating", true)

	sheetsVar, _ := oleutil.GetProperty(w.App, "Sheets")
	sheets := w.register(sheetsVar.ToIDispatch())

	tocName := GetTOCName()
	tocSheet := GetSheetByName(sheets, tocName)
	if tocSheet == nil {
		return fmt.Errorf("目次シートが見つかりません")
	}
	w.register(tocSheet)

	var sheetNames []string
	row := 4
	for {
		cellVar, err := oleutil.GetProperty(tocSheet, "Range", fmt.Sprintf("C%d", row))
		if err != nil {
			break
		}
		cell := w.register(cellVar.ToIDispatch())
		valVar, err := oleutil.GetProperty(cell, "Value")
		valStr := ""
		if err == nil && valVar != nil {
			valStr = strings.TrimSpace(fmt.Sprint(valVar.Value()))
		}
		if valStr == "" || valStr == "<nil>" || valStr == "0" {
			break
		}
		if valStr != tocName {
			sheetNames = append(sheetNames, valStr)
		}
		row++
	}

	firstSheetVar, err := oleutil.GetProperty(sheets, "Item", 1)
	if err == nil && firstSheetVar != nil {
		firstSheet := w.register(firstSheetVar.ToIDispatch())
		oleutil.CallMethod(tocSheet, "Move", firstSheet, nil)
	}

	prevSheet := tocSheet
	for _, name := range sheetNames {
		targetSheet := GetSheetByName(sheets, name)
		if targetSheet == nil {
			continue
		}
		w.register(targetSheet)
		oleutil.CallMethod(targetSheet, "Move", nil, prevSheet)
		prevSheet = targetSheet
	}

	oleutil.CallMethod(tocSheet, "Activate")
	return nil
}

// SetupBidirectionalLinks は目次シートと各シートの間に双方向リンクを設置します。
func SetupBidirectionalLinks(w *ExcelWrapper, ws *ole.IDispatch, tocSheet *ole.IDispatch, row int, rawSheetName string, tocName string) {
	sheetName := sanitizeSheetNameLocal(rawSheetName)
	safeTocName := sanitizeSheetNameLocal(tocName)

	if config.CurrentConfig.EnableTocLink {
		text := config.CurrentConfig.ReturnButtonText
		if text == "" {
			text = "← 目次へ"
		}
		if cellA1Var, err := GetPropertyIDispatch(ws, "Range", "A1"); err == nil {
			cellA1 := w.register(cellA1Var)
			oleutil.PutProperty(cellA1, "Value", text)
			if hlksWsVar, err := GetPropertyIDispatch(ws, "Hyperlinks"); err == nil {
				hlksWs := w.register(hlksWsVar)
				oleutil.CallMethod(hlksWs, "Add", cellA1, "", fmt.Sprintf("'%s'!A3", safeTocName))
			}
		}
	}

	if cellEVar, err := oleutil.GetProperty(tocSheet, "Cells", row, 5); err == nil {
		cellE := w.register(cellEVar.ToIDispatch())
		oleutil.PutProperty(cellE, "Value", "link")
		if hlksTocVar, err := GetPropertyIDispatch(tocSheet, "Hyperlinks"); err == nil {
			hlksToc := w.register(hlksTocVar)
			oleutil.CallMethod(hlksToc, "Add", cellE, "", fmt.Sprintf("'%s'!A1", sheetName))
		}
	}
}

// sanitizeSheetNameLocal はシート名の禁止文字を置換します（パッケージ内部用）。
func sanitizeSheetNameLocal(name string) string {
	r := strings.NewReplacer(
		"\\", "_", "/", "_", "?", "_", "*", "_",
		"[", "_", "]", "_", ":", "_",
	)
	result := r.Replace(name)
	if len([]rune(result)) > 31 {
		result = string([]rune(result)[:31])
	}
	return result
}
