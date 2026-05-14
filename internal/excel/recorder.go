package excel

import (
	"Evidence-Capture/internal/capture"
	"Evidence-Capture/internal/clipboard"
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/winapi"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-ole/go-ole"
)

// Recorder はExcel証跡記録のワークフローを管理する構造体です。
type Recorder struct {
	config types.AppConfig
}

// NewRecorder は新しいRecorderインスタンスを生成します。
func NewRecorder(cfg types.AppConfig) *Recorder {
	return &Recorder{config: cfg}
}

// ProcessCapture は撮影後のExcel貼り付け一連のワークフローを実行します。
func (r *Recorder) ProcessCapture(captureTime time.Time) error {
	// 1. COMの初期化（非同期呼び出し時のパニック防止）
	err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
	if err != nil {
		errorCode := uint32(0)
		if oleErr, ok := err.(*ole.OleError); ok {
			errorCode = uint32(oleErr.Code())
		}
		if errorCode != 1 { // 1 = S_FALSE
			config.Log("ERROR", "ProcessCapture内でのCOM初期化に失敗しました: %v", "COM initialization failed in ProcessCapture: %v", err)
			return fmt.Errorf("COMの初期化に失敗しました")
		}
	}
	defer ole.CoUninitialize()

	w, err := NewExcelWrapper()
	if err != nil {
		winapi.ShowAutoCloseDialog("Excel接続エラー", "Excelへの接続に失敗しました。", 5)
		return err
	}
	defer w.Release()

	if err := w.Open(r.config.ExcelOutputPath); err != nil {
		return err
	}

	// 貼り付け前に、もしゾンビ化や別窓で開いていたら警告・終了させる
	// process.KillExcelProcesses(r.config.ExcelOutputPath)

	// 編集モード検知
	if err := w.CheckEditMode(); err != nil {
		config.Log("ERROR", "Excelが編集モードのため、撮影をスキップしました", "Excel is busy")
		return err
	}

	targetCell := r.DetermineInsertCell(w)

	// ① ローカル保存を先に実行
	r.SaveToLocal(w.App, targetCell, captureTime)

	// ② クリップボードから画像を取得
	img, err := clipboard.GetClipboardImage()
	if err != nil {
		config.Log("ERROR", "クリップボード画像の取得失敗: %v", "Failed to fetch clipboard image: %v", err)
		winapi.ShowDialog(config.T("画像取得失敗", "Capture Error"), config.T("クリップボードから画像を取得できませんでした。", "Failed to get image from clipboard."), winapi.MB_ICONERROR)
		return fmt.Errorf("%v: %v", config.NewMultiLangError("クリップボード画像の取得に失敗しました", "Failed to get clipboard image"), err)
	}

	// ③ 画像を一時ファイルに保存
	tempPath := filepath.Join(config.AppTempDir, fmt.Sprintf("excel_insert_%d.png", captureTime.UnixNano()))
	defer os.Remove(tempPath)

	if err := capture.SaveImageToFile(img, tempPath, 1); err != nil { // 1 = PNG format
		config.Log("ERROR", "一時画像ファイルの保存失敗: %v", "Failed to save temp image: %v", err)
		winapi.ShowDialog(config.T("一時ファイル保存失敗", "Temp File Error"), config.T("画像の一時保存に失敗しました。", "Failed to save temporary image file."), winapi.MB_ICONERROR)
		return fmt.Errorf("%v: %v", config.NewMultiLangError("一時画像ファイルの保存に失敗しました", "Failed to save temp image"), err)
	}

	// ④ 直接ファイルパスでExcelに画像を挿入（リトライ付き）
	success := false
	for i := 0; i < 3; i++ {
		if err := w.InsertImageFromFile(tempPath, targetCell); err == nil {
			success = true
			break
		} else {
			config.Log("ERROR", "画像挿入失敗 (試行 %d/3): %v", "Insert image failed (Attempt %d/3): %v", i+1, err)
		}
		time.Sleep(300 * time.Millisecond)
	}

	if !success {
		// データ救済策：失敗時は一時ファイルを削除せず、rescue_images フォルダに移動する
		exePath, _ := os.Executable()
		rescueDir := filepath.Join(filepath.Dir(exePath), "rescue_images")
		os.MkdirAll(rescueDir, 0755)
		rescuePath := filepath.Join(rescueDir, filepath.Base(tempPath))

		errRename := os.Rename(tempPath, rescuePath)
		msg := config.T("Excelへの画像挿入に失敗しました。", "Failed to insert image to Excel.")
		if errRename == nil {
			msg += fmt.Sprintf(config.T("\n\n撮影データは救済フォルダに保存されました。後ほど手動で貼り付けてください：\n%s", "\n\nRescued image saved to folder. Please paste manually later:\n%s"), rescuePath)
			config.Log("WARN", "画像挿入失敗. 救済パス: %s", "Image rescued to: %s", rescuePath)
		} else {
			// Renameに失敗した場合は元の場所（Temp）を案内する（defer Removeがあるので実際には消える可能性があるが、案内としては残す）
			msg += config.T("\n一時ファイルの保存にも失敗した可能性があります。", "\nAlso failed to save temporary file.")
		}

		winapi.ShowDialog(config.T("画像挿入失敗", "Insert Image Failed"), msg, winapi.MB_ICONERROR)
		return config.NewMultiLangError("試行回数超過により画像挿入に失敗しました", "Failed to insert image after retries")
	}

	// Excel側の描画・認識待ち
	time.Sleep(200 * time.Millisecond)

	// C. タイムスタンプの記入
	if r.config.EnableTimestamp {
		_, row, _ := ParseCellAddress(targetCell)
		tsCol, _, _ := ParseCellAddress(r.config.TimestampCell)
		dynamicTSCell := fmt.Sprintf("%s%d", tsCol, row-1)

		if err := w.WriteTimestamp(dynamicTSCell, captureTime, r.config.TimestampFormat); err != nil {
			config.Log("WARN", "タイムスタンプの記入失敗: %v", "Failed to write timestamp: %v", err)
		}
	}

	// D. 共通の後続処理
	r.AfterPasteActions(w, targetCell)

	// E. 自動保存（設定が有効な場合）
	if config.CurrentConfig.AutoSaveExcel == 2 {
		if err := w.Save(); err != nil {
			config.Log("ERROR", "Excelの自動保存に失敗しました: %v", "Auto Save Failed: %v", err)
		} else {
			config.Log("INFO", "Excelを自動保存しました", "Excel Auto Saved")
		}
	}

	return nil
}

// DetermineInsertCell は最適な挿入位置（セル）を決定します。
func (r *Recorder) DetermineInsertCell(w *ExcelWrapper) string {
	// 1. すでにある画像の一番下の行を取得
	bottomRow := w.GetLatestShapeBottomRow()

	// 設定値から「開始列」と「開始行」を取得
	startCol, initRow, err := ParseCellAddress(r.config.ImageInsertStartCell)
	if err != nil {
		config.Log("ERROR", "セルの解析失敗: %v", "Failed to parse ImageInsertStartCell: %v", err)
	}

	// 2. セルに入力されている文字の最終行を高速取得
	lastCellRow := w.GetLastDataRow(startCol)

	// 3. 画像の底と文字の底、より「下」にある方を採用する
	currentMaxRow := bottomRow
	if lastCellRow > currentMaxRow {
		currentMaxRow = lastCellRow
	}

	// 4. 基準位置にマージンを足す
	targetRow := currentMaxRow + r.config.Margin

	// 5. 画像が1枚もない場合は、設定された初期行を使用
	if bottomRow <= 0 {
		targetRow = initRow
	}

	// 6. 決定したセル自体に文字が入っていないか確認
	for i := 0; i < 100; i++ {
		if w.IsCellEmpty(fmt.Sprintf("%s%d", startCol, targetRow)) {
			break
		}
		targetRow++
	}

	return fmt.Sprintf("%s%d", startCol, targetRow)
}

// SaveToLocal はクリップボード画像をローカルフォルダに非同期保存します。
func (r *Recorder) SaveToLocal(excelApp *ole.IDispatch, targetCell string, t time.Time) {
	if !r.config.LocalSaveOn || r.config.LocalSaveFolder == "" {
		return
	}

	// --- Excelから必要情報を先に取得（COMはgoroutine内で触らない） ---
	wbName, wsName := GetWorkbookAndSheetName(excelApp)

	// サニタイズ処理
	wbName = sanitizeFileName(wbName)
	wsName = sanitizeFileName(wsName)

	if wbName == "" {
		wbName = "UnknownBook"
	}
	if wsName == "" {
		wsName = "UnknownSheet"
	}

	if len([]rune(wbName)) > 30 {
		wbName = string([]rune(wbName)[:30])
	}
	if len([]rune(wsName)) > 30 {
		wsName = string([]rune(wsName)[:30])
	}

	// フォルダ準備
	if err := os.MkdirAll(r.config.LocalSaveFolder, 0755); err != nil {
		config.Log("ERROR", "ローカル保存フォルダの作成失敗: %v", "Failed to create local save folder: %v", err)
		return
	}

	// --- ファイル名の組み立て ---
	timestamp := t.Format("20060102_150405")
	extMap := map[int]string{1: "png", 2: "jpg", 3: "bmp"}
	ext, ok := extMap[r.config.LocalSaveFormat]
	if !ok {
		ext = "png"
	}
	safeCell := sanitizeFileName(targetCell)
	fileName := fmt.Sprintf("%s_%s_%s_%s.%s", wbName, wsName, safeCell, timestamp, ext)
	savePath := filepath.Join(r.config.LocalSaveFolder, fileName)

	// --- 先にクリップボードから画像だけ取得 ---
	img, err := clipboard.GetClipboardImage()
	if err != nil {
		config.Log("ERROR", "クリップボード画像の取得失敗: %v", "Failed to fetch clipboard image: %v", err)
		return
	}

	// --- 重いエンコード + ファイル保存を非同期化 ---
	go func() {
		if err := capture.SaveImageToFile(img, savePath, r.config.LocalSaveFormat); err != nil {
			config.Log("ERROR", "ローカル保存失敗: %v", "Failed to save locally: %v", err)
		}
	}()
}

// AfterPasteActions は貼り付け後の共通処理（スクロール、ステータスバー表示）を行います。
func (r *Recorder) AfterPasteActions(w *ExcelWrapper, targetCell string) {
	_, row, err := ParseCellAddress(targetCell)
	if err != nil {
		return
	}

	// スクロール位置の調整
	scrollTarget := row - 5
	if scrollTarget < 1 {
		scrollTarget = 1
	}

	w.ScrollToRow(scrollTarget)
	w.PutStatusText(fmt.Sprintf("証跡取得完了: %s", targetCell))
}

func sanitizeFileName(name string) string {
	r := strings.NewReplacer(
		"\\", "_", "/", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return r.Replace(name)
}
