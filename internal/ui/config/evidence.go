package uiconfig

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/winapi"
	"Evidence-Capture/internal/ui/common"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type ConfigUI struct {
	Base struct {
		TriggerTopLeft           *walk.CheckBox
		TriggerTopRight          *walk.CheckBox
		TriggerRightDbl          *walk.CheckBox
		ExitCtrlShiftX           *walk.CheckBox
		ExitWinKey               *walk.CheckBox
		ExitLeftDbl              *walk.CheckBox
		ExitTimeout              *walk.CheckBox
		WaitTimeSeconds          *walk.NumberEdit
		EdgeSensitivity          *walk.NumberEdit
		AutoSwitchTimeoutMinutes *walk.NumberEdit
		EditModeBehavior         *walk.ComboBox
		DefaultCaptureRange      *walk.ComboBox
		ExcelOutputPath          *walk.LineEdit
		MinimizeExcel            *walk.CheckBox
		TopMostTOC               *walk.CheckBox
		EnableBeepSound          *walk.CheckBox
		EnableMutex              *walk.ComboBox
		AutoSaveExcel            *walk.ComboBox
		EnableLogOutput          *walk.CheckBox
		LogOutputPath            *walk.LineEdit
		LogOutputPathBtn         *walk.PushButton
		LogMaxSizeMB             *walk.NumberEdit
		EnableFullWindowCapture  *walk.CheckBox
	}
	TabWidget *walk.TabWidget
}

func (ui *ConfigUI) saveConfigEvidenceSettings(dlg *walk.Dialog) bool {
	if !uicommon.ValidateUIInputs(dlg,
		uicommon.ValidationItem{
			Widget:   ui.Base.ExcelOutputPath,
			Msg:      "証跡格納先Excelファイルパスは必須です",
			Required: true,
		},
		uicommon.ValidationItem{
			Widget:   ui.Base.EdgeSensitivity,
			Msg:      "感度は 0.0 から 1.0 の範囲で入力してください。",
			IsNumber: true,
			MinVal:   0.0,
			MaxVal:   1.0,
		},
		uicommon.ValidationItem{
			Widget:   ui.Base.WaitTimeSeconds,
			Msg:      "待機時間は 0.1 から 60.0 秒の間で入力してください。",
			IsNumber: true,
			MinVal:   0.1,
			MaxVal:   60.0,
		},
	) {
		// バリデーション失敗時に該当タブ（インデックス2: Excel・ファイル）へ切り替え
		// ※基本設定項目の多くがここにあるため。必要に応じて個別判定も可能。
		if ui.TabWidget != nil {
			ui.TabWidget.SetCurrentIndex(2)
		}
		return false
	}

	newCfg := config.CurrentConfig

	newCfg.EnableTriggerTopLeft = ui.Base.TriggerTopLeft.Checked()
	newCfg.EnableTriggerTopRight = ui.Base.TriggerTopRight.Checked()
	newCfg.EnableTriggerRightDbl = ui.Base.TriggerRightDbl.Checked()
	newCfg.EnableTopMostTOC = ui.Base.TopMostTOC.Checked()
	newCfg.EnableExitLeftDbl = ui.Base.ExitLeftDbl.Checked()
	newCfg.EnableExitShortcut = ui.Base.ExitCtrlShiftX.Checked()
	newCfg.EnableExitWinKey = ui.Base.ExitWinKey.Checked()
	newCfg.EnableAutoTimeout = ui.Base.ExitTimeout.Checked()
	newCfg.AutoSwitchTimeoutMinutes = int(ui.Base.AutoSwitchTimeoutMinutes.Value())
	newCfg.MinimizeExcel = ui.Base.MinimizeExcel.Checked()
	newCfg.EnableBeep = ui.Base.EnableBeepSound.Checked()
	newCfg.WaitTimeSeconds = ui.Base.WaitTimeSeconds.Value()
	newCfg.EdgeSensitivity = ui.Base.EdgeSensitivity.Value()
	newCfg.EnableMutex = uicommon.GetComboValue(ui.Base.EnableMutex)
	newCfg.AutoSaveExcel = uicommon.GetComboValue(ui.Base.AutoSaveExcel)
	newCfg.EditModeBehavior = uicommon.GetComboValue(ui.Base.EditModeBehavior)
	newCfg.DefaultCaptureRange = uicommon.GetComboValue(ui.Base.DefaultCaptureRange)
	newCfg.ExcelOutputPath = ui.Base.ExcelOutputPath.Text()
	newCfg.EnableLogOutput = ui.Base.EnableLogOutput.Checked()
	newCfg.LogOutputPath = ui.Base.LogOutputPath.Text()
	newCfg.LogMaxSizeMB = int(ui.Base.LogMaxSizeMB.Value())
	newCfg.EnableFullWindowCapture = ui.Base.EnableFullWindowCapture.Checked()

	if err := config.SaveAppConfig(newCfg); err != nil {
		config.Log("ERROR", "設定の保存に失敗しました。詳細: %v", "Failed to save settings: %v", err)
		winapi.ShowDialog("エラー", "設定の保存に失敗しました。", winapi.MB_ICONERROR)
		return false
	}

	config.InitLogger(newCfg)
	return true
}

func RunConfigDialogEvidence(owner walk.Form, onAccept func()) {
	ui := new(ConfigUI)

	uicommon.RunBaseConfigDialog(owner, uicommon.ConfigParams{
		Title:      config.T("設定1：証跡・システム設定", "Settings 1: Evidence & System"),
		MinSize:    Size{Width: 600, Height: 450},
		IniPath:    config.IniFile1,
		DefaultIni: config.DefaultIni1,
		GetPages: func(dlg *walk.Dialog) []TabPage {
			return ui.createBaseTabPages(dlg)
		},
		OnSave: func(dlg *walk.Dialog) bool {
			return ui.saveConfigEvidenceSettings(dlg)
		},
		TabWidget: &ui.TabWidget,
	}, onAccept)
}

func (ui *ConfigUI) createBaseTabPages(_ *walk.Dialog) []TabPage {
	c := &config.CurrentConfig

	stampItems := []types.ComboItem{
		{Name: "1: ユーザー通知・警告", Value: 1},
	}

	captureRangeItems := []types.ComboItem{
		{Name: "1: アクティブウィンドウ", Value: 1},
	}

	mutexItems := []types.ComboItem{
		{Name: "1: 証跡格納先Excelファイルのみ強制終了", Value: 1},
		{Name: "2: 全てのExcelファイルを強制終了", Value: 2},
		{Name: "3: Excelファイルの強制終了しない", Value: 3},
	}
	autoSaveItems := []types.ComboItem{
		{Name: "1: 無効（手動で保存）", Value: 1},
		{Name: "2: 有効（撮影ごとに上書き保存）", Value: 2},
	}

	updateTimeoutUI := func() {
		if ui.Base.ExitTimeout == nil || ui.Base.AutoSwitchTimeoutMinutes == nil {
			return
		}
		ui.Base.AutoSwitchTimeoutMinutes.SetEnabled(ui.Base.ExitTimeout.Checked())
	}

	updateCaptureUI := func() {
		if ui.Base.DefaultCaptureRange == nil || ui.Base.EnableFullWindowCapture == nil {
			return
		}
		isWindowMode := uicommon.GetComboValue(ui.Base.DefaultCaptureRange) == 1
		ui.Base.EnableFullWindowCapture.SetEnabled(isWindowMode)
		if !isWindowMode {
			ui.Base.EnableFullWindowCapture.SetChecked(false)
		}
	}

	updateLogUI := func() {
		if ui.Base.EnableLogOutput == nil || ui.Base.LogOutputPath == nil || ui.Base.LogMaxSizeMB == nil || ui.Base.LogOutputPathBtn == nil {
			return
		}
		enabled := ui.Base.EnableLogOutput.Checked()
		ui.Base.LogOutputPath.SetEnabled(enabled)
		ui.Base.LogMaxSizeMB.SetEnabled(enabled)
		ui.Base.LogOutputPathBtn.SetEnabled(enabled)
	}

	return []TabPage{
		{
			Title:  config.T("トリガー", "Triggers"),
			Layout: VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
			Children: []Widget{
				Composite{
					Layout: HBox{MarginsZero: true},
					Children: []Widget{
						GroupBox{
							Title:  config.T("撮影トリガー設定", "Capture Triggers"),
							Layout: VBox{Margins: Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}},
							Children: []Widget{
								uicommon.CheckBoxRow(config.T("画面左上へのマウス移動", "Mouse to Top-Left"), &ui.Base.TriggerTopLeft, c.EnableTriggerTopLeft, nil),
								uicommon.CheckBoxRow(config.T("画面右上へのマウス移動", "Mouse to Top-Right"), &ui.Base.TriggerTopRight, c.EnableTriggerTopRight, nil),
								uicommon.CheckBoxRow(config.T("右ダブルクリック", "Right Double-Click"), &ui.Base.TriggerRightDbl, c.EnableTriggerRightDbl, nil),
								VSpacer{},
							},
						},
						GroupBox{
							Title:  config.T("終了・中断トリガー設定", "Exit/Interrupt Triggers"),
							Layout: VBox{Margins: Margins{Left: 5, Top: 5, Right: 5, Bottom: 5}},
							Children: []Widget{
								uicommon.CheckBoxRow(config.T("画面左端でのダブルクリック", "Double-Click at Left Edge"), &ui.Base.ExitLeftDbl, c.EnableExitLeftDbl, nil),
								uicommon.CheckBoxRow(config.T("Ctrl+Shift+Xキー", "Ctrl+Shift+X Key"), &ui.Base.ExitCtrlShiftX, c.EnableExitShortcut, nil),
								uicommon.CheckBoxRow(config.T("Windowsキー", "Windows Key"), &ui.Base.ExitWinKey, c.EnableExitWinKey, nil),
								uicommon.CheckBoxRow(config.T("タイムアウト監視", "Timeout Monitoring"), &ui.Base.ExitTimeout, c.EnableAutoTimeout, updateTimeoutUI),
								uicommon.NumberEditRow(config.T("タイムアウト時間(分):", "Timeout (min):"), &ui.Base.AutoSwitchTimeoutMinutes, float64(c.AutoSwitchTimeoutMinutes), 1, 1440, 0),
							},
						},
					},
				},
				VSpacer{},
			},
		},
		{
			Title:  config.T("撮影・環境", "Capture & Environment"),
			Layout: VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
			Children: []Widget{
				GroupBox{
					Title:  config.T("キャプチャ・詳細設定", "Capture Details"),
					Layout: Grid{Columns: 1},
					Children: []Widget{
						uicommon.NumberEditRow(config.T("撮影後待機時間(秒):", "Wait Time After Capture (sec):"), &ui.Base.WaitTimeSeconds, c.WaitTimeSeconds, 0.5, 60, 1),
						uicommon.ComboBoxRowCustom(config.T("標準撮影範囲:", "Default Capture Range:"), &ui.Base.DefaultCaptureRange, &types.ComboModel{Items: captureRangeItems}, c.DefaultCaptureRange-1, updateCaptureUI),
						uicommon.CheckBoxRowEnabled(config.T("アクティブウィンドウ撮影時：画面外・裏側も完全に取得する", "Capture full window even if obscured"), &ui.Base.EnableFullWindowCapture, c.EnableFullWindowCapture, c.DefaultCaptureRange == 1, nil),
						uicommon.NumberEditRow(config.T("画面端判定感度:", "Edge Sensitivity:"), &ui.Base.EdgeSensitivity, c.EdgeSensitivity, 0.001, 0.1, 3),
						uicommon.CheckBoxRow(config.T("通知音を鳴らす:", "Enable Notification Sound:"), &ui.Base.EnableBeepSound, c.EnableBeep, nil),
					},
				},
				VSpacer{},
			},
		},
		{
			Title:  config.T("Excel・ファイル", "Excel & Files"),
			Layout: VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
			Children: []Widget{
				Composite{
					Layout:          VBox{MarginsZero: true},
					OnBoundsChanged: func() { updateCaptureUI(); updateLogUI() },
					Children: []Widget{
						GroupBox{
							Title:  config.T("出力・制御設定", "Output & Control"),
							Layout: Grid{Columns: 2},
							Children: []Widget{
								Label{Text: config.T("証跡格納先Excelファイル:", "Evidence Excel File:"), Alignment: AlignHFarVCenter},
								Composite{
									Layout: HBox{MarginsZero: true},
									Children: []Widget{
										LineEdit{AssignTo: &ui.Base.ExcelOutputPath, Text: c.ExcelOutputPath},
										PushButton{
											Text: config.T("参照", "Browse"),
											OnClicked: func() {
												le := ui.Base.ExcelOutputPath
												dlg := new(walk.FileDialog)
												if ok, _ := dlg.ShowOpen(le.Form()); ok {
													le.SetText(dlg.FilePath)
												}
											},
										},
									},
								},

								Label{Text: config.T("編集モード検知動作:", "Edit Mode Detection:"), Alignment: AlignHFarVCenter},
								ComboBox{
									AssignTo:      &ui.Base.EditModeBehavior,
									Model:         &types.ComboModel{Items: stampItems},
									DisplayMember: "Name",
									CurrentIndex:  c.EditModeBehavior - 1,
								},

								Label{Text: config.T("二重起動チェック:", "Double Launch Check:"), Alignment: AlignHFarVCenter},
								ComboBox{
									AssignTo:      &ui.Base.EnableMutex,
									Model:         &types.ComboModel{Items: mutexItems},
									DisplayMember: "Name",
									CurrentIndex:  c.EnableMutex - 1,
								},

								Label{Text: config.T("撮影ごと自動保存:", "Auto Save Each Capture:"), Alignment: AlignHFarVCenter},
								ComboBox{
									AssignTo:      &ui.Base.AutoSaveExcel,
									Model:         &types.ComboModel{Items: autoSaveItems},
									DisplayMember: "Name",
									CurrentIndex:  c.AutoSaveExcel - 1,
								},

								// チェックボックス群の2カラム化
								Composite{
									Layout:     Grid{Columns: 4, MarginsZero: true, SpacingZero: true},
									ColumnSpan: 2,
									Children: []Widget{
										Label{Text: config.T("Excel最小化:", "Minimize Excel:"), Alignment: AlignHFarVCenter, MinSize: Size{Width: 150}},
										CheckBox{AssignTo: &ui.Base.MinimizeExcel, Checked: c.MinimizeExcel},
										Label{Text: config.T("目次: 最前面:", "TOC: Always on Top:"), Alignment: AlignHFarVCenter, MinSize: Size{Width: 150}},
										CheckBox{AssignTo: &ui.Base.TopMostTOC, Checked: c.EnableTopMostTOC},

										Label{Text: config.T("ログ出力有効:", "Enable Log Output:"), Alignment: AlignHFarVCenter, MinSize: Size{Width: 150}},
										CheckBox{AssignTo: &ui.Base.EnableLogOutput, Checked: c.EnableLogOutput, OnClicked: updateLogUI},
										// 空セル（列合わせ用）
										Label{Text: ""},
										Label{Text: ""},
									},
								},

								Label{Text: config.T("ログ出力フォルダ:", "Log Output Folder:"), Alignment: AlignHFarVCenter},
								Composite{
									Layout: HBox{MarginsZero: true},
									Children: []Widget{
										LineEdit{AssignTo: &ui.Base.LogOutputPath, Text: c.LogOutputPath},
										PushButton{
											AssignTo: &ui.Base.LogOutputPathBtn,
											Text:     config.T("参照", "Browse"),
											OnClicked: func() {
												le := ui.Base.LogOutputPath
												dlg := new(walk.FileDialog)
												if ok, _ := dlg.ShowBrowseFolder(le.Form()); ok {
													le.SetText(dlg.FilePath)
												}
											},
										},
									},
								},

								Label{Text: config.T("ログ警告サイズ(MB):", "Log Warn Size (MB):"), Alignment: AlignHFarVCenter},
								NumberEdit{
									AssignTo: &ui.Base.LogMaxSizeMB,
									Value:    float64(c.LogMaxSizeMB),
									MinValue: 1,
									MaxValue: 10000,
									Decimals: 0,
								},
							},
						},
						VSpacer{},
					},
				},
			},
		},
	}
}
