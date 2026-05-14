package uicommon

import (
	"Evidence-Capture/internal/config"
	"os/exec"
	"strconv"

	"github.com/lxn/walk"
)

// ShowErrorWithLogPrompt はエラーダイアログを表示し、ユーザーの選択に応じてログファイルをメモ帳で開きます。
func ShowErrorWithLogPrompt(owner walk.Form, title, message string) {
	prompt := message + "\n\n詳細なログファイルを開いて確認しますか？"

	if walk.MsgBox(owner, title, prompt, walk.MsgBoxIconError|walk.MsgBoxYesNo) == walk.DlgCmdYes {
		logPath := config.GetCurrentLogPath()
		if logPath != "" {
			// セキュリティ対策として、cmd.exe経由ではなくnotepad.exeを直接起動する
			exec.Command("notepad.exe", logPath).Start()
		}
	}
}
func ShowLogCleanupPrompt(owner walk.Form, currentSizeMB, maxSizeMB int) bool {
	title := "ログフォルダ容量警告"
	message := "ログフォルダのサイズ（" + strconv.Itoa(currentSizeMB) + " MB）が、設定された警告サイズ（" + strconv.Itoa(maxSizeMB) + " MB）を超過しています。\n\nストレージ容量を確保するため、過去の古いログアーカイブ（ZIPファイル）を削除しますか？"

	return walk.MsgBox(owner, title, message, walk.MsgBoxIconWarning|walk.MsgBoxYesNo) == walk.DlgCmdYes
}
