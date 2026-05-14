package uicommon

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/winapi"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/lxn/walk"
)

// ValidationItem は入力検証のルールを定義する汎用構造体です。
type ValidationItem struct {
	Widget   interface{}    // 対象ウィジェット (*walk.LineEdit, *walk.NumberEdit)
	Msg      string         // エラー時のメッセージ
	Required bool           // 必須入力か
	IsNumber bool           // 数値範囲チェックを行うか
	MinVal   float64        // 最小値
	MaxVal   float64        // 最大値
	Regex    *regexp.Regexp // 特定書式の正規表現
	IsFolder bool           // フォルダの実在チェックを行うか
}

// ValidateUIInputs は渡された検証ルールのリストを一括でチェックする共通機能です。
func ValidateUIInputs(dlg *walk.Dialog, items ...ValidationItem) bool {
	for _, item := range items {
		if item.Widget == nil {
			continue
		}

		var valStr string
		var valNum float64

		switch w := item.Widget.(type) {
		case *walk.LineEdit:
			valStr = strings.TrimSpace(w.Text())
			w.SetBackground(WhiteBrush) // 背景リセット
			if n, err := strconv.ParseFloat(valStr, 64); err == nil {
				valNum = n
			}
		case *walk.NumberEdit:
			valNum = w.Value()
			valStr = fmt.Sprintf("%g", valNum)
			w.SetBackground(WhiteBrush) // 背景リセット
		}

		if item.Required && valStr == "" {
			return HandleValidationError(dlg, item.Widget, item.Msg)
		}
		if item.IsNumber && (valNum < item.MinVal || valNum > item.MaxVal) {
			return HandleValidationError(dlg, item.Widget, item.Msg)
		}
		if item.Regex != nil && valStr != "" && !item.Regex.MatchString(valStr) {
			return HandleValidationError(dlg, item.Widget, item.Msg)
		}
		// フォルダ存在チェック機能の追加
		if item.IsFolder && valStr != "" {
			if fi, err := os.Stat(valStr); err != nil || !fi.IsDir() {
				return HandleValidationError(dlg, item.Widget, item.Msg)
			}
		}
	}
	return true
}

// HandleValidationError はエラー時のUIアクション（背景赤色、フォーカス、ダイアログ表示）を共通処理します。
func HandleValidationError(dlg *walk.Dialog, widget interface{}, msg string) bool {
	switch w := widget.(type) {
	case *walk.LineEdit:
		w.SetBackground(ErrorBrush)
		w.Invalidate()
		w.SetFocus()
	case *walk.NumberEdit:
		w.SetBackground(ErrorBrush)
		w.Invalidate()
		w.SetFocus()
	}

	if dlg == nil {
		winapi.ShowDialog("入力エラー", msg, winapi.MB_ICONWARNING)
	} else {
		walk.MsgBox(dlg, "入力エラー", msg, walk.MsgBoxIconWarning)
	}

	config.Log("ERROR", "入力エラー: %v", "Validation error: %v", msg)
	return false
}
