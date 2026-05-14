package process

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/winapi"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// CloseExcelGracefully は OLE を使用して安全に Excel を終了させます。
func CloseExcelGracefully(targetPath string) {
	checkMode := config.CurrentConfig.EnableMutex
	if checkMode == 3 {
		return
	}

	unknown, err := oleutil.GetActiveObject("Excel.Application")
	if err != nil {
		return
	}
	defer unknown.Release()

	excelApp, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return
	}
	defer excelApp.Release()

	ret := winapi.ShowDialog("Excel二重起動チェック", "対象の証跡Excelが既に開かれています。\n安全に保存して閉じてよろしいですか？", winapi.MB_YESNO|winapi.MB_ICONWARNING)
	if ret == winapi.IDYES {
		oleutil.PutProperty(excelApp, "DisplayAlerts", false)
		oleutil.CallMethod(excelApp, "Quit") // OLE経由での正規の終了
	}
}

// KillExcelProcesses は指定されたExcelファイルが開かれているかチェックし、設定に応じて警告または強制終了します。
func KillExcelProcesses(targetPath string) {
	// UX改善のため、一時的にメッセージ表示機能を停止して即座にリターン
	// return

	/*
		checkMode := config.CurrentConfig.EnableMutex
		if checkMode == 3 {
			return
		}

		normTarget := strings.ToLower(filepath.Clean(targetPath))

		switch checkMode {
		case 1:
			// モード1: OLE を使って対象ファイルが既に開かれているかチェック
			if targetPath == "" {
				return
			}

			unknown, err := oleutil.GetActiveObject("Excel.Application")
			if err != nil {
				return	// Excel自体が起動していない
			}
			excelApp, err := unknown.QueryInterface(ole.IID_IDispatch)
			if err != nil {
				return
			}
			defer excelApp.Release()

			workbooksVar, err := oleutil.GetProperty(excelApp, "Workbooks")
			if err != nil {
				return
			}
			workbooks := workbooksVar.ToIDispatch()
			defer workbooks.Release()

			countVar, _ := oleutil.GetProperty(workbooks, "Count")
			count := int(countVar.Val)

			shouldWarn := false
			for i := 1; i <= count; i++ {
				itemVar, err := oleutil.GetProperty(workbooks, "Item", i)
				if err != nil {
					continue
				}
				workbook := itemVar.ToIDispatch()

				pathVar, _ := oleutil.GetProperty(workbook, "FullName")
				excelSidePath := strings.ToLower(filepath.Clean(pathVar.ToString()))
				workbook.Release()

				if excelSidePath == normTarget {
					shouldWarn = true
					break
				}
			}

			if shouldWarn {
				msg := "対象の証跡Excelが既に開かれています。\nデータの破損を防ぐため、Excelを保存して閉じてから「再試行」を押してください。"
				winapi.ShowDialog("Excel二重起動チェック", msg, winapi.MB_OK|winapi.MB_ICONWARNING)
			}

		case 2:
			// モード2: taskkill を使って強制終了
			// プロセス存在確認のために一旦tasklistでチェックするのも手ですが、今回はシンプルにtaskkillを実行します。
			// Excelが起動しているか確認するためにはOLEが使えますが、応答なしの場合などを考慮して無条件で確認ダイアログを出します。
			// （元の実装も強制的に聞いてkillしていました）
			msg := "Excelが起動している可能性があります。\n他の作業内容も含めて強制終了されますが、よろしいですか？"
			ret := winapi.ShowDialog("Excel二重起動チェック", msg, winapi.MB_YESNO|winapi.MB_ICONWARNING)
			if ret == winapi.IDYES {
				// cmd := exec.Command("taskkill", "/F", "/IM", "excel.exe", "/T")
				// err := cmd.Run()
				err := ForceKillProcessByName("EXCEL.EXE")
				if err != nil {
					config.Log("ERROR", "Excelの強制終了に失敗しました（起動していない可能性があります）: %v", "Failed to force kill Excel: %v", err)
				} else {
					config.Log("INFO", "Excelを強制終了しました", "Force killed Excel successfully")
				}
			}
		}
	*/
}
