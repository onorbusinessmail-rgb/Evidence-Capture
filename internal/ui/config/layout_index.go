package uiconfig

import (
	"Evidence-Capture/internal/config"
	uicommon "Evidence-Capture/internal/ui/common"
	"fmt"
	"regexp"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type StatusItem struct {
	ID      string
	Label   string
	Color   string
	Enabled string
	IsOn    bool
}

var (
	toggleUsageBtn *walk.PushButton
)

// walk.TableModelを実装するためのモデル
type StatusModel struct {
	walk.TableModelBase
	items []*StatusItem
	tv    *walk.TableView
}

func (m *StatusModel) RowCount() int {
	return len(m.items)
}

func (m *StatusModel) Value(row, col int) interface{} {
	item := m.items[row]
	switch col {
	case 0:
		return item.ID
	case 1:
		return item.Label
	case 2:
		return item.Color
	case 3:
		return "" // 背景色プレビュー用（Styleメソッドで色が付く列）
	case 4:
		if item.IsOn {
			return config.StatusLabelOn
		}
		return config.StatusLabelOff
	}
	return nil
}

func (ui *ConfigUILayout) createIndexTabPage() TabPage {
	c := &config.CurrentConfig

	items := []*StatusItem{}
	for _, raw := range c.StatusDefinitions {
		if raw == "" {
			continue
		}
		parts := strings.Split(raw, "|")
		if len(parts) >= 3 {
			isOn := false
			if len(parts) >= config.StatusPartsWithOn && parts[3] == config.StatusEnabled {
				isOn = true
			}
			items = append(items, &StatusItem{
				ID:    parts[0],
				Label: parts[1],
				Color: parts[2],
				IsOn:  isOn,
			})
		}
	}
	ui.statusModel = &StatusModel{items: items}

	var newID, newLabel, newColor *walk.LineEdit
	var tv *walk.TableView
	var addBtn, updateBtn, deleteBtn *walk.PushButton

	validateInput := func(id, label, color string, isUpdate bool) bool {
		if id == "" || label == "" || color == "" {
			walk.MsgBox(tv.Form(), config.T("入力エラー", "Input Error"), config.T("すべての項目を入力してください。", "Please fill in all fields."), walk.MsgBoxIconError)
			return false
		}
		currentIndex := tv.CurrentIndex()
		for i, item := range ui.statusModel.items {
			if isUpdate && i == currentIndex {
				continue
			}
			if item.ID == id {
				walk.MsgBox(tv.Form(), config.T("エラー", "Error"), config.T("このIDは既に登録されています。", "This ID is already registered."), walk.MsgBoxIconError)
				return false
			}
		}
		match, _ := regexp.MatchString(config.RegexHexColor6, color)
		if !match {
			res := walk.MsgBox(tv.Form(), config.T("確認", "Confirm"),
				config.T("背景色は通常 'RRGGBB' 形式（例: FF0000）で指定します。\n現在の入力値で登録してもよろしいですか？", "Background color is normally specified in 'RRGGBB' format (e.g. FF0000).\nDo you want to register with the current value?"),
				walk.MsgBoxIconWarning|walk.MsgBoxYesNo)
			if res == walk.DlgCmdNo {
				return false
			}
		}
		return true
	}

	return TabPage{
		Title:  config.T("インデックス設定", "Index Settings"),
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, MarginsZero: true},
						Children: []Widget{
							uicommon.CheckBoxRow(config.T("インデックスシートを作成:", "Create Index Sheet:"), &ui.Index.Enable, c.MakeIndex, ui.updateUILayoutControls),
						},
					},
					GroupBox{
						AssignTo: &ui.Index.Group,
						Title:    config.T("詳細設定", "Detailed Settings"),
						Enabled:  c.MakeIndex,
						Layout:   VBox{},
						Children: []Widget{
							Composite{
								Layout: Grid{Columns: 2, MarginsZero: true},
								Children: []Widget{
									uicommon.LineEditRow(config.T("インデックスシート名:", "Index Sheet Name:"),
										&ui.Index.Name,
										c.IndexSheetName,
										true,
										nil),

								},
							},

							// --- 入力エリア ---
							Composite{
								Layout: Grid{Columns: 4, MarginsZero: true},
								Children: []Widget{
									Label{Text: config.T("ID:", "ID:"), Alignment: AlignHFarVCenter},
									LineEdit{AssignTo: &newID},
									Label{Text: config.T("状態名:", "Status Label:"), Alignment: AlignHFarVCenter},
									LineEdit{AssignTo: &newLabel},
									Label{Text: config.T("背景色:", "BG Color:"), Alignment: AlignHFarVCenter},
									LineEdit{AssignTo: &newColor},
								},
							},

							// --- 操作ボタン群（UX最適化：追加、更新、削除、トグルを一箇所に集約） ---
							Composite{
								Layout: HBox{MarginsZero: true},
								Children: []Widget{
									PushButton{
										AssignTo: &addBtn,
										Text:     config.T("追加", "Add"),
										OnClicked: func() {
											id, label, color := newID.Text(), newLabel.Text(), newColor.Text()
											if !validateInput(id, label, color, false) {
												return
											}
											ui.statusModel.items = append(ui.statusModel.items, &StatusItem{
												ID:      id,
												Label:   label,
												Color:   color,
												Enabled: config.StatusLabelOn,
												IsOn:    true,
											})
											ui.statusModel.PublishRowsReset()
											newID.SetText("")
											newLabel.SetText("")
											newColor.SetText("")
										},
									},
									PushButton{
										AssignTo: &updateBtn,
										Text:     config.T("更新", "Update"),
										Enabled:  false,
										OnClicked: func() {
											idx := tv.CurrentIndex()
											if idx < 0 || idx >= len(ui.statusModel.items) {
												return
											}
											id, label, color := newID.Text(), newLabel.Text(), newColor.Text()
											if !validateInput(id, label, color, true) {
												return
											}
											ui.statusModel.items[idx].ID = id
											ui.statusModel.items[idx].Label = label
											ui.statusModel.items[idx].Color = color
											ui.statusModel.PublishRowsReset()
										},
									},
									PushButton{
										AssignTo: &deleteBtn,
										Text:     config.T("削除", "Delete"),
										Enabled:  false,
										OnClicked: func() {
											idx := tv.CurrentIndex()
											if idx < 0 || idx >= len(ui.statusModel.items) {
												return
											}
											if walk.MsgBox(tv.Form(), config.T("確認", "Confirm"), config.T("選択した行を削除しますか？", "Do you want to delete the selected row?"), walk.MsgBoxIconQuestion|walk.MsgBoxYesNo) == walk.DlgCmdNo {
												return
											}
											ui.statusModel.items = append(ui.statusModel.items[:idx], ui.statusModel.items[idx+1:]...)
											ui.statusModel.PublishRowsReset()
											newID.SetText("")
											newLabel.SetText("")
											newColor.SetText("")
											tv.SetCurrentIndex(-1)
										},
									},
									HSpacer{Size: 10},
									PushButton{
										AssignTo: &toggleUsageBtn,
										Text:     config.T("切り替え", "Toggle"),
										Enabled:  false,
										OnClicked: func() {
											idx := tv.CurrentIndex()
											if idx < 0 || idx >= len(ui.statusModel.items) {
												return
											}
											item := ui.statusModel.items[idx]
											item.IsOn = !item.IsOn // トグル
											ui.statusModel.PublishRowsReset()

											// ボタンテキストを即座に更新
											if item.IsOn {
												toggleUsageBtn.SetText(config.T("無効化する", "Deactivate"))
											} else {
												toggleUsageBtn.SetText(config.T("有効化する", "Activate"))
											}
											tv.Invalidate()
										},
									},
								},
							},

							Label{Text: config.T("ステータス定義一覧:", "Status Definitions:"), MinSize: Size{Width: 200}, Alignment: AlignHNearVCenter},
							TableView{
								AssignTo: &tv,
								Model:    ui.statusModel,
								Columns: []TableViewColumn{
									{Title: "ID", Width: 60, Alignment: AlignCenter},
									{Title: config.T("状態名", "Status Label"), Width: 100, Alignment: AlignCenter},
									{Title: config.T("背景色", "BG Color"), Width: 100, Alignment: AlignCenter},
									{Title: config.T("実際の色", "Actual Color"), Width: 120, Alignment: AlignCenter},
									{Title: config.T("使用有無", "Enabled"), Width: 100, Alignment: AlignCenter},
								},
								StyleCell: ui.statusModel.Style,
								MinSize:   Size{Height: 150},
								OnCurrentIndexChanged: func() {
									if ui.statusModel.tv == nil {
										ui.statusModel.tv = tv
									}
									idx := tv.CurrentIndex()
									if idx >= 0 && idx < len(ui.statusModel.items) {
										item := ui.statusModel.items[idx]
										newID.SetText(item.ID)
										newLabel.SetText(item.Label)
										newColor.SetText(item.Color)
										updateBtn.SetEnabled(true)
										deleteBtn.SetEnabled(true)
										toggleUsageBtn.SetEnabled(true)

										// 状態に応じてトグルボタンのテキストを変更
										if item.IsOn {
											toggleUsageBtn.SetText(config.T("無効化する", "Deactivate"))
										} else {
											toggleUsageBtn.SetText(config.T("有効化する", "Activate"))
										}
									} else {
										updateBtn.SetEnabled(false)
										deleteBtn.SetEnabled(false)
										toggleUsageBtn.SetEnabled(false)
										toggleUsageBtn.SetText(config.T("切り替え", "Toggle"))
									}
									// 再描画を促してStyleを反映させる
									tv.Invalidate()
								},
							},
						},
					},
					VSpacer{},
				},
			},
		},
	}
}

// StatusModel に Style メソッドを追加
func (m *StatusModel) Style(style *walk.CellStyle) {
	item := m.items[style.Row()]

	// 「実際の色のプレビュー」カラム（インデックス 3）の場合
	if style.Col() == 3 {
		if c, err := hexToWalkColor(item.Color); err == nil {
			style.BackgroundColor = c
			return // プレビュー列はハイライト（選択色）を無視して常に色を表示
		}
	}

	// 選択されている行かどうかを判定
	isSelected := m.tv != nil && m.tv.CurrentIndex() == style.Row()

	if !item.IsOn {
		if isSelected {
			// 選択中かつ無効状態の場合は、視認性を高めるため濃いめの色にするかリセット
			style.TextColor = walk.RGB(60, 60, 60)
		} else {
			style.TextColor = walk.RGB(160, 160, 160) // 通常の無効行はグレーアウト
		}
	} else if isSelected {
		// 有効行が選択されている場合も念のため色を保証
		style.TextColor = walk.RGB(0, 0, 0)
	}
}

// 16進数文字列 "E2F0D9" を walk.Color に変換するヘルパー
func hexToWalkColor(hex string) (walk.Color, error) {
	var r, g, b uint8
	n, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil || n != 3 {
		return 0, fmt.Errorf("invalid hex color")
	}
	return walk.RGB(r, g, b), nil
}
