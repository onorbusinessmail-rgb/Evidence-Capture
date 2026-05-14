package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/excel"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// =========================================================================
// プレビューコントローラー（責務分離）
// =========================================================================

// previewController はプレビュー表示に関する状態と操作を管理します。
type previewController struct {
	tv               *walk.TreeView
	previewComposite *walk.ScrollView
	currentBitmaps   []*walk.Bitmap
	imageView        *walk.ImageView
	errorLabel       *walk.Label
}

// clearPreview はプレビューエリアをクリアします。
func (pc *previewController) clearPreview() {
	if pc.previewComposite == nil {
		return
	}
	pc.previewComposite.SetSuspended(true)
	defer func() {
		pc.previewComposite.SetSuspended(false)
		pc.previewComposite.Invalidate()
	}()

	children := pc.previewComposite.Children()
	for children.Len() > 0 {
		child := children.At(children.Len() - 1)
		child.SetParent(nil)
		child.Dispose()
	}
	for _, bmp := range pc.currentBitmaps {
		bmp.Dispose()
	}
	pc.currentBitmaps = nil
}

// dispose はダイアログ終了時にリソースを解放します。
func (pc *previewController) dispose() {
	for _, bmp := range pc.currentBitmaps {
		bmp.Dispose()
	}
	pc.currentBitmaps = nil
}

// updatePreview は指定されたファイルのプレビューを表示します。
func (pc *previewController) updatePreview(f *File) {
	if pc.previewComposite == nil || f == nil {
		return
	}

	pc.clearPreview()

	if f.path == "" {
		return
	}

	pc.previewComposite.SetSuspended(true)
	defer func() {
		pc.previewComposite.SetSuspended(false)
		pc.previewComposite.Invalidate()
	}()

	bmp, err := walk.NewBitmapFromFile(f.path)
	if err != nil {
		// 読み込み失敗時はラベルを表示
		lbl, _ := walk.NewLabel(pc.previewComposite)
		lbl.SetText(fmt.Sprintf("[%s]\n画像の読み込みに失敗しました", f.name))
		lbl.SetAlignment(walk.AlignHCenterVNear)
		return
	}
	pc.currentBitmaps = append(pc.currentBitmaps, bmp)

	iv, _ := walk.NewImageView(pc.previewComposite)
	iv.SetImage(bmp)
	iv.SetMode(walk.ImageViewModeZoom)

	// プレビューサイズの計算（ScrollView の 90%）
	viewBounds := pc.previewComposite.ClientBoundsPixels()
	maxW := int(float64(viewBounds.Width) * 0.90)
	maxH := int(float64(viewBounds.Height) * 0.90)
	if maxW <= 0 {
		maxW = 400
	}
	if maxH <= 0 {
		maxH = 400
	}

	origSize := bmp.Size()
	w, h := calcScaledSize(origSize.Width, origSize.Height, maxW, maxH)
	iv.SetMinMaxSize(
		walk.Size{Width: w, Height: h},
		walk.Size{Width: w, Height: h},
	)
	iv.SetSize(walk.Size{Width: w, Height: h})
	iv.Invalidate()
}

// calcScaledSize はアスペクト比を保持しながら maxW×maxH に収まるサイズを計算します。
func calcScaledSize(origW, origH, maxW, maxH int) (int, int) {
	if origW <= 0 || origH <= 0 || maxW <= 0 || maxH <= 0 {
		return origW, origH
	}
	w := maxW
	h := origH * maxW / origW
	if h > maxH {
		h = maxH
		w = origW * maxH / origH
	}
	return w, h
}

// =========================================================================
// ダイアログ本体
// =========================================================================

// RunImageImportDialog は画像インポート用のダイアログを表示します。
func RunImageImportDialog(owner walk.Form) {
	var dlg *walk.Dialog
	var tv *walk.TreeView
	var execBtn *walk.PushButton
	var scrollView *walk.ScrollView

	// 既存画像の抽出
	details, tempDir, err := excel.ExtractImagesFromExcel(config.CurrentConfig.ExcelOutputPath)
	if err != nil {
		walk.MsgBox(owner, "エラー", "既存画像の取得に失敗しました:\n"+err.Error(), walk.MsgBoxIconError)
		// 失敗してもダイアログ自体は開く（空の状態で）
	}
	defer func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
			config.Log("INFO", "一時画像ファイルを削除しました: %s", "Temporary images cleaned up: %s", tempDir)
		}
	}()

	model := NewDirectoryTreeModel(details)
	pc := &previewController{}

	err = Dialog{
		AssignTo: &dlg,
		Title:    "画像のインポート管理",
		Icon:     2,
		MinSize:  Size{Width: 800, Height: 500},
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: "追加先のシートを選択して画像をドロップしてください。Delキーで削除できます。"},
			HSplitter{
				Children: []Widget{
					TreeView{
						AssignTo: &tv,
						Model:    model,
						OnCurrentItemChanged: func() {
							item := tv.CurrentItem()
							switch t := item.(type) {
							case *File:
								pc.updatePreview(t)
							case *Directory:
								pc.clearPreview()
							default:
								pc.clearPreview()
							}
						},
					},
					ScrollView{
						AssignTo: &scrollView,
						Layout:   VBox{MarginsZero: true, Alignment: AlignHNearVNear},
					},
				},
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{
						AssignTo: &execBtn,
						Text:     "Excelへ反映して保存",
						Enabled:  false,
						OnClicked: func() {
							data := model.ExtractData()
							count, err := excel.BulkInsertImages(data)
							if err != nil {
								walk.MsgBox(dlg, "エラー", "Excelが開かれているため、書き込みに失敗しました。\n\n詳細: "+err.Error(), walk.MsgBoxIconError)
								config.Log("ERROR", "Excelが開かれているため、書き込みに失敗しました。\n\n詳細: %v", "Failed to insert images to Excel. Detail: %v", err)
								return
							}
							walk.MsgBox(dlg, "完了", fmt.Sprintf("%d枚の画像をExcelに反映しました。", count), walk.MsgBoxIconInformation)
							config.Log("INFO", "%d枚の画像をExcelに反映しました。", "%d images reflected in Excel.", count)
							dlg.Accept()
						},
					},
					PushButton{
						Text:      "閉じる",
						OnClicked: func() { dlg.Accept() },
					},
				},
			},
		},
	}.Create(owner)

	if err != nil {
		return
	}

	pc.tv = tv
	pc.previewComposite = scrollView

	for _, root := range model.roots {
		if len(root.children) > 0 {
			tv.SetExpanded(root, true)
		}
	}

	if model.GetTotalImageCount() > 0 {
		execBtn.SetEnabled(true)
	}

	tv.DropFiles().Attach(func(files []string) {
		item := tv.CurrentItem()
		if item == nil {
			walk.MsgBox(dlg, "通知", "シートを選択してください", walk.MsgBoxIconInformation)
			return
		}

		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".bmp" {
				model.AddImageToNode(item, f)
			}
		}

		if model.GetTotalImageCount() > 0 {
			execBtn.SetEnabled(true)
		}

		switch t := item.(type) {
		case *File:
			tv.SetExpanded(t.parent, true)
		case *Directory:
			tv.SetExpanded(t, true)
		}
	})

	tv.KeyDown().Attach(func(key walk.Key) {
		if key == walk.KeyDelete {
			item := tv.CurrentItem()
			if item == nil {
				return
			}
			model.RemoveNode(item)
			execBtn.SetEnabled(model.GetTotalImageCount() > 0)
			pc.clearPreview()
		}
	})

	dlg.Run()
	pc.dispose()
}
