package uicommon

import (
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/winapi"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var (
	ErrorBrush, _ = walk.NewSolidColorBrush(walk.RGB(255, 200, 200)) // 薄い赤
	WhiteBrush, _ = walk.NewSolidColorBrush(walk.RGB(255, 255, 255)) // 白
)

func GetCurrentDiskFree() float64 {
	return winapi.GetDiskFreeSpaceGB("C:\\")
}

// ConfigParams はダイアログ構成に必要なパラメータをまとめます
type ConfigParams struct {
	Title            string
	MinSize          Size
	IniPath          string
	DefaultIni       string
	GetPages         func(dlg *walk.Dialog) []TabPage
	OnSave           func(dlg *walk.Dialog) bool
	OnSizeChanged    func() // オプション（nil可）
	IsForceLimitMode bool   // デバッグ用：trueなら強制的に制限モードのUIを表示
	InitialTabIndex  int    // 起動時に表示するタブインデックス
	TabWidget        **walk.TabWidget
}

type MyLineEdit struct {
	LineEdit
	BannerText string
}

func (m MyLineEdit) Create(builder *Builder) error {
	if err := m.LineEdit.Create(builder); err != nil {
		return err
	}
	le := m.LineEdit.AssignTo
	if m.BannerText != "" && le != nil {
		builder.Defer(func() error {
			window := (*le).Handle()
			winapi.SetCueBanner(uintptr(window), m.BannerText)
			return nil
		})
	}
	return nil
}

func ShowError(owner walk.Form, msg string) {
	walk.MsgBox(owner, "入力エラー", msg, walk.MsgBoxIconError)
}

// ヘルパー関数：チェックボックス付きの行を生成する
func CheckRow(label string, assign **walk.CheckBox, checked bool) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			CheckBox{AssignTo: assign, Checked: checked},
		},
	}
}

// ヘルパー関数：チェックボックス付きの行を生成する
func CheckRowCustom(label string, assign **walk.CheckBox, checked bool, onClicked func()) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			CheckBox{
				AssignTo:  assign,
				Checked:   checked,
				OnClicked: onClicked,
			},
		},
	}
}

// ヘルパー関数：単一の Composite (Widget) を返す
func PathBrowseComposite(label string, assign **walk.LineEdit, text string, isFolder bool) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{
						AssignTo: assign,
						Text:     text,
					},
					PushButton{
						Text: "参照",
						OnClicked: func() {
							targetLineEdit := *assign
							if targetLineEdit == nil {
								return
							}

							dlg := new(walk.FileDialog)
							owner := targetLineEdit.Form()

							if isFolder {
								dlg.Title = "フォルダを選択"
								if ok, _ := dlg.ShowBrowseFolder(owner); ok {
									targetLineEdit.SetText(dlg.FilePath)
								}
							} else {
								dlg.Title = "ファイルを選択"
								dlg.Filter = "Excelファイル (*.xlsx;*.xlsm)|*.xlsx;*.xlsm|すべてのファイル (*.*)|*.*"
								if ok, _ := dlg.ShowOpen(owner); ok {
									targetLineEdit.SetText(dlg.FilePath)
								}
							}
						},
					},
				},
			},
		},
	}
}

// ヘルパー関数：選択式（プルダウン）の行を生成する
func ComboRow(label string, assign **walk.ComboBox, current int, items []types.ComboItem) Widget {
	currIdx := 0
	for i, item := range items {
		if item.Value == current {
			currIdx = i
			break
		}
	}
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			ComboBox{
				AssignTo:     assign,
				Model:        &types.ComboModel{Items: items}, // 構造体リストを渡す
				CurrentIndex: currIdx,
			},
		},
	}
}

// ヘルパー関数：範囲制限付きの数値入力行を生成する
func NumberRowWithRange(label string, assign **walk.NumberEdit, value float64, min, max float64, decimals int) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			NumberEdit{
				AssignTo: assign,
				Value:    value,
				MinValue: min,
				MaxValue: max,
				Decimals: decimals,
			},
		},
	}
}

// ヘルパー関数：ComboBoxで選択されている項目の内部数値(Value)を返す
func GetComboValue(cb *walk.ComboBox) int {
	if cb == nil || cb.CurrentIndex() < 0 {
		return 1
	}
	if model, ok := cb.Model().(*types.ComboModel); ok {
		return model.Items[cb.CurrentIndex()].Value
	}
	return 1
}

// ヘルパー関数：入力中だけ表示し1秒後に伏せ字にするパスワード入力行を生成する
// 入力は半角英数字と一般的な記号のみに制限される
func PasswordRowWithTimer(label string, assign **walk.LineEdit, text string) Widget {
	validRegexp := regexp.MustCompile(`[ -~]*`)
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			LineEdit{
				AssignTo:     assign,
				Text:         text,
				PasswordMode: text != "",
				OnTextChanged: func() {
					le := *assign
					if le == nil {
						return
					}

					rawText := le.Text()

					// 1. 許可されていない文字（全角など）を除去
					// 正規表現で見つかった「正しい文字の塊」を結合して再構築
					cleanText := strings.Join(validRegexp.FindAllString(rawText, -1), "")

					// もし不正な文字が混じっていたら、テキストを差し替える
					if rawText != cleanText {
						// SetTextを呼ぶと再びOnTextChangedが走るが、次は rawText == cleanText になるため無限ループはしません
						le.SetText(cleanText)
						return
					}

					// 2. 伏せ字制御ロジック
					if cleanText == "" {
						le.SetPasswordMode(false)
						return
					}

					// 入力した瞬間は見える状態にする
					le.SetPasswordMode(false)

					// 1秒後に伏せ字にするタイマー
					time.AfterFunc(1000*time.Millisecond, func() {
						le.Synchronize(func() {
							// 現在のテキストが空でなければ伏せ字にする
							if le.Text() != "" {
								le.SetPasswordMode(true)
							}
						})
					})
				},
			},
		},
	}
}

// ラベルとチェックボックスを横並びにしたWidgetを返します
func CheckBoxRow(label string, assign **walk.CheckBox, checked bool, onClicked func()) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			CheckBox{
				AssignTo:  assign,
				Checked:   checked,
				OnClicked: onClicked,
			},
		},
	}
}

// ラベルとチェックボックスを横並びにしたWidgetを返し、有効状態を指定可能です
func CheckBoxRowEnabled(label string, assign **walk.CheckBox, checked bool, enabled bool, onClicked func()) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			CheckBox{
				AssignTo:  assign,
				Checked:   checked,
				Enabled:   enabled,
				OnClicked: onClicked,
			},
		},
	}
}

func LineEditRow(label string, assign **walk.LineEdit, text string, enabled bool, onChanged func()) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			LineEdit{
				AssignTo:      assign,
				Text:          text,
				Enabled:       enabled,
				OnTextChanged: onChanged,
			},
		},
	}
}

// ラベルとコンボボックスを横並びにしたWidgetを返します
func ComboBoxRow(label string, assign **walk.ComboBox, model interface{}, currentIndex int) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			ComboBox{
				AssignTo:      assign,
				Model:         model,
				DisplayMember: "Name",
				CurrentIndex:  currentIndex,
			},
		},
	}
}

// ラベルとコンボボックスを横並びにしたWidgetを返し、変更イベントをサポートします
func ComboBoxRowCustom(label string, assign **walk.ComboBox, model interface{}, currentIndex int, onIndexChanged func()) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			ComboBox{
				AssignTo:              assign,
				Model:                 model,
				DisplayMember:         "Name",
				CurrentIndex:          currentIndex,
				OnCurrentIndexChanged: onIndexChanged,
			},
		},
	}
}

// ラベル、テキスト入力、参照ボタンを横並びにしたWidgetを返します
func FolderBrowserRow(label string, assign **walk.LineEdit, text string) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					LineEdit{
						AssignTo: assign,
						Text:     text,
					},
					PushButton{
						Text: "参照",
						OnClicked: func() {
							le := *assign
							if le == nil {
								return
							}
							dlg := new(walk.FileDialog)
							if ok, _ := dlg.ShowBrowseFolder(le.Form()); ok {
								le.SetText(dlg.FilePath)
							}
						},
					},
				},
			},
		},
	}
}

// ラベルと数値入力を横並びにしたWidgetを返します
func NumberEditRow(label string, assign **walk.NumberEdit, value float64, min, max float64, decimals int) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			NumberEdit{
				AssignTo: assign,
				Value:    value,
				MinValue: min,
				MaxValue: max,
				Decimals: decimals,
			},
		},
	}
}

// 3つの入力フィールドと追加ボタンを横並びにしたWidgetを返します
func StatusInputRow(label string, idAssign, labelAssign, colorAssign **walk.LineEdit, onAdd func()) Widget {
	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 200},
				Alignment: AlignHFarVCenter,
			},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					MyLineEdit{LineEdit: LineEdit{AssignTo: idAssign}, BannerText: "ID"},
					MyLineEdit{LineEdit: LineEdit{AssignTo: labelAssign}, BannerText: "状態名"},
					MyLineEdit{LineEdit: LineEdit{AssignTo: colorAssign}, BannerText: "背景色"},
					PushButton{
						Text:      "追加",
						OnClicked: onAdd,
					},
				},
			},
		},
	}
}

// 左側に200pxのラベル（空でも可）を置き、右側にボタンを配置します
func ActionButtonRow(label string, assign **walk.PushButton, text string, enabled bool, onClick func()) Widget {
	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			Label{Text: label, MinSize: Size{Width: 200}},
			PushButton{
				AssignTo:  assign,
				Text:      text,
				Enabled:   enabled,
				OnClicked: onClick,
			},
		},
	}
}

func TableViewRow(label string, view TableView) Widget {
	return Composite{
		Layout: HBox{MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 200},
				Alignment: AlignHFarVCenter,
			},
			view,
		},
	}
}

// ラベルと固定テキストを横並びにしたWidgetを返します
func StaticTextRow(label, value string) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			Label{
				Text:      value,
				Alignment: AlignHFarVCenter,
			},
		},
	}
}

// ラベルと読み取り専用のテキスト入力（コピー用）を横並びにしたWidgetを返します
func ReadOnlyLineEditRow(label, text string) Widget {
	return Composite{
		Layout: Grid{Columns: 2, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			LineEdit{
				Text:     text,
				ReadOnly: true,
			},
		},
	}
}

// チェックボックスとテキスト入力を横並びにしたWidgetを返します
func CheckBoxWithEditRow(label1 string, assignCB **walk.CheckBox, checked bool, onClick func(), label2 string, assignLE **walk.LineEdit, text string, enabled bool, onChanged func(), label2Width int) Widget {
	return Composite{
		Layout: Grid{Columns: 4, MarginsZero: true},
		Children: []Widget{
			Label{
				Text:      label1,
				MinSize:   Size{Width: 150},
				Alignment: AlignHFarVCenter,
			},
			CheckBox{
				AssignTo:  assignCB,
				Checked:   checked,
				OnClicked: onClick,
			},
			Label{
				Text:      label2,
				MinSize:   Size{Width: label2Width},
				Alignment: AlignHFarVCenter,
			},
			LineEdit{
				AssignTo:      assignLE,
				Text:          text,
				Enabled:       enabled,
				OnTextChanged: onChanged,
			},
		},
	}
}

func RunBaseConfigDialog(owner walk.Form, p ConfigParams, onAccept func()) {
	var dlg *walk.Dialog
	var tabWidget *walk.TabWidget

	// 呼び出し元が TabWidget の参照を提供していない場合は、ローカル変数を使用
	twAssign := p.TabWidget
	if twAssign == nil {
		twAssign = &tabWidget
	}

	d := Dialog{
		AssignTo: &dlg,
		Title:    p.Title,
		Icon:     2,
		MinSize:  p.MinSize,
		Layout:   VBox{},
		Children: []Widget{
			TabWidget{
				AssignTo: twAssign,
				Pages:    p.GetPages(dlg),
			},
			VSpacer{Size: 15},
			ConfigDialogFooter(&dlg, p.IniPath, p.DefaultIni, p.OnSave, onAccept),
		},
		OnSizeChanged: func() {
			if p.OnSizeChanged != nil {
				p.OnSizeChanged()
			}
		},
	}

	if err := d.Create(owner); err != nil {
		return
	}

	// 初期表示タブの指定がある場合、ウィジェット生成後に設定する
	if twAssign != nil && *twAssign != nil && p.InitialTabIndex > 0 {
		(*twAssign).SetCurrentIndex(p.InitialTabIndex)
	}

	dlg.Run()
}

// 「初期化」「保存」「キャンセル」のボタンセットを生成します
func ConfigDialogFooter(dlg **walk.Dialog, iniPath string, defaultIni string, onSave func(dlg *walk.Dialog) bool, onAccept func()) Composite {
	return Composite{
		Layout: HBox{},
		Children: []Widget{
			PushButton{
				Text: "初期化",
				OnClicked: func() {
					if winapi.ShowDialog("警告", "設定を初期化しますか？", winapi.MB_ICONWARNING|winapi.MB_YESNO) == winapi.IDYES {
						os.WriteFile(iniPath, []byte(defaultIni), 0666)
						winapi.ShowDialog("完了", "初期化しました。", 0)
						(*dlg).Cancel()
					}
				},
			},
			HSpacer{},
			PushButton{
				Text: "保存",
				OnClicked: func() {
					if onSave(*dlg) {
						winapi.ShowDialog("完了", "設定を保存しました。", 0)
						(*dlg).Accept()
						if onAccept != nil {
							onAccept()
						}
					}
				},
			},
			PushButton{
				Text:      "キャンセル",
				OnClicked: func() { (*dlg).Cancel() },
			},
		},
	}
}
