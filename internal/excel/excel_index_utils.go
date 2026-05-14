package excel

import (
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// FindSheetRowByName は指定されたシート内で、B列（シート名列）から対象名を検索し、その行番号を返します。
// 見つからない場合は 0 を返します。
func FindSheetRowByName(idxSheet *ole.IDispatch, sheetName string) (int, error) {
	// B列全体を対象にする (Columns(2))
	colsVar, err := oleutil.GetProperty(idxSheet, "Columns", 2)
	if err != nil {
		return 0, err
	}
	cols := colsVar.ToIDispatch()
	defer cols.Release()

	// Findメソッドの呼び出し: Find(What:=sheetName, LookAt:=1 /* xlWhole */)
	// xlWhole(1) を指定することで部分一致を避け、完全一致で検索します
	foundCellVar, err := oleutil.CallMethod(cols, "Find", sheetName, ole.IID_NULL, -4163, 1)
	if err != nil || foundCellVar.Value() == nil {
		return 0, nil // 見つからない場合はエラーではなく 0 を返す
	}

	foundCell := foundCellVar.ToIDispatch()
	defer foundCell.Release()

	rowVar, err := oleutil.GetProperty(foundCell, "Row")
	if err != nil {
		return 0, err
	}

	return int(rowVar.Val), nil
}

// GetNextIndexRow はインデックスシートの末尾行（次の書き込み位置）を特定します。
func GetNextIndexRow(idxSheet *ole.IDispatch) (int, error) {
	// B列の最終行から上方向に検索 (XlUp = -4162)
	rowsVar, _ := oleutil.GetProperty(idxSheet, "Rows")
	rows := rowsVar.ToIDispatch()
	defer rows.Release()

	countVar, _ := oleutil.GetProperty(rows, "Count")
	maxRow := countVar.Val

	cellVar, _ := oleutil.GetProperty(idxSheet, "Cells", maxRow, 2)
	bottomCell := cellVar.ToIDispatch()
	defer bottomCell.Release()

	endVar, err := oleutil.GetProperty(bottomCell, "End", -4162)
	if err != nil {
		return 4, nil // 失敗時はデフォルトの開始行
	}
	lastCell := endVar.ToIDispatch()
	defer lastCell.Release()

	rowVar, _ := oleutil.GetProperty(lastCell, "Row")
	lastRow := int(rowVar.Val)

	// データが全くない場合やヘッダーのみの場合は4行目から
	if lastRow < 4 {
		return 4, nil
	}
	return lastRow + 1, nil
}
