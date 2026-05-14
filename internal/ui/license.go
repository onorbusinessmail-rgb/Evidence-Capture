package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/license"
	"Evidence-Capture/internal/ui/common"
	"fmt"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func RunLicenseManager(owner walk.Form) {
	var dlg *walk.Dialog

	err := Dialog{
		AssignTo:	&dlg,
		Title:		"ライセンス情報",
		Icon:		2,
		MinSize:	Size{Width: 450, Height: 480},
		Layout:		VBox{Margins: Margins{Left: 20, Top: 20, Right: 20, Bottom: 20}, Spacing: 10},
		Children: []Widget{
			GroupBox{
				Title:	"ライセンス利用状況",
				Layout:	VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}, Spacing: 5},
				Children: []Widget{
					uicommon.ReadOnlyLineEditRow("ライセンスモード：", config.CurrentConfig.LicenseMode),
					uicommon.ReadOnlyLineEditRow("ツール利用開始日：", getDefaultIfEmpty(config.CurrentConfig.StartDate)),
					uicommon.ReadOnlyLineEditRow("ツール利用終了日：", getDefaultIfEmpty(config.CurrentConfig.EndDate)),
					uicommon.ReadOnlyLineEditRow("残り利用可能回数：", fmt.Sprintf("%d 回", config.CurrentConfig.RemainingCount)),
					uicommon.ReadOnlyLineEditRow("現在の撮影枚数：", fmt.Sprintf("%d 枚", config.CurrentConfig.SessionImageCount)),
					uicommon.ReadOnlyLineEditRow("前回利用年月日：", getDefaultIfEmpty(config.CurrentConfig.LastRunDate)),

					// 制限モード時の待機時間がある場合は表示
					func() Widget {
						if config.CurrentConfig.IsRestricted && config.CurrentConfig.NextAvailableTime != "" {
							return Label{
								Text:		fmt.Sprintf("※ 次回利用可能時刻: %s", config.CurrentConfig.NextAvailableTime),
								TextColor:	walk.RGB(255, 0, 0),
								Font:		Font{PointSize: 9, Bold: true},
							}
						}
						return Label{}
					}(),
				},
			},
			GroupBox{
				Title:	"システム情報",
				Layout:	VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}, Spacing: 5},
				Children: []Widget{
					uicommon.ReadOnlyLineEditRow("ツールバージョン：", config.CurrentConfig.Version),
					uicommon.ReadOnlyLineEditRow("マシン固有ID：", license.GetMachineID()),
				},
			},
			VSpacer{},
			Composite{
				Layout:	HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text:	"閉じる",
						OnClicked: func() {
							dlg.Accept()
						},
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		config.Log("ERROR", "ライセンス情報ダイアログの作成に失敗しました: %v", "Failed to create license dialog: %v", err)
		return
	}

	dlg.Run()
}

// ヘルパー関数：日付が空文字の場合に "--/--/--" を表示させる
func getDefaultIfEmpty(val string) string {
	if val == "" {
		return "--/--/--"
	}
	return val
}
