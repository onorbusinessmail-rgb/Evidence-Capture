// Package excel は Excel OLE操作のコアロジックを提供します。
// このパッケージは config と winapi に依存します。
package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/winapi"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/lxn/walk"
	"github.com/xuri/excelize/v2"
)

// =========================================================================
// ExcelWrapper: OLE操作の共通ラッパー
// =========================================================================

// ExcelWrapper は Excel OLE 操作を抽象化し、リソース管理を一元化する構造体です。
type ExcelWrapper struct {
	App         *ole.IDispatch
	Workbook    *ole.IDispatch
	ActiveSheet *ole.IDispatch
	dispatches  []*ole.IDispatch // 自動解放用
}

// NewExcelWrapper は Excel インスタンスを取得または新規起動し、ラッパーを返します。
// OLE の初期化 (CoInitialize) も内部で行います。
func NewExcelWrapper() (*ExcelWrapper, error) {
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		errorCode := uint32(0)
		if oleErr, ok := err.(*ole.OleError); ok {
			errorCode = uint32(oleErr.Code())
		}
		// S_FALSE (0x00000001) 以外はエラーとして扱う
		if errorCode != winapi.S_FALSE {
			config.Log("ERROR", "COMの初期化に失敗しました: %v", "COM initialization failed: %v", err)
			return nil, fmt.Errorf("COMの初期化に失敗しました")
		}
	}

	unknown, err := oleutil.GetActiveObject("Excel.Application")
	if err != nil {
		unknown, err = oleutil.CreateObject("Excel.Application")
		if err != nil {
			// CoInitializeEx が成功または S_FALSE を返している場合のみ Uninitialize する
			ole.CoUninitialize()
			config.Log("ERROR", "Excelの起動に失敗しました: %v", "Failed to start Excel: %v", err)
			return nil, fmt.Errorf("Excelの起動に失敗しました")
		}
	}
	// QueryInterface 後は unknown 自体は不要になるため、適切に解放する
	excelApp, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		unknown.Release() // CoUninitialize 呼び出し前に明示的に解放
		ole.CoUninitialize()
		return nil, err
	}
	defer unknown.Release() // 正常系では関数の最後に解放されるようにする

	w := &ExcelWrapper{
		App: excelApp,
	}
	w.register(excelApp)
	return w, nil
}

// register は IDispatch オブジェクトを管理スライスに追加し、一括解放の対象にします。
func (w *ExcelWrapper) register(d *ole.IDispatch) *ole.IDispatch {
	if d != nil {
		w.dispatches = append(w.dispatches, d)
	}
	return d
}

// Release は保持している全ての OLE オブジェクトを逆順で解放します。
func (w *ExcelWrapper) Release() {
	for i := len(w.dispatches) - 1; i >= 0; i-- {
		if w.dispatches[i] != nil {
			w.dispatches[i].Release()
			w.dispatches[i] = nil
		}
	}
	w.dispatches = nil
	w.App = nil
	w.Workbook = nil
	w.ActiveSheet = nil
	ole.CoUninitialize()
}

// CheckEditMode は Excel が編集モードでないか確認し、編集中であればユーザーに通知します。
func (w *ExcelWrapper) CheckEditMode() error {
	// 1. SendMessageTimeoutによる応答(ハング/ビジー)チェック
	hwndVar, _ := oleutil.GetProperty(w.App, "Hwnd")
	if hwndVar != nil {
		hwnd := uintptr(hwndVar.Val)
		var result uintptr
		const StrictTimeout = 1000 // 1秒
		ret, _, _ := winapi.SendMessageTimeout.Call(
			hwnd,
			winapi.WM_NULL,
			0,
			0,
			winapi.SMTO_ABORTIFHUNG,
			StrictTimeout,
			uintptr(unsafe.Pointer(&result)),
		)
		if ret == 0 {
			config.Log("WARN", "Excelが応答しません。操作を中断します。", "Excel Busy/Hung")
			winapi.ShowAutoCloseDialog("Excelビジー", "Excelが編集中のため、安全のために処理を中断しました。Enter等で編集を終了してください。", 5)
			return fmt.Errorf("Excel is busy or hung")
		}
	}

	// 2. Readyプロパティによる編集モードチェック
	readyVar, err := oleutil.GetProperty(w.App, "Ready")
	if err != nil {
		return err
	}

	if readyVar.Value().(bool) {
		return nil // 編集モードではない（正常）
	}

	// 3. 編集モード検知時の処理（ウィンドウを復元して前面に持ってくる）
	oleutil.PutProperty(w.App, "WindowState", XlNormal)

	// ExcelWrapperに既に実装されている共通メソッドを再利用
	w.BringToForeground()

	title := "Excel編集中アラート"
	msg := "【重要】Excelが入力中または編集中です。\n\n1. 確定(Enter)または中止(Esc)してください。\n2. その後、再度実行してください。"
	winapi.ShowAutoCloseDialog(title, msg, 5)
	return fmt.Errorf("Excel is in edit mode")
}

// Save は現在開いているブックを上書き保存します。
func (w *ExcelWrapper) Save() error {
	if w.Workbook == nil {
		return fmt.Errorf("no workbook opened")
	}
	_, err := oleutil.CallMethod(w.Workbook, "Save")
	return err
}

// Open は指定されたパスのブックを開きます。既に開いている場合はアクティブにします。
func (w *ExcelWrapper) Open(targetPath string) error {
	if err := w.CheckEditMode(); err != nil {
		return err
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = targetPath
	}
	absPath = filepath.Clean(absPath)

	oleutil.PutProperty(w.App, "Visible", true)

	workbooksVar, err := oleutil.GetProperty(w.App, "Workbooks")
	if err != nil {
		return fmt.Errorf("Workbooksの取得に失敗: %v", err)
	}
	workbooks := w.register(workbooksVar.ToIDispatch())

	// 既に開いているか判定
	countVar, _ := oleutil.GetProperty(workbooks, "Count")
	count := int(countVar.Val)
	found := false

	for i := 1; i <= count; i++ {
		itemVar, err := oleutil.GetProperty(workbooks, "Item", i)
		if err != nil {
			continue
		}
		workbook := itemVar.ToIDispatch()

		pathVar, pErr := oleutil.GetProperty(workbook, "FullName")
		if pErr != nil {
			workbook.Release() // パス取得失敗時も解放
			continue
		}
		excelSidePath := filepath.Clean(pathVar.ToString())

		if strings.EqualFold(excelSidePath, absPath) {
			found = true
			w.Workbook = w.register(workbook) // ここで管理対象にする
			oleutil.CallMethod(workbook, "Activate")
			break
		}
		workbook.Release() // 一致しない場合は即座に解放
	}

	if !found {
		wbVar, err := oleutil.CallMethod(workbooks, "Open", absPath)
		if err != nil {
			config.Log("ERROR", "ファイルのオープンに失敗しました: %s", "Failed to open file: %s", absPath)
			return fmt.Errorf("ファイルのオープンに失敗しました: %s", absPath)
		}
		w.Workbook = w.register(wbVar.ToIDispatch())
	}

	// ActiveSheet をセット
	sheetVar, err := oleutil.GetProperty(w.App, "ActiveSheet")
	if err == nil {
		w.ActiveSheet = w.register(sheetVar.ToIDispatch())
	}

	return nil
}

// EnsureSheet は指定された名前のシートが存在することを確認し、アクティブにします。
// シートがない場合は新規作成し、ActiveSheet へのセットまで行います。
func (w *ExcelWrapper) EnsureSheet(name string) error {
	if w.Workbook == nil {
		return fmt.Errorf("Workbook が開かれていません")
	}

	sheetsVar, err := oleutil.GetProperty(w.Workbook, "Sheets")
	if err != nil {
		return err
	}
	sheets := w.register(sheetsVar.ToIDispatch())

	// 内部で GetSheetByName を利用しつつ、w.ActiveSheet を更新する
	sheet := GetSheetByName(sheets, name)
	if sheet != nil {
		w.ActiveSheet = w.register(sheet)
		oleutil.CallMethod(w.ActiveSheet, "Activate")
	} else {
		// 存在しない場合は新規作成
		newSheetVar, err := oleutil.CallMethod(sheets, "Add")
		if err != nil {
			return fmt.Errorf("シート [%s] の作成に失敗しました: %v", name, err)
		}
		w.ActiveSheet = w.register(newSheetVar.ToIDispatch())
		oleutil.PutProperty(w.ActiveSheet, "Name", name)
	}
	return nil
}

// PasteImage はクリップボードの画像を貼り付け、リサイズを行います。
func (w *ExcelWrapper) PasteImage(cellAddr string, scale float64) error {
	// 1. 編集モードチェック (TryPasteToExcelOnly の内部仕様を継承)
	if err := w.CheckEditMode(); err != nil {
		return err
	}

	if w.ActiveSheet == nil {
		return fmt.Errorf("ActiveSheet が取得されていません")
	}

	// 2. 指定セルを選択して貼り付け
	cellVar, err := oleutil.GetProperty(w.ActiveSheet, "Range", cellAddr)
	if err != nil {
		return err
	}
	cell := w.register(cellVar.ToIDispatch())
	oleutil.CallMethod(cell, "Select")

	_, err = oleutil.CallMethod(w.ActiveSheet, "Paste")
	if err != nil {
		return err
	}

	// 3. 貼り付け後のリサイズを実行 (ResizeLastShape の代替手段を内部呼び出し)
	if scale > 0 {
		w.ResizeLastShape(scale)
	}

	return nil
}

// ResizeLastShape は現在の選択（画像）をリサイズします。
func (w *ExcelWrapper) ResizeLastShape(scale float64) {
	selectionVar, err := oleutil.GetProperty(w.App, "Selection")
	if err != nil {
		return
	}
	selection := w.register(selectionVar.ToIDispatch())

	shpRangeVar, err := oleutil.GetProperty(selection, "ShapeRange")
	if err != nil {
		return
	}
	shpRange := w.register(shpRangeVar.ToIDispatch())

	// アスペクト比を維持してスケーリング
	oleutil.CallMethod(shpRange, "ScaleWidth", scale, 0, 0)
	oleutil.CallMethod(shpRange, "ScaleHeight", scale, 0, 0)
}

// InsertImageFromFile は指定されたパスの画像をExcelの指定セル位置に挿入します
func (w *ExcelWrapper) InsertImageFromFile(filePath string, cellName string) error {
	// 1. ファイルの存在確認
	_, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("画像ファイルが見つかりません: %s", filePath)
	}

	if w.ActiveSheet == nil {
		return fmt.Errorf("ActiveSheet が取得されていません")
	}

	// 2. 挿入位置（セル）の取得
	rangeObj, err := oleutil.GetProperty(w.ActiveSheet, "Range", cellName)
	if err != nil {
		return fmt.Errorf("セルの取得に失敗: %w", err)
	}
	rangeIDisp := w.register(rangeObj.ToIDispatch())

	// 3. セルの位置情報（Left, Top）を取得
	leftVar, _ := oleutil.GetProperty(rangeIDisp, "Left")
	topVar, _ := oleutil.GetProperty(rangeIDisp, "Top")

	leftVal := leftVar.Value()
	topVal := topVar.Value()

	// 4. Shapesコレクションを取得
	shapesVar, err := oleutil.GetProperty(w.ActiveSheet, "Shapes")
	if err != nil {
		return fmt.Errorf("Shapesの取得に失敗: %w", err)
	}
	shapesIDisp := w.register(shapesVar.ToIDispatch())

	// 5. AddPictureメソッドの実行
	// 引数: Filename, LinkToFile, SaveWithDocument, Left, Top, Width, Height
	// -1 を指定すると画像本来のサイズで挿入されます
	_, err = oleutil.CallMethod(shapesIDisp, "AddPicture", filePath, false, true, leftVal, topVal, -1, -1)
	if err != nil {
		return fmt.Errorf("画像の挿入に失敗: %w", err)
	}

	return nil
}

// WriteCell は指定したセルアドレスに値を書き込みます。
func (w *ExcelWrapper) WriteCell(cellAddr string, value interface{}) error {
	if w.ActiveSheet == nil {
		return fmt.Errorf("ActiveSheet が取得されていません")
	}

	cellVar, err := oleutil.GetProperty(w.ActiveSheet, "Range", cellAddr)
	if err != nil {
		return err
	}
	cell := w.register(cellVar.ToIDispatch())

	_, err = oleutil.PutProperty(cell, "Value", value)
	return err
}

// SetWindowState はウィンドウの表示状態を切り替えます。
func (w *ExcelWrapper) SetWindowState(minimize bool) {
	if minimize {
		oleutil.PutProperty(w.App, "WindowState", XlMinimized)
	} else {
		oleutil.PutProperty(w.App, "Visible", true)
		oleutil.PutProperty(w.App, "WindowState", XlNormal)
		w.BringToForeground()
	}
}

// BringToForeground は Excel を最前面に表示します。
func (w *ExcelWrapper) BringToForeground() {
	if captionVar, err := oleutil.GetProperty(w.App, "Caption"); err == nil {
		caption := captionVar.ToString()
		pCaption, _ := syscall.UTF16PtrFromString(caption)
		hwnd, _, _ := winapi.FindWindow.Call(0, uintptr(unsafe.Pointer(pCaption)))
		if hwnd != 0 {
			winapi.ShowWindow.Call(hwnd, winapi.SW_RESTORE)
			winapi.SetForegroundWindow.Call(hwnd)
		}
	}
}

// GetLatestShapeBottomRow は画像データの最大行番号を取得します。
func (w *ExcelWrapper) GetLatestShapeBottomRow() int {
	if w.ActiveSheet == nil {
		return 0
	}

	shapesVar, err := oleutil.GetProperty(w.ActiveSheet, "Shapes")
	if err != nil {
		return 0
	}
	shapesObj := w.register(shapesVar.ToIDispatch())

	countVar, _ := oleutil.GetProperty(shapesObj, "Count")
	count := int(countVar.Val)

	maxRow := 0
	for i := 1; i <= count; i++ {
		shapeVar, err := oleutil.CallMethod(shapesObj, "Item", i)
		if err != nil {
			continue
		}
		sObj := shapeVar.ToIDispatch()
		// 各図形の右下端セルから行番号を取得
		brCellVar, err := oleutil.GetProperty(sObj, "BottomRightCell")
		if err == nil {
			brObj := brCellVar.ToIDispatch()
			rowVar, _ := oleutil.GetProperty(brObj, "Row")
			currentRow := int(rowVar.Val)
			if currentRow > maxRow {
				maxRow = currentRow
			}
			brObj.Release()
		}
		sObj.Release()
	}
	return maxRow
}

// GetLastDataRow は指定列の最終データ行を取得します。
func (w *ExcelWrapper) GetLastDataRow(col string) int {
	if w.ActiveSheet == nil {
		return 0
	}

	rowsVar, _ := oleutil.GetProperty(w.ActiveSheet, "Rows")
	rows := w.register(rowsVar.ToIDispatch())
	rowCountVar, _ := oleutil.GetProperty(rows, "Count")
	maxRow := rowCountVar.Val

	cellVar, _ := oleutil.GetProperty(w.ActiveSheet, "Cells", maxRow, col)
	bottomCell := w.register(cellVar.ToIDispatch())

	endVar, err := oleutil.GetProperty(bottomCell, "End", XlUp)
	if err != nil {
		return 0
	}
	lastCell := w.register(endVar.ToIDispatch())

	rowVar, _ := oleutil.GetProperty(lastCell, "Row")
	lastRow := int(rowVar.Val)

	if lastRow == 1 {
		val, _ := oleutil.GetProperty(lastCell, "Value")
		if val.Value() == nil || val.ToString() == "" {
			return 0
		}
	}
	return lastRow
}

// PutStatusText はステータスバーにメッセージを表示します。
func (w *ExcelWrapper) PutStatusText(text string) {
	oleutil.PutProperty(w.App, "StatusBar", text)
}

// ResetStatusText はステータスバーを初期状態に戻します。
func (w *ExcelWrapper) ResetStatusText() {
	oleutil.PutProperty(w.App, "StatusBar", false)
}

// ScrollToRow は指定行にスクロールします。
func (w *ExcelWrapper) ScrollToRow(row int) {
	windowVar, err := oleutil.GetProperty(w.App, "ActiveWindow")
	if err == nil && windowVar.Value() != nil {
		window := w.register(windowVar.ToIDispatch())
		oleutil.PutProperty(window, "ScrollRow", row)
	}
}

// handleError はエラー発生時の共通処理を行います。
func (w *ExcelWrapper) handleError(method string, r interface{}) {
	title := "Excel操作エラー"
	msg := fmt.Sprintf("【重要】Excel操作（%s）中にエラーが発生しました。\n\n%v", method, r)
	winapi.ShowAutoCloseDialog(title, msg, 5)
	config.Log("ERROR", "メソッド %s でパニックが発生しました: %v", "Excel operation error: method %s, panic: %v", method, r)
}

// =========================================================================
// ユーティリティ関数
// =========================================================================

// GetPropertyIDispatch はプロパティを取得して IDispatch 型で返します。
func GetPropertyIDispatch(obj *ole.IDispatch, name string, args ...interface{}) (*ole.IDispatch, error) {
	v, err := oleutil.GetProperty(obj, name, args...)
	if err != nil {
		return nil, err
	}
	return v.ToIDispatch(), nil
}

// GetSheetByName は指定された名前のシートを返します。
func GetSheetByName(sheets *ole.IDispatch, name string) *ole.IDispatch {
	itemVar, err := oleutil.GetProperty(sheets, "Item", name)
	if err != nil {
		return nil
	}
	return itemVar.ToIDispatch()
}

// GetTOCName は設定から目次シート名を取得します。
func GetTOCName() string {
	if config.CurrentConfig.TocSheetName == "" {
		return "目次"
	}
	return config.CurrentConfig.TocSheetName
}

// // HandleExcelEditMode は Excel の編集モードを検知し、ユーザーに通知します。
// func HandleExcelEditMode(excelApp *ole.IDispatch, targetMw *walk.MainWindow) (bool, error) {
// 	hwndVar, _ := oleutil.GetProperty(excelApp, "Hwnd")
// 	if hwndVar != nil {
// 		hwnd := uintptr(hwndVar.Val)
// 		var result uintptr
// 		const StrictTimeout = 1000 // 1秒
// 		ret, _, _ := winapi.SendMessageTimeout.Call(
// 			hwnd,
// 			winapi.WM_NULL,
// 			0,
// 			0,
// 			winapi.SMTO_ABORTIFHUNG,
// 			StrictTimeout,
// 			uintptr(unsafe.Pointer(&result)),
// 		)
// 		if ret == 0 {
// 			// エラー詳細を確認し、単なるタイムアウト以外（プロセス消失等）もケア
// 			config.LogDual("WARN", "Excel Busy/Hung", "Excelが応答しません。操作を中断します。")
// 			winapi.ShowAutoCloseDialog("Excelビジー", "Excelが編集中のため、安全のために処理を中断しました。Enter等で編集を終了してください。", 5)
// 			return false, nil
// 		}
// 	}

// 	readyVar, err := oleutil.GetProperty(excelApp, "Ready")
// 	if err != nil {
// 		return false, err
// 	}

// 	if readyVar.Value().(bool) {
// 		return true, nil
// 	}

// 	if targetMw != nil {
// 		targetMw.Hide()
// 		defer targetMw.Show()
// 	}

// 	oleutil.PutProperty(excelApp, "WindowState", XlNormal)

// 	if captionVar, err := oleutil.GetProperty(excelApp, "Caption"); err == nil {
// 		caption := captionVar.ToString()
// 		pCaption, _ := syscall.UTF16PtrFromString(caption)
// 		hwnd, _, _ := winapi.FindWindow.Call(0, uintptr(unsafe.Pointer(pCaption)))
// 		if hwnd != 0 {
// 			winapi.ShowWindow.Call(hwnd, winapi.SW_RESTORE)
// 			winapi.SetForegroundWindow.Call(hwnd)
// 		}
// 	}

// 	title := "Excel編集中アラート"
// 	msg := "【重要】Excelが入力中または編集中です。\n\n1. 確定(Enter)または中止(Esc)してください。\n2. その後、再度実行してください。"
// 	winapi.ShowAutoCloseDialog(title, msg, 5)
// 	return false, nil
// }

// -------------------------------------------------------------------------
// 互換用関数 (古いコード用。徐々に削除)
// -------------------------------------------------------------------------

// func GetOrOpenExcel(path string) (*ole.IDispatch, error) {
// 	w, err := NewExcelWrapper()
// 	if err != nil {
// 		return nil, err
// 	}
// 	// この関数は *ole.IDispatch を返す必要があるため、Wrapper を捨てて App だけ返す（危険）
// 	// 将来的には全ての呼び出し元を ExcelWrapper に置き換えるべき。
// 	return w.App, nil
// }

// func TryPasteToExcelOnly(excel *ole.IDispatch, cellAddr string) bool {
// 	w := &ExcelWrapper{App: excel}
// 	sheetVar, _ := oleutil.GetProperty(excel, "ActiveSheet")
// 	w.ActiveSheet = sheetVar.ToIDispatch()
// 	err := w.PasteImage(cellAddr, 0)
// 	w.ActiveSheet.Release()
// 	return err == nil
// }

// func ResizeLastShape(excelApp *ole.IDispatch, scale float64) {
// 	w := &ExcelWrapper{App: excelApp}
// 	w.ResizeLastShape(scale)
// }

// func GetLatestShapeBottomRow(excelApp *ole.IDispatch) int {
// 	w := &ExcelWrapper{App: excelApp}
// 	sheetVar, _ := oleutil.GetProperty(excelApp, "ActiveSheet")
// 	w.ActiveSheet = sheetVar.ToIDispatch()
// 	res := w.GetLatestShapeBottomRow()
// 	w.ActiveSheet.Release()
// 	return res
// }

func GetLastDataRow(excelApp *ole.IDispatch, col string) int {
	w := &ExcelWrapper{App: excelApp}
	sheetVar, _ := oleutil.GetProperty(excelApp, "ActiveSheet")
	w.ActiveSheet = sheetVar.ToIDispatch()
	res := w.GetLastDataRow(col)
	w.ActiveSheet.Release()
	return res
}

// WriteTimestamp は指定セルに設定された書式で時刻を書き込みます。
func (w *ExcelWrapper) WriteTimestamp(cellAddr string, t time.Time, format int) error {
	var layout string
	switch format {
	case 1:
		layout = "2006/01/02 15:04:05"
	case 2:
		layout = "2006-01-02 15:04:05"
	case 3:
		layout = "2006年01月02日 15:04:05"
	default:
		layout = "2006/01/02 15:04:05"
	}

	// 文字列として認識させるためシングルクォートを付与
	formattedTime := "'" + t.Format(layout)
	return w.WriteCell(cellAddr, formattedTime)
}

func ScrollToRow(excel *ole.IDispatch, row int) {
	w := &ExcelWrapper{App: excel}
	w.ScrollToRow(row)
}

// func PutStatusText(excelApp *ole.IDispatch, text string) {
// 	w := &ExcelWrapper{App: excelApp}
// 	w.PutStatusText(text)
// }

// func ResetStatusText(excelApp *ole.IDispatch) {
// 	oleutil.PutProperty(excelApp, "StatusBar", false)
// }

// IsCellEmpty は指定したセルが空（nil または空文字）であるか判定します。
func (w *ExcelWrapper) IsCellEmpty(cellAddr string) bool {
	if w.ActiveSheet == nil {
		return true
	}
	cellVar, err := oleutil.GetProperty(w.ActiveSheet, "Range", cellAddr)
	if err != nil {
		return true
	}
	cell := cellVar.ToIDispatch()
	defer cell.Release()

	valVar, _ := oleutil.GetProperty(cell, "Value")
	val := valVar.Value()
	return val == nil || strings.TrimSpace(fmt.Sprint(val)) == ""
}

func GetWorkbookAndSheetName(excelApp *ole.IDispatch) (string, string) {
	var wbName, wsName string
	if wbVar, err := oleutil.GetProperty(excelApp, "ActiveWorkbook"); err == nil {
		wb := wbVar.ToIDispatch()
		if n, err := oleutil.GetProperty(wb, "Name"); err == nil {
			wbName = n.ToString()
			wbName = strings.TrimSuffix(wbName, filepath.Ext(wbName))
		}
		wb.Release()
	}
	if wsVar, err := oleutil.GetProperty(excelApp, "ActiveSheet"); err == nil {
		ws := wsVar.ToIDispatch()
		if n, err := oleutil.GetProperty(ws, "Name"); err == nil {
			wsName = n.ToString()
		}
		ws.Release()
	}
	return wbName, wsName
}

func BringExcelToForeground(excelApp *ole.IDispatch) {
	w := &ExcelWrapper{App: excelApp}
	w.BringToForeground()
}

func SetExcelWindowState(targetPath string, minimize bool) {
	w, err := NewExcelWrapper()
	if err != nil {
		return
	}
	defer w.Release()

	if err := w.Open(targetPath); err != nil {
		return
	}

	w.SetWindowState(minimize)
}

// CreateNewExcelAndSaveAs はexcelizeを使用して新規Excelを作成し、
// エクスプローラーを開いて即時「名前を付けて保存」を行います。
func CreateNewExcelAndSaveAs(owner walk.Form) (string, error) {
	// 1. excelizeで新規ブック作成（一般的なレイアウト、設定は無視）
	f := excelize.NewFile()
	defer f.Close()

	// デフォルトの「Sheet1」をアクティブにする
	f.SetActiveSheet(0)

	// 2. 保存先ダイアログのデフォルトパスを取得
	// config_evidence.ini の [1_証跡取得-基本動作設定] 環境設定：証跡格納先Excelファイルパス
	defaultPath := config.CurrentConfig.ExcelOutputPath

	// パスが空またはディレクトリでない場合は実行ディレクトリをデフォルトに
	initialDir := filepath.Dir(defaultPath)
	if defaultPath == "" || initialDir == "." {
		exePath, _ := os.Executable()
		initialDir = filepath.Dir(exePath)
	}

	// 3. 「名前を付けて保存」ダイアログを開く
	dlg := new(walk.FileDialog)
	dlg.Title = config.T("新規証跡ファイル作成", "Create New Evidence File")
	dlg.Filter = "Excel Files (*.xlsx)|*.xlsx"
	// 存在しないフィールドを削除し、正しいフィールド名に修正
	dlg.InitialDirPath = initialDir

	if ok, err := dlg.ShowSave(owner); err != nil {
		return "", err
	} else if !ok {
		return "", fmt.Errorf("保存がキャンセルされました")
	}

	savePath := dlg.FilePath
	if !strings.HasSuffix(strings.ToLower(savePath), ".xlsx") {
		savePath += ".xlsx"
	}

	// 4. 指定されたパスに保存
	if err := f.SaveAs(savePath); err != nil {
		return "", fmt.Errorf("ファイルの保存に失敗しました: %v", err)
	}

	// 5. 保存したファイルをそのままExcel Appで開く（ユーザーがすぐ使えるように）
	if w, err := NewExcelWrapper(); err == nil {
		w.Open(savePath)
		w.Release()
	} else {
		config.Log("WARN", "ファイルは作成されましたが、Excelで開く際にエラーが発生しました: %v", "Created but failed to open in Excel App: %v", err)
	}

	return savePath, nil
}

// CreateAndRegisterNewExcel は新規Excelを作成し、アプリ設定への登録と保存までを行います。
func CreateAndRegisterNewExcel(owner walk.Form) (string, error) {
	// 1. ファイル作成と保存ダイアログ処理
	newPath, err := CreateNewExcelAndSaveAs(owner)
	if err != nil {
		return "", err
	}

	// 2. 成功した場合、グローバル設定のパスを更新
	config.CurrentConfig.ExcelOutputPath = newPath

	// 3. 設定ファイル(INI)に永続化
	if err := config.SaveAppConfig(config.CurrentConfig); err != nil {
		config.Log("ERROR", "設定の保存に失敗しました: %v", "Config Save Failed: %v", err)
		return newPath, fmt.Errorf("ファイルは作成されましたが、設定の保存に失敗しました: %w", err)
	}

	return newPath, nil
}

// SelectExistingExcel は既存のExcelファイルを選択し、アプリ設定への登録と保存までを行います。
func SelectExistingExcel(owner walk.Form) (string, error) {
	// 1. ファイル選択ダイアログ処理
	dlg := new(walk.FileDialog)
	dlg.Title = config.T("既存のExcelを選択", "Select Existing Excel")
	dlg.Filter = config.T("Excelファイル (*.xlsx;*.xlsm)|*.xlsx;*.xlsm", "Excel Files (*.xlsx;*.xlsm)|*.xlsx;*.xlsm")

	if ok, err := dlg.ShowOpen(owner); err != nil {
		return "", err
	} else if !ok {
		return "", nil // キャンセル時はエラーにせず空文字を返す
	}

	selectedPath := dlg.FilePath

	// 2. グローバル設定のパスを更新
	config.CurrentConfig.ExcelOutputPath = selectedPath

	// 3. 設定ファイル(INI)に永続化
	if err := config.SaveAppConfig(config.CurrentConfig); err != nil {
		config.Log("ERROR", "設定の保存に失敗しました: %v", "Config Save Failed: %v", err)
		return selectedPath, fmt.Errorf("ファイルは選択されましたが、設定の保存に失敗しました: %w", err)
	}

	return selectedPath, nil
}
