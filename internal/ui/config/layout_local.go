package uiconfig

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/ui/common"

	. "github.com/lxn/walk/declarative"
)

func (ui *ConfigUILayout) createLocalSaveTabPage() TabPage {
	c := &config.CurrentConfig
	fmtItems := []types.ComboItem{{Name: config.T("1: PNG", "1: PNG"), Value: 1}}
	pasteItems := []types.ComboItem{{Name: config.T("1: 埋め込み", "1: Embedded"), Value: 1}}

	return TabPage{
		Title:  config.T("ローカル保存", "Local Save"),
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: VBox{Margins: Margins{Left: 15, Top: 15, Right: 15, Bottom: 15}},
				Children: []Widget{
					Composite{
						Layout: Grid{Columns: 2, MarginsZero: true},
						Children: []Widget{
							uicommon.CheckBoxRow(config.T("画像をパソコン内に保存:", "Save images to local computer:"),
								&ui.Local.SaveOn,
								c.LocalSaveOn,
								ui.updateUILayoutControls),
						},
					},
					GroupBox{
						AssignTo: &ui.Local.Group,
						Title:    config.T("詳細設定", "Detailed Settings"),
						Enabled:  c.LocalSaveOn,
						Layout:   Grid{Columns: 2},
						Children: []Widget{
							uicommon.ComboBoxRow(config.T("保存形式:", "Save Format:"),
								&ui.Local.Format,
								&types.ComboModel{Items: fmtItems},
								c.LocalSaveFormat-1),
							uicommon.FolderBrowserRow(config.T("保存先フォルダ:", "Save Folder:"),
								&ui.Local.Folder,
								c.LocalSaveFolder),
							uicommon.ComboBoxRow(config.T("貼り付け方式:", "Paste Mode:"),
								&ui.Local.PasteMode,
								&types.ComboModel{Items: pasteItems},
								c.PasteMode-1),
						},
					},
					VSpacer{},
				},
			},
		},
	}
}
