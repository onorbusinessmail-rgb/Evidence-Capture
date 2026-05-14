package ui

import (
	"fmt"
	"path/filepath"

	"Evidence-Capture/internal/excel"
	"Evidence-Capture/internal/types"

	"github.com/lxn/walk"
)

// =========================================================================
// TreeView ノード定義
// =========================================================================

// Directory はツリーの親ノード（シート）を表します
type Directory struct {
	name     string
	parent   *Directory
	children []walk.TreeItem
}

var _ walk.TreeItem = new(Directory)

func (d *Directory) Text() string {
	return fmt.Sprintf("%s (%d)", d.name, len(d.children))
}

func (d *Directory) Parent() walk.TreeItem {
	if d.parent == nil {
		return nil
	}
	return d.parent
}

func (d *Directory) ChildCount() int {
	return len(d.children)
}

func (d *Directory) ChildAt(index int) walk.TreeItem {
	return d.children[index]
}

func (d *Directory) Image() interface{} {
	return nil
}

// File はツリーの子ノード（画像）を表します
type File struct {
	name       string
	path       string // ファイルのフルパス（既存画像の場合は一時保存先のパス）
	parent     *Directory
	isExisting bool
}

var _ walk.TreeItem = new(File)

func (f *File) Text() string {
	return f.name
}

func (f *File) Parent() walk.TreeItem {
	return f.parent
}

func (f *File) ChildCount() int {
	return 0
}

func (f *File) ChildAt(index int) walk.TreeItem {
	return nil
}

func (f *File) Image() interface{} {
	return nil
}

func (f *File) Path() string {
	return f.path
}

// =========================================================================
// TreeView モデル定義
// =========================================================================

// DirectoryTreeModel は walk.TreeModel を実装します。
type DirectoryTreeModel struct {
	walk.TreeModelBase
	roots []*Directory
}

func NewDirectoryTreeModel(details []types.SheetInfo) *DirectoryTreeModel {
	m := &DirectoryTreeModel{}
	for _, info := range details {
		root := &Directory{
			name: info.Name,
		}
		// 既存の画像を抽出して子ノードに追加
		for _, imgPath := range info.ExistingImages {
			root.children = append(root.children, &File{
				name:       "[既存] " + filepath.Base(imgPath),
				path:       imgPath, // すでに一時パスが渡されている想定
				parent:     root,
				isExisting: true,
			})
		}
		m.roots = append(m.roots, root)
	}
	return m
}

func (m *DirectoryTreeModel) RootCount() int {
	return len(m.roots)
}

func (m *DirectoryTreeModel) RootAt(index int) walk.TreeItem {
	return m.roots[index]
}

func (m *DirectoryTreeModel) LazyPopulate(parent walk.TreeItem) error {
	return nil
}

// AddImageToNode は指定したノードに画像を追加し、ツリーを更新します
func (m *DirectoryTreeModel) AddImageToNode(target walk.TreeItem, path string) {
	var parent *Directory
	switch t := target.(type) {
	case *Directory:
		parent = t
	case *File:
		parent = t.parent
	default:
		return
	}

	if parent == nil {
		return
	}

	child := &File{
		name:   filepath.Base(path),
		path:   path,
		parent: parent,
	}
	parent.children = append(parent.children, child)

	// TreeView を再描画
	m.PublishItemsReset(parent)
	m.PublishItemChanged(parent) // 親ノードの件数表示を更新
}

// RemoveNode は指定したノードをモデルから削除します。
func (m *DirectoryTreeModel) RemoveNode(target walk.TreeItem) {
	if target == nil {
		return
	}

	switch t := target.(type) {
	case *File:
		parent := t.parent
		if parent == nil {
			return
		}
		for i, child := range parent.children {
			if child == t {
				parent.children = append(parent.children[:i], parent.children[i+1:]...)
				break
			}
		}
		m.PublishItemsReset(parent)
		m.PublishItemChanged(parent)
	case *Directory:
		// シートノードは配下の画像をすべてクリア
		t.children = nil
		m.PublishItemsReset(t)
		m.PublishItemChanged(t)
	}
}

// ExtractData はモデルを走査し、シート名と紐づく画像フルパスのリストをシートの位置順で抽出します。
func (m *DirectoryTreeModel) ExtractData() []excel.SheetImages {
	var result []excel.SheetImages
	for _, root := range m.roots {
		var newImages []string
		for _, child := range root.children {
			if f, ok := child.(*File); ok {
				if !f.isExisting {
					newImages = append(newImages, f.path)
				}
			}
		}
		if len(newImages) > 0 {
			result = append(result, excel.SheetImages{
				SheetName: root.name,
				Images:    newImages,
			})
		}
	}
	return result
}

// GetTotalImageCount はツリー全体に追加された画像の総数を返します。
func (m *DirectoryTreeModel) GetTotalImageCount() int {
	count := 0
	for _, root := range m.roots {
		for _, child := range root.children {
			if f, ok := child.(*File); ok {
				if !f.isExisting {
					count++
				}
			}
		}
	}
	return count
}
