package ui

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"
	uicommon "Evidence-Capture/internal/ui/common"
	"Evidence-Capture/internal/winapi"
	"fmt"

	"github.com/go-ole/go-ole"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var statusBar *walk.TextLabel

func RunTOCManager(owner walk.Form) {
	var dlg *walk.Dialog
	var createTOCBtn *walk.PushButton
	// ボタンを個別に変数で持つ
	var syncBtn, sortBtn, renameBtn *walk.PushButton

	exists, hasData, _ := excel.EnsureTOCSheet()

	if err := (Dialog{
		AssignTo: &dlg,
		Title:    "目次管理メニュー",
		Icon:     2,
		MinSize:  Size{Width: 224, Height: 196},
		Layout:   VBox{},
		Children: []Widget{
			PushButton{
				AssignTo: &createTOCBtn,
				Text:     "目次シート作成",
				MinSize:  Size{Height: 50},
				Enabled:  !exists,
				OnClicked: func() {
					createTOCBtn.SetEnabled(false)
					if statusBar != nil {
						statusBar.SetText("目次シート作成中...")
					}
					go func() {
						ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
						defer ole.CoUninitialize()

						err := excel.TOC_CreateTOCSheet()

						dlg.Synchronize(func() {
							if err != nil {
								uicommon.ShowErrorWithLogPrompt(dlg, "エラー", err.Error())
								createTOCBtn.SetEnabled(true)
								if statusBar != nil {
									statusBar.SetText("エラーが発生しました。")
								}
							} else {
								if statusBar != nil {
									statusBar.SetText("目次シートを作成しました。C4セルから下に入力してください。")
								}
								if syncBtn != nil {
									syncBtn.SetEnabled(true)
								}
								if sortBtn != nil {
									sortBtn.SetEnabled(true)
								}
								if renameBtn != nil {
									renameBtn.SetEnabled(true)
								}
							}
						})
					}()
				},
			},
			PushButton{
				AssignTo: &syncBtn,
				Text:     "シート作成",
				MinSize:  Size{Height: 50},
				Enabled:  exists && hasData,
				OnClicked: func() {
					syncBtn.SetEnabled(false)
					if statusBar != nil {
						statusBar.SetText("シート作成中...")
					}
					go func() {
						ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
						defer ole.CoUninitialize()

						err := excel.TOC_SyncSheetsFromList()

						dlg.Synchronize(func() {
							syncBtn.SetEnabled(true)
							if err != nil {
								uicommon.ShowErrorWithLogPrompt(dlg, "エラー", err.Error())
								if statusBar != nil {
									statusBar.SetText("エラーが発生しました。")
								}
							} else {
								if statusBar != nil {
									statusBar.SetText("不足しているシートを作成しました。")
								}
							}
						})
					}()
				},
			},
			PushButton{
				AssignTo: &sortBtn,
				Text:     "シート並び替え",
				MinSize:  Size{Height: 50},
				Enabled:  exists,
				OnClicked: func() {
					sortBtn.SetEnabled(false)
					if statusBar != nil {
						statusBar.SetText("並び替え中...")
					}
					go func() {
						ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED)
						defer ole.CoUninitialize()

						err := excel.TOC_SortSheets()

						dlg.Synchronize(func() {
							sortBtn.SetEnabled(true)
							if err != nil {
								uicommon.ShowErrorWithLogPrompt(dlg, "エラー", fmt.Sprintf("インデックスシートの作成に失敗しました:\n%v", err))
								if statusBar != nil {
									statusBar.SetText("エラーが発生しました。")
								}
							} else {
								if statusBar != nil {
									statusBar.SetText("並び替えが完了しました。")
								}
							}
						})
					}()
				},
			},
			VSpacer{Size: 10},

			PushButton{
				Text:    "インデックス管理メニューへ",
				MinSize: Size{Height: 50},
				OnClicked: func() {
					RunIndexManager(dlg)
				},
			},

			VSpacer{Size: 20},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					HSpacer{},
					PushButton{
						Text: "目次管理を終了",
						OnClicked: func() {
							dlg.Close(0)
						},
					},
				},
			},
			TextLabel{
				AssignTo: &statusBar,
				Text:     "",
			},
		},
	}).Create(owner); err != nil {
		config.Log("ERROR", "目次管理メニュー画面作成失敗: %v", "Failed to create TOC management menu: %v", err)
		return
	}

	if statusBar != nil {
		statusBar.SetText("準備完了")
	} else {
		config.Log("ERROR", "目次管理メニュー: statusBarへのバインド失敗 (nil)", "TOC management menu: statusBar binding failed (nil)")
	}

	winapi.SetAlwaysOnTop(dlg, config.CurrentConfig.EnableTopMostTOC)

	dlg.Run()
}
