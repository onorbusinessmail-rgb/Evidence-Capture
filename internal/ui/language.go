package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/process"
	"Evidence-Capture/internal/winapi"

	"github.com/go-ole/go-ole"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// RunLanguageDialog は言語設定ダイアログを表示します。
func RunLanguageDialog(owner walk.Form) {
	var dlg *walk.Dialog
	var cmb *walk.ComboBox

	// 現在の言語を取得
	currentLang := config.CurrentConfig.Language
	if currentLang == "" {
		currentLang = "Japanese"
	}

	model := []string{"Japanese", "English"}

	Dialog{
		AssignTo:      &dlg,
		Title:         "言語設定 / Language Settings",
		Icon:          2,
		MinSize:       Size{Width: 300, Height: 150},
		Layout:        VBox{},
		DefaultButton: nil,
		Children: []Widget{
			Label{
				Text: config.T("表示言語を選択してください:", "Select display language:"),
			},
			ComboBox{
				AssignTo: &cmb,
				Model:    model,
				Value:    currentLang,
			},
			VSpacer{},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: config.T("保存", "Save"),
						OnClicked: func() {
							selected := cmb.Text()
							config.CurrentConfig.Language = selected
							if err := config.SaveAppConfig(config.CurrentConfig); err != nil {
								walk.MsgBox(dlg, config.T("エラー", "Error"), config.T("保存に失敗しました:\n", "Failed to save:\n")+err.Error(), walk.MsgBoxIconError)
								return
							}

							dlg.Accept()
							if mw != nil {
								winapi.ShowDialog(
									config.T("自動再起動", "Auto Restart"),
									config.T("言語設定を変更しました。設定を適用するため、ツールを自動再起動します。", "Language settings updated. The application will now restart automatically to apply the changes."),
									winapi.MB_ICONINFORMATION,
								)
								ole.CoUninitialize()
								process.RestartSelf()
							}
						},
					},
					PushButton{
						Text: config.T("キャンセル", "Cancel"),
						OnClicked: func() {
							dlg.Cancel()
						},
					},
				},
			},
		},
	}.Run(owner)
}
