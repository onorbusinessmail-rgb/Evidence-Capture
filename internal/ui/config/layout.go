package uiconfig

import (
	"Evidence-Capture/internal/config"
	uicommon "Evidence-Capture/internal/ui/common"
	"fmt"
	"regexp"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
	toggleOnBtn  *walk.PushButton
	toggleOffBtn *walk.PushButton
)

var (
	lightGrayColor = walk.RGB(240, 240, 240)
)

func cellValidation(le *walk.LineEdit) {
	if le == nil {
		return
	}
	// Excelセル形式の正規表現（A～Zで始まり、1以上の数字が続く）
	match, _ := regexp.MatchString(`^[a-zA-Z]+[1-9][0-9]*$`, le.Text())

	if match || le.Text() == "" {
		le.SetBackground(uicommon.WhiteBrush)
	} else {
		le.SetBackground(uicommon.ErrorBrush)
	}
}

type ConfigUILayout struct {
	Toc struct {
		Enable     *walk.CheckBox
		Name       *walk.LineEdit
		LinkEnable *walk.CheckBox
		BackText   *walk.LineEdit
		WinFix     *walk.CheckBox
		FixCell    *walk.LineEdit
		RowHeight  *walk.NumberEdit
		ColWidth   *walk.NumberEdit
		Zoom       *walk.NumberEdit
		HideGrid   *walk.CheckBox
		PageBreak  *walk.CheckBox
		WrapText   *walk.CheckBox
		CenterV    *walk.CheckBox
		Group      *walk.GroupBox
	}
	Index struct {
		Enable        *walk.CheckBox
		Name          *walk.LineEdit
		StatusPreview *walk.TextEdit
		Group         *walk.GroupBox
	}
	Image struct {
		TimestampOn *walk.CheckBox
		TimeFormat  *walk.ComboBox
		TimeCell    *walk.LineEdit
		StartCell   *walk.LineEdit
		Margin      *walk.NumberEdit
		Scale       *walk.NumberEdit
		Group       *walk.GroupBox
	}
	Local struct {
		SaveOn    *walk.CheckBox
		Format    *walk.ComboBox
		Folder    *walk.LineEdit
		PasteMode *walk.ComboBox

		Group *walk.GroupBox
	}
	Detail struct {
		Warn      *walk.CheckBox
		LimitCnt  *walk.NumberEdit
		LimitSize *walk.NumberEdit
		LimitDay  *walk.NumberEdit
		CompTime  *walk.ComboBox
		CompTool  *walk.ComboBox
		CmdPath   *walk.LineEdit
		CmdArgs   *walk.LineEdit
		PassOn    *walk.CheckBox
		PassText  *walk.LineEdit
		AutoDel   *walk.CheckBox
		DelDay    *walk.NumberEdit
		DiskWarn  *walk.NumberEdit
		Group     *walk.GroupBox
	}
	statusModel *StatusModel
	TabWidget   *walk.TabWidget
}

// 構造体の定義
type ToolItem struct {
	Id   int
	Name string
}

func (ui *ConfigUILayout) updateUILayoutControls() {
	if ui.Toc.Enable == nil || ui.Toc.Group == nil {
		return
	}
	// 親スイッチの状態をグループ全体に反映
	parentOn := ui.Toc.Enable.Checked()
	ui.Toc.Group.SetEnabled(parentOn)

	if parentOn {
		// 目次遷移リンクのON/OFFに合わせて「戻りボタン表示文字」を制御
		if ui.Toc.LinkEnable != nil && ui.Toc.BackText != nil {
			ui.Toc.BackText.SetEnabled(ui.Toc.LinkEnable.Checked())
		}

		// ウィンドウ固定のON/OFFに合わせて「固定するセル」を制御
		if ui.Toc.WinFix != nil && ui.Toc.FixCell != nil {
			ui.Toc.FixCell.SetEnabled(ui.Toc.WinFix.Checked())
		}
	}

	if ui.Index.Enable != nil && ui.Index.Group != nil {
		ui.Index.Group.SetEnabled(ui.Index.Enable.Checked())
	}

	if ui.Image.TimestampOn != nil && ui.Image.Group != nil {
		ui.Image.Group.SetEnabled(ui.Image.TimestampOn.Checked())
		if ui.Image.TimeFormat != nil {
			ui.Image.TimeFormat.SetEnabled(ui.Image.TimestampOn.Checked())
		}
		if ui.Image.TimeCell != nil {
			ui.Image.TimeCell.SetEnabled(ui.Image.TimestampOn.Checked())
		}
	}

	if ui.Local.SaveOn != nil && ui.Local.Group != nil {
		ui.Local.Group.SetEnabled(ui.Local.SaveOn.Checked())
	}

	if ui.Detail.Warn != nil && ui.Detail.Group != nil {
		ui.Detail.Group.SetEnabled(ui.Detail.Warn.Checked())
		if ui.Detail.PassOn != nil && ui.Detail.PassText != nil {
			ui.Detail.PassText.SetEnabled(ui.Detail.Warn.Checked() && ui.Detail.PassOn.Checked())
		}
		if ui.Detail.AutoDel != nil && ui.Detail.DelDay != nil {
			ui.Detail.DelDay.SetEnabled(ui.Detail.Warn.Checked() && ui.Detail.AutoDel.Checked())
		}
	}
}

// 全項目の入力チェックを一括で行う
func (ui *ConfigUILayout) validateLayoutInputs(dlg *walk.Dialog) bool {
	if ui.Local.SaveOn == nil || ui.Local.Folder == nil {
		return true
	}

	// 「画像をパソコン内に保存」が ON の場合のみチェック
	if ui.Local.SaveOn.Checked() {
		// 共通バリデーションエンジンで必須入力およびフォルダ存在チェックを一括実行
		if !uicommon.ValidateUIInputs(dlg, uicommon.ValidationItem{
			Widget:   ui.Local.Folder,
			Msg:      "指定された「保存先フォルダ」が正しくありません。\n実在するフォルダパスを指定してください。",
			Required: true,
			IsFolder: true,
		}) {
			// バリデーション失敗時に該当タブ（インデックス3: ローカル保存）へ切り替え
			if ui.TabWidget != nil {
				ui.TabWidget.SetCurrentIndex(3)
			}
			return false
		}
	} else {
		// チェックOFF時は背景色をリセット
		ui.Local.Folder.SetBackground(uicommon.WhiteBrush)
	}

	return true
}

// 設定ダイアログの保存処理
func (ui *ConfigUILayout) saveConfigLayoutSettings(dlg *walk.Dialog) bool {
	// 保存前にバリデーションを実行
	if !ui.validateLayoutInputs(dlg) {
		return false
	}

	newCfg := config.CurrentConfig

	// 1. 目次設定
	newCfg.EnableAutoToc = ui.Toc.Enable.Checked()
	newCfg.TocSheetName = ui.Toc.Name.Text()
	newCfg.EnableTocLink = ui.Toc.LinkEnable.Checked()
	newCfg.ReturnButtonText = ui.Toc.BackText.Text()
	newCfg.EnableWindowFreeze = ui.Toc.WinFix.Checked()
	newCfg.FreezePaneCell = ui.Toc.FixCell.Text()
	newCfg.DefaultRowHeight = ui.Toc.RowHeight.Value()
	newCfg.DefaultColWidth = ui.Toc.ColWidth.Value()
	newCfg.ZoomPercent = int(ui.Toc.Zoom.Value())
	newCfg.HideGridlines = ui.Toc.HideGrid.Checked()

	// 2. インデックス設定
	newCfg.MakeIndex = ui.Index.Enable.Checked()
	newCfg.IndexSheetName = ui.Index.Name.Text()


	newCfg.StatusDefinitions = []string{}
	if ui.statusModel != nil {
		for _, item := range ui.statusModel.items {
			if item.ID == "" && item.Label == "" {
				continue
			}
			statusStr := "OFF"
			if item.IsOn {
				statusStr = "ON"
			}
			val := fmt.Sprintf("%s|%s|%s|%s", item.ID, item.Label, item.Color, statusStr)
			newCfg.StatusDefinitions = append(newCfg.StatusDefinitions, val)
		}
	}

	// 3. 画像挿入設定
	newCfg.EnableTimestamp = ui.Image.TimestampOn.Checked()
	newCfg.TimestampCell = ui.Image.TimeCell.Text()
	newCfg.ImageInsertStartCell = ui.Image.StartCell.Text()
	newCfg.Margin = int(ui.Image.Margin.Value())
	newCfg.ImageScale = ui.Image.Scale.Value()
	newCfg.TimestampFormat = uicommon.GetComboValue(ui.Image.TimeFormat)

	// 4. ローカル保存設定
	newCfg.LocalSaveOn = ui.Local.SaveOn.Checked()
	newCfg.LocalSaveFolder = ui.Local.Folder.Text()
	newCfg.LocalSaveFormat = uicommon.GetComboValue(ui.Local.Format)
	newCfg.PasteMode = uicommon.GetComboValue(ui.Local.PasteMode)

	// 5. 保存詳細設定 (UIが有効な場合のみ)
	if ui.Detail.Warn != nil {
		newCfg.AutoOrganizeWarn = ui.Detail.Warn.Checked()
		newCfg.LimitCount = int(ui.Detail.LimitCnt.Value())
		newCfg.LimitSizeMB = int(ui.Detail.LimitSize.Value())
		newCfg.LimitDays = int(ui.Detail.LimitDay.Value())
		newCfg.CompTiming = uicommon.GetComboValue(ui.Detail.CompTime)
		newCfg.CompTool = uicommon.GetComboValue(ui.Detail.CompTool)
		newCfg.EnableCompPass = ui.Detail.PassOn.Checked()
		newCfg.CompPass = ui.Detail.PassText.Text()
		newCfg.EnableAutoDelZip = ui.Detail.AutoDel.Checked()
		newCfg.DelZipDays = int(ui.Detail.DelDay.Value())
		newCfg.DiskFreeThresholdGB = ui.Detail.DiskWarn.Value()
	}

	// 保存実行
	if err := config.SaveAppConfig(newCfg); err != nil {
		walk.MsgBox(dlg, "保存失敗", fmt.Sprintf("設定の保存に失敗しました:\n%v", err), walk.MsgBoxIconError)
		return false
	}

	return true
}

// 設定ダイアログのフッター（初期化、保存、キャンセルボタン）
func RunConfigDialogLayout(owner walk.Form, onAccept func(), initialTabIndex int) {
	ui := new(ConfigUILayout)

	uicommon.RunBaseConfigDialog(owner, uicommon.ConfigParams{
		Title:      config.T("設定2：Excelレイアウト設定", "Settings 2: Excel Layout"),
		MinSize:    Size{Width: 560, Height: 400},
		IniPath:    config.IniFile2,
		DefaultIni: config.DefaultIni2,
		GetPages: func(dlg *walk.Dialog) []TabPage {
			return []TabPage{
				ui.createTocTabPage(),
				ui.createIndexTabPage(),
				ui.createImageTabPage(),
				ui.createLocalSaveTabPage(),
			}
		},
		OnSave: func(dlg *walk.Dialog) bool {
			return ui.saveConfigLayoutSettings(dlg)
		},
		OnSizeChanged: func() {
			ui.updateUILayoutControls()
		},
		InitialTabIndex: initialTabIndex,
		TabWidget:       &ui.TabWidget,
	}, onAccept)
}
