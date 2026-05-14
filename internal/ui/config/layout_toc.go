package uiconfig

import (
	"Evidence-Capture/internal/config"
	uicommon "Evidence-Capture/internal/ui/common"

	. "github.com/lxn/walk/declarative"
)

func (ui *ConfigUILayout) createTocTabPage() TabPage {
	c := &config.CurrentConfig

	return TabPage{
		Title:  config.T("目次設定", "TOC Settings"),
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, MarginsZero: true},
						Children: []Widget{
							uicommon.CheckBoxRow(config.T("目次シートを自動作成:", "Auto-create TOC Sheet:"),
								&ui.Toc.Enable,
								c.EnableAutoToc,
								ui.updateUILayoutControls),
						},
					},

					GroupBox{
						AssignTo: &ui.Toc.Group,
						Title:    config.T("詳細設定", "Detailed Settings"),
						Enabled:  c.EnableAutoToc,
						Layout:   Grid{Columns: 2},
						Children: []Widget{
							uicommon.LineEditRow(config.T("目次シートの名前:", "TOC Sheet Name:"), &ui.Toc.Name, c.TocSheetName, true, nil),
							Composite{
								Layout: Grid{Columns: 2, MarginsZero: true, SpacingZero: true},
								Children: []Widget{
									uicommon.CheckBoxWithEditRow(
										config.T("目次遷移リンク:", "TOC Navigation Link:"),
										&ui.Toc.LinkEnable,
										c.EnableTocLink,
										ui.updateUILayoutControls,
										config.T("戻り遷移先シート名:", "Return Sheet Name:"),
										&ui.Toc.BackText,
										c.ReturnButtonText,
										true,
										nil,
										100,
									),
								},
							},

							Composite{
								Layout: Grid{Columns: 2, MarginsZero: true, SpacingZero: true},
								Children: []Widget{
									uicommon.CheckBoxWithEditRow(
										config.T("ウィンドウ固定を有効:", "Enable Window Freeze:"),
										&ui.Toc.WinFix,
										c.EnableWindowFreeze,
										ui.updateUILayoutControls,
										config.T("固定するセル:", "Freeze Cell:"),
										&ui.Toc.FixCell,
										c.FreezePaneCell,
										c.EnableAutoToc && c.EnableWindowFreeze,
										func() {
											cellValidation(ui.Toc.FixCell)
										},
										50,
									),
								},
							},
							uicommon.NumberEditRow(config.T("新規シートの行の高さ（px）:", "New Sheet Row Height (px):"),
								&ui.Toc.RowHeight,
								c.DefaultRowHeight,
								1,
								409,
								0),
							uicommon.NumberEditRow(config.T("新規シートの列の幅（px）:", "New Sheet Column Width (px):"),
								&ui.Toc.ColWidth,
								c.DefaultColWidth,
								1,
								255,
								0),
							uicommon.NumberEditRow(config.T("表示倍率（％）:", "Zoom Level (%):"),
								&ui.Toc.Zoom,
								float64(c.ZoomPercent),
								10,
								400,
								0),
							uicommon.CheckBoxRow(config.T("枠線（目盛り線）を表示しない:", "Hide Gridlines:"),
								&ui.Toc.HideGrid,
								c.HideGridlines,
								nil),
						},
					},
					VSpacer{},
				},
			},
		},
	}
}
