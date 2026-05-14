package support

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/winapi"
	"fmt"
	"os"
	"runtime/debug"
)

// GlobalPanicHandler はアプリケーション全体で発生した予期せぬクラッシュを捕捉し、
// ログ記録とユーザーへの通知を行った後に安全に終了します。
// main 関数の先頭で defer して使用してください。
func GlobalPanicHandler() {
	if r := recover(); r != nil {
		// エラー内容とスタックトレースをログに記録
		config.Log("ERROR", "アプリケーションがクラッシュしました: %v\n%s", "Application panicked: %v\n%s", r, debug.Stack())

		// ユーザー向けの日本語エラーメッセージを組み立て
		msg := fmt.Sprintf("予期せぬエラーが発生したため、ツールを終了します。\n\n【原因】\n%v\n\n", r)
		msg += "※何度も発生する場合は、「サポート用ログ出力」または logs/latest.log を管理者へ送付してください。"

		// ポップアップで通知
		winapi.ShowDialog("システムエラー", msg, winapi.MB_ICONERROR|winapi.MB_SYSTEMMODAL)
		os.Exit(1)
	}
}
