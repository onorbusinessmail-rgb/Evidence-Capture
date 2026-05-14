package main

import (
	"Evidence-Capture/internal/config"
	"Evidence-Capture/internal/process"
	"Evidence-Capture/internal/support"
	"Evidence-Capture/internal/ui"
	uicommon "Evidence-Capture/internal/ui/common"
	"flag"
	"log"
)

func main() {
	// 標準ロガーの初期フラグ設定
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// グローバルパニックハンドラ
	support.GlobalPanicHandler()

	// 起動時のリソースクリーンアップ（前回残った一時フォルダを削除）
	process.CleanupTemporaryFolders()

	// 二重起動チェック
	const mutexName = "Global\\Evidence-Capture-Unique-ID-2026"
	mutexHandle := process.EnsureSingleInstance(mutexName)
	defer process.ReleaseMutex(mutexHandle)

	// 初期化シーケンス：設定を一度だけ読み込む
	cfg, err := config.LoadAppConfig()
	if err != nil {
		log.Printf("初期設定の読み込みに失敗しました: %v", err)
	}
	config.CurrentConfig = cfg

	// ログ設定の反映 (設定ファイルの内容に基づき出力先を決定)
	logFile, _ := config.InitLogger(cfg)
	if logFile != nil {
		defer logFile.Close()
		// ログフォルダの容量監視と自動クリーンアップ（リファクタリング後の呼び出し）
		config.HandleLogCapacity(cfg, func(curr, max int) bool {
			return uicommon.ShowLogCleanupPrompt(nil, curr, max)
		})
	}

	// 設定の読み込み完了後のロギング開始
	config.Log("INFO", "アプリを起動しています...", "Evidence-Capture Starting...")

	// 引数の定義
	forceLimit := flag.Bool("force-limit", false, "Force display of license limit UI")
	flag.Parse()

	// 引数が指定された場合、強制的にライセンス状態を「制限」に書き換える
	if *forceLimit {
		// licenseパッケージ等に状態を上書きする関数がある想定
		// もし変数を直接操作できるなら： license.CurrentStatus = license.StatusLimited
		ui.IsForceLimitMode = true
	}

	// アプリケーション（UI）の起動
	ui.StartApp()
}
