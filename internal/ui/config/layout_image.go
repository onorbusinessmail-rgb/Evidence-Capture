package uiconfig

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/ui/common"

	. "github.com/lxn/walk/declarative"
)

func (ui *ConfigUILayout) createImageTabPage() TabPage {
	c := &config.CurrentConfig

	stampItems := []types.ComboItem{
		{Name: config.T("1: スラッシュ (2023/01/01)", "1: Slash (2023/01/01)"), Value: 1},
		{Name: config.T("2: ハイフン (2023-01-01)", "2: Hyphen (2023-01-01)"), Value: 2},
		{Name: config.T("3: 日本語 (2023年01月01日)", "3: Japanese (2023Y01M01D)"), Value: 3},
	}

	return TabPage{
		Title:  config.T("挿入設定", "Insertion Settings"),
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, MarginsZero: true},
						Children: []Widget{
							uicommon.CheckBoxRow(config.T("タイムスタンプを記入:", "Insert Timestamp:"),
								&ui.Image.TimestampOn,
								c.EnableTimestamp,
								ui.updateUILayoutControls),
						},
					},
					GroupBox{
						AssignTo: &ui.Image.Group,
						Title:    config.T("詳細設定", "Detailed Settings"),
						Enabled:  c.EnableTimestamp,
						Layout:   Grid{Columns: 2},
						Children: []Widget{
							uicommon.ComboBoxRow(config.T("タイムスタンプ:書式選択:", "Timestamp Format:"),
								&ui.Image.TimeFormat,
								&types.ComboModel{Items: stampItems},
								c.TimestampFormat-1),
							uicommon.LineEditRow(config.T("タイムスタンプ:初期記入セル:", "Timestamp: Initial Cell:"),
								&ui.Image.TimeCell, c.TimestampCell,
								true,
								func() { cellValidation(ui.Image.TimeCell) }),
							uicommon.LineEditRow(config.T("画像:初期貼り付けセル:", "Image: Initial Cell:"),
								&ui.Image.StartCell,
								c.ImageInsertStartCell,
								true,
								func() { cellValidation(ui.Image.StartCell) }),
							uicommon.NumberEditRow(config.T("画像:上下余白:", "Image: Top/Bottom Margin:"),
								&ui.Image.Margin,
								float64(c.Margin),
								0,
								100,
								0),
							uicommon.NumberEditRow(config.T("画像:縮小率:", "Image: Scale Ratio:"),
								&ui.Image.Scale,
								c.ImageScale,
								0.1,
								1.0,
								2),
						},
					},
					VSpacer{},
				},
			},
		},
	}
}
