package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"
	"Evidence-Capture/internal/winapi"
	"os"
	"path/filepath"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// openExcelParentFolder は設定されたExcelパスの親フォルダ（実在するもの）をエクスプローラーで開きます。
func openExcelParentFolder(owner walk.Form, path string) {
	if path == "" {
		walk.MsgBox(owner, config.T("通知", "Notice"), config.T("保存先パスが設定されていないため、フォルダを特定できません。", "The save path is not set, so the folder cannot be identified."), walk.MsgBoxIconInformation)
		return
	}

	dir := filepath.Dir(path)
	for dir != "" {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			winapi.OpenFolder(dir)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// どこも見つからない場合はデスクトップなどを開く試み（フォールバック）
	winapi.OpenFolder(os.Getenv("USERPROFILE") + "\\Desktop")
}

// RunInitialSetupDialog はアプリ起動時にExcelパスが未設定、またはファイルが存在しない場合に表示するダイアログです。
// reason: 1=未設定(Empty), 2=ファイル紛失(Missing)
// 設定が完了した場合は true, キャンセルされた場合は false を返します。
func RunInitialSetupDialog(reason int) bool {
	// ツール起動時のパス状態を保持（条件A判定用）
	initialPath := config.CurrentConfig.ExcelOutputPath

	var dlg *walk.Dialog
	setupDone := false

	// メッセージの決定
	var msg string
	if reason == 2 {
		msg = config.T("設定されたExcelファイルが見つかりません。\n移動または削除された可能性があります。",
			"The configured Excel file was not found.\nIt may have been moved or deleted.")
	} else {
		msg = config.T("証跡を保存するExcelファイルが設定されていません。",
			"Excel file for saving evidence is not set.")
	}

	err := Dialog{
		AssignTo: &dlg,
		Title:    config.T("初期設定：証跡保存先の確認", "Initial Setup: Verify Excel Destination"),
		Icon:     2,
		MinSize:  Size{Width: 480, Height: 300},
		Layout:   VBox{Margins: Margins{Left: 25, Top: 25, Right: 25, Bottom: 20}, Spacing: 15},
		Children: []Widget{
			// ヘッダーメッセージ
			Label{
				Text: msg,
				Font: Font{Family: "Meiryo UI", PointSize: 10, Bold: true},
			},
			Label{
				Text: config.T("再度、新規作成するか既存のファイルを選択してください。", "Please create a new one or select an existing file again."),
			},

			// ファイル紛失時のみ、現在の設定状況をグループ化して表示
			GroupBox{
				Visible: reason == 2,
				Title:   config.T("現在の設定状況", "Current Settings"),
				Layout:  VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}, Spacing: 10},
				Children: []Widget{
					Label{
						Text: initialPath,
						Font: Font{Family: "Consolas", PointSize: 9},
					},
					Composite{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							PushButton{
								Text:    config.T("📁 存在する親フォルダを開く", "📁 Open Existing Parent Folder"),
								MinSize: Size{Height: 30},
								OnClicked: func() {
									openExcelParentFolder(dlg, initialPath)
								},
							},
							HSpacer{},
						},
					},
				},
			},

			VSpacer{},

			// メインアクションボタン
			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 15},
				Children: []Widget{
					PushButton{
						Text:    config.T("✨ Excelを新規作成", "✨ Create New Excel"),
						MinSize: Size{Width: 180, Height: 45},
						Font:    Font{Family: "Meiryo UI", PointSize: 10, Bold: true},
						OnClicked: func() {
							newPath, err := excel.CreateAndRegisterNewExcel(dlg)
							if err != nil {
								return
							}
							if newPath != "" {
								setupDone = true
								dlg.Accept()
							}
						},
					},
					PushButton{
						Text:    config.T("🔍 既存のExcelを選択", "🔍 Select Existing Excel"),
						MinSize: Size{Width: 180, Height: 45},
						Font:    Font{Family: "Meiryo UI", PointSize: 10},
						OnClicked: func() {
							openExcelParentFolder(dlg, initialPath)

							fd := new(walk.FileDialog)
							fd.Title = config.T("証跡保存先のExcelファイルを選択", "Select Excel File for Evidence")
							fd.Filter = config.T("Excelファイル (*.xlsx;*.xlsm)|*.xlsx;*.xlsm|すべてのファイル (*.*)|*.*", "Excel Files (*.xlsx;*.xlsm)|*.xlsx;*.xlsm|All Files (*.*)|*.*")

							if ok, _ := fd.ShowOpen(dlg); ok {
								if fi, err := os.Stat(fd.FilePath); err != nil || fi.IsDir() {
									walk.MsgBox(dlg, config.T("エラー", "Error"), config.T("選択されたファイルが読み取れないか、無効なパスです。", "The selected file is unreadable or is an invalid path."), walk.MsgBoxIconWarning)
									return
								}

								if initialPath != "" {
									confirmMsg := config.T("選択されたファイルは既に存在します。このファイルを証跡保存先として使用しますか？\n（既存の内容は維持されます）",
										"The selected file already exists. Do you want to use this file as the evidence destination?\n(Existing content will be preserved)")
									if walk.MsgBox(dlg, config.T("確認", "Confirmation"), confirmMsg, walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) == walk.DlgCmdNo {
										return
									}
								}

								config.CurrentConfig.ExcelOutputPath = fd.FilePath
								if err := config.SaveAppConfig(config.CurrentConfig); err != nil {
									walk.MsgBox(dlg, config.T("エラー", "Error"), config.T("設定の保存に失敗しました:\n", "Failed to save settings:\n")+err.Error(), walk.MsgBoxIconError)
									return
								}

								setupDone = true
								dlg.Accept()
							}
						},
					},
				},
			},

			// フッター（キャンセル）
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: config.T("キャンセル", "Cancel"),
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Create(nil)

	if err != nil {
		config.Log("ERROR", "初期設定ダイアログの作成に失敗しました: %v", "Failed to create Initial Setup Dialog: %v", err)
		return false
	}

	dlg.Run()

	if !setupDone {
		walk.MsgBox(nil, config.T("通知", "Notice"), config.T("必須設定が完了していないため、ツールを終了します。", "Tool will exit because required settings are not completed."), walk.MsgBoxIconInformation)
	}

	return setupDone
}

// IsValidExcelPath は指定されたパスが有効なExcelファイルかどうかをチェックします。
func IsValidExcelPath(path string) bool {
	if path == "" {
		return false
	}
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.IsDir() {
		return false
	}
	return true
}
