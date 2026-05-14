package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"
	"Evidence-Capture/internal/ui/common"
	uiconfig "Evidence-Capture/internal/ui/config"
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

func RunIndexManager(owner walk.Form) {
	var dlg *walk.Dialog
	var createBtn, showMenuBtn *walk.PushButton

	exists := excel.IsIndexSheetExists()

	err := Dialog{
		AssignTo: &dlg,
		Title:    "インデックス管理",
		Icon:     2,
		MinSize:  Size{Width: 210, Height: 105},
		Layout:   VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}},
		Children: []Widget{
			PushButton{
				AssignTo: &createBtn,
				Text:     "インデックスシート作成",
				MinSize:  Size{Height: 50},
				Enabled:  !exists,
				OnClicked: func() {
					if err := excel.CreateIndexSheet(); err != nil {
						uicommon.ShowErrorWithLogPrompt(dlg, "エラー", fmt.Sprintf("インデックスシートの作成に失敗しました:\n%v", err))
					} else {
						updateButtonStates()
						createBtn.SetEnabled(false)
						showMenuBtn.SetEnabled(true)
						if statusBar != nil {
							statusBar.SetText("インデックスシートを作成しました。C4セルから下に入力してください。")
						}
					}
				},
			},
			PushButton{
				AssignTo: &showMenuBtn,
				Text:     "インデックスボタンメニューを表示",
				MinSize:  Size{Height: 50},
				Enabled:  exists,
				OnClicked: func() {
					RunIndexButtonManager(dlg)
				},
			},
			PushButton{
				Text:    "インデックス設定を変更",
				MinSize: Size{Height: 30},
				OnClicked: func() {
					uiconfig.RunConfigDialogLayout(dlg, func() {
						// 設定が保存された後に実行される
						newExists := excel.IsIndexSheetExists()
						createBtn.SetEnabled(!newExists)
						showMenuBtn.SetEnabled(newExists)
						updateButtonStates()
					}, 1) // 1 = インデックス設定タブ
				},
			},
			VSpacer{Size: 20},
			PushButton{
				Text:      "キャンセル",
				OnClicked: func() { dlg.Cancel() },
			},
		},
	}.Create(owner)

	if err != nil {
		config.Log("ERROR", "IndexManagerダイアログの作成に失敗しました。詳細: %v", "Failed to create IndexManager dialog. Detail: %v", err)
		return
	}

	if dlg != nil {
		win.SetWindowPos(dlg.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_SHOWWINDOW)
	}

	dlg.Run()
}
