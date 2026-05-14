package uiconfig

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/ui/common"
	"fmt"

	. "github.com/lxn/walk/declarative"
)

func (ui *ConfigUILayout) createLocalDetailTabPage() TabPage {
	c := &config.CurrentConfig
	timeItems := []types.ComboItem{
		{Name: config.T("1: 日次", "1: Daily"), Value: 1},
		{Name: config.T("2: 週次", "2: Weekly"), Value: 2},
		{Name: config.T("3: 2週次", "3: Bi-weekly"), Value: 3},
		{Name: config.T("4: 月次", "4: Monthly"), Value: 4},
		{Name: config.T("5: 3月次", "5: Quarterly"), Value: 5},
	}
	toolItems := []types.ComboItem{
		{Name: config.T("1: 標準ZIP", "1: Standard ZIP"), Value: 1},
		{Name: config.T("2: 7-Zip", "2: 7-Zip"), Value: 2},
		{Name: config.T("3: 外部コマンド", "3: External Command"), Value: 3},
	}

	return TabPage{
		Title:  config.T("保存詳細", "Save Details"),
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, MarginsZero: true},
						Children: []Widget{
							uicommon.CheckBoxRow(config.T("自動整理警告を表示:", "Show Auto-Organize Warning:"),
								&ui.Detail.Warn,
								c.AutoOrganizeWarn,
								ui.updateUILayoutControls),
						},
					},
					GroupBox{
						AssignTo: &ui.Detail.Group,
						Title:    config.T("詳細設定", "Detailed Settings"),
						Enabled:  c.AutoOrganizeWarn,
						Layout:   Grid{Columns: 2},
						Children: []Widget{
							uicommon.NumberEditRow(config.T("制限枚数:", "Limit Count:"),
								&ui.Detail.LimitCnt,
								float64(c.LimitCount),
								1,
								10000,
								0),
							uicommon.NumberEditRow(config.T("制限容量MB:", "Limit Size MB:"),
								&ui.Detail.LimitSize,
								float64(c.LimitSizeMB),
								1,
								10000,
								0),
							uicommon.NumberEditRow(config.T("制限日数:", "Limit Days:"),
								&ui.Detail.LimitDay,
								float64(c.LimitDays),
								1,
								365,
								0),
							uicommon.ComboBoxRow(config.T("圧縮実行タイミング:", "Compression Timing:"),
								&ui.Detail.CompTime,
								&types.ComboModel{Items: timeItems},
								c.CompTiming-1),
							uicommon.ComboBoxRow(config.T("圧縮ツール選択:", "Compression Tool:"),
								&ui.Detail.CompTool,
								&types.ComboModel{Items: toolItems}, c.CompTool-1),
							uicommon.CheckBoxRow(config.T("画像圧縮パスワード設定:", "Enable Compression Password:"),
								&ui.Detail.PassOn,
								c.EnableCompPass,
								ui.updateUILayoutControls),
							uicommon.LineEditRow(config.T("圧縮パスワード:", "Compression Password:"),
								&ui.Detail.PassText,
								c.CompPass,
								true,
								nil),
							uicommon.CheckBoxRow(config.T("画像圧縮ファイル自動削除:", "Auto-delete Compressed Files:"),
								&ui.Detail.AutoDel,
								c.EnableAutoDelZip,
								ui.updateUILayoutControls),
							uicommon.NumberEditRow(config.T("経過日数:", "Elapsed Days:"),
								&ui.Detail.DelDay,
								float64(c.DelZipDays),
								1,
								365,
								0),
							uicommon.NumberEditRow(
								config.T(fmt.Sprintf("空き容量警告閾値(GB) [現在: %.1f GB]:", uicommon.GetCurrentDiskFree()),
									fmt.Sprintf("Disk Free Warn Threshold (GB) [Current: %.1f GB]:", uicommon.GetCurrentDiskFree())),
								&ui.Detail.DiskWarn,
								c.DiskFreeThresholdGB,
								0.1, 50, 2,
							),
						},
					},
					VSpacer{},
				},
			},
		},
	}
}
