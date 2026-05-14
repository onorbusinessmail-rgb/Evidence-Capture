package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"
	"Evidence-Capture/internal/types"
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

// インデックスボタンメニュー
func RunIndexButtonManager(owner walk.Form) {
	var dlg *walk.Dialog
	var cbApplyColor *walk.CheckBox

	// 設定ファイルからステータス定義を取得
	var statuses []types.IndexStatus
	for _, val := range config.CurrentConfig.StatusDefinitions {
		parts := strings.Split(val, "|")
		// 4項目目が存在し、かつそれが「可」または「ON」である場合のみボタンを生成
		if len(parts) >= 4 {
			st := strings.TrimSpace(parts[3])
			if st == "可" || st == "ON" {
				name := strings.TrimSpace(parts[1])
				status := types.IndexStatus{
					ID:    parts[0],
					Name:  name,
					Color: parts[2],
				}
				statuses = append(statuses, status)
			}
		}
	}
	numStatus := len(statuses)
	buttonFontSize := 11
	buttonHeight := 50   // 高さを抑えてスマートにする
	columns := numStatus // ステータス数と同じにすることで全ボタンを横一列にする

	// ボタン数が少ない（例：5つ以下）場合は、より1つを大きく見せる調整も可能
	if numStatus <= 5 {
		buttonHeight = 35
		buttonFontSize = 11
	}

	// ステータスボタンのスライス（キャンセルは別枠で配置）
	statusButtons := []Widget{}
	for _, st := range statuses {
		status := st
		statusButtons = append(statusButtons, PushButton{
			Text:    status.Name,
			MinSize: Size{Height: buttonHeight}, // 幅の固定値を外しOSに自動調整させる
			Font:    Font{PointSize: buttonFontSize, Bold: true},
			OnClicked: func() {
				config.Log("INFO", "ステータスが選択されました: %s", "Status selected: %s", status.Name)
				if err := excel.UpdateIndexStatus(status.Name, cbApplyColor.Checked(), status.Color); err != nil {
					walk.MsgBox(dlg, "Excel更新エラー", fmt.Sprintf("Excelの更新に失敗しました:\n%v", err), walk.MsgBoxIconError)
				}

			},
		})
	}

	err := Dialog{
		AssignTo: &dlg,
		Title:    "ステータス更新 - インデックス",
		Icon:     2,
		MinSize: Size{
			Width:  400, // OSのスケーリングに合わせ、最低限の幅を確保して自動拡張させる
			Height: 120,
		},
		// 上下のMarginsを0にし、Spacing（要素間の隙間）も0にする
		Layout: VBox{Margins: Margins{Left: 10, Top: 0, Right: 10, Bottom: 0}, Spacing: 0},
		Children: []Widget{
			Label{
				Text:      "更新するステータスを選択してください：",
				Alignment: AlignHNearVCenter,
				Font:      Font{PointSize: 10},
				MinSize:   Size{Height: 25},
			},

			// メインのステータスボタンエリア
			Composite{
				Layout:   Grid{Columns: columns, Spacing: 2, MarginsZero: true},
				Children: statusButtons,
			},

			// チェックボックスとキャンセルのエリア
			Composite{
				Layout: HBox{MarginsZero: true, Spacing: 0},
				Children: []Widget{
					CheckBox{
						AssignTo: &cbApplyColor,
						Text:     "選択したステータスの背景色を反映する",
						Font:     Font{PointSize: 9},
						Checked:  true,
					},
					HSpacer{},
					PushButton{
						Text:      "キャンセル",
						MinSize:   Size{Width: 100, Height: 30},
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		config.Log("ERROR", "IndexManagerダイアログの作成に失敗しました。詳細: %v", "Failed to create IndexManager dialog. Detail: %v", err)
		return
	}
	if dlg != nil {
		dlg.SetVisible(true)
		win.SetWindowPos(dlg.Handle(), win.HWND_TOPMOST, 0, 0, 0, 0, win.SWP_NOMOVE|win.SWP_NOSIZE|win.SWP_SHOWWINDOW)
	}
	dlg.Run()
}
