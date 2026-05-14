============================================================
  Project Structure (Tree View) - Evidence Capture
============================================================
E01_Evidence-Capture/
├── Evidence-Capture.exe          (メイン実行ファイル：エビデンス取得)
├── Evidence-Captureガイド.xlsx    (ユーザー向け操作マニュアル)
├── cmd/
│   └── capture/
│       └── main.go               (プログラムのエントリポイント・UI起動)
├── internal/
│   ├── capture/
│   │   ├── capture.go            (Win32 APIによるアクティブウィンドウキャプチャ)
│   │   ├── destination.go        (保存先・出力先パス管理)
│   │   └── workflow.go           (キャプチャ実行フローの制御)
│   ├── clipboard/
│   │   └── clipboard.go          (クリップボードからの画像抽出・ファイル保存)
│   ├── config/
│   │   ├── config.go             (INI/DATファイルの読み込み・保存・絶対パス管理)
│   │   ├── constants.go          (INIセクション・キー・UIカラー等の定数定義)
│   │   ├── defaults.go           (各設定ファイルのデフォルトテンプレート管理)
│   │   └── validator.go          (設定値のバリデーション・不正値の自動復旧)
│   ├── excel/
│   │   ├── constants.go          (Excel OLE操作用の数値定数定義)
│   │   ├── destination.go        (Excel出力先・保存管理)
│   │   ├── excel.go              (Excel.Application操作・OLEコアロジック)
│   │   ├── excel_bulk.go         (証跡画像の一括挿入・PDFエクスポート)
│   │   ├── excel_image.go        (画像挿入・座標計算・オートシェイプ管理)
│   │   ├── excel_index.go        (インデックスシートの自動生成・書式設定)
│   │   ├── excel_toc.go          (目次シートの生成・ハイパーリンク設定)
│   │   ├── excel_utils.go        (Excel操作用ヘルパー関数群)
│   │   ├── recorder.go           (Excel操作のレコーディング・状態記録)
│   │   └── types.go              (Excelモジュール専用の型定義)
│   ├── imageutil/
│   │   └── imageutil.go          (画像リサイズ・フォーマット変換・メタデータ処理)
│   ├── license/
│   │   ├── checker.go            (利用制限・時刻逆転監視・整合性チェック)
│   │   ├── crypto.go             (system.datの暗号化・復号化エンジン)
│   │   └── manager.go            (マシンID生成・ライセンスマスター情報管理)
│   ├── process/
│   │   ├── excel_process.go      (Excelプロセスの死活監視・強制終了処理)
│   │   ├── mutex.go              (ミューテックスによる多重起動の防止)
│   │   └── process.go            (OSプロセスの列挙・ハンドル操作)
│   ├── support/
│   │   ├── diagnostic.go         (動作ログ出力・診断メッセージ管理)
│   │   └── recovery.go           (グローバルパニックリカバリ・エラー復旧)
│   ├── types/
│   │   └── types.go              (アプリ全体で共有する構造体・型の定義)
│   ├── ui/
│   │   ├── ui_main.go            (メイン画面：MainWindow定義)
│   │   ├── ui_widget.go          (カスタムウィジェット・UIパーツ定義)
│   │   └── (ui_*.go.bak)         (開発中・バックアップ中の設定/機能UIモジュール群)
│   ├── utils/
│   │   └── utils.go              (汎用ユーティリティ：パス正規化等)
│   └── winapi/
│       └── winapi.go             (Windows API (user32/kernel32) の呼び出し定義)
├── setting/
│   ├── config_evidence.ini       (基本動作・撮影トリガー設定)
│   ├── config_layout.ini         (Excelシート・画像レイアウト設定)
│   └── config_license.ini        (ライセンス情報の公開データ)
├── doc/
│   ├── 01_doc_format/            (各種ドキュメントの標準フォーマット)
│   │   └── 01_doc_format.md
│   ├── 02_architecture/          (システムの詳細設計・アーキテクチャ)
│   │   └── 02_architecture.md
│   ├── 03_rules/                 (開発制約・現場デプロイ要件)
│   │   └── 03_rules.md
│   ├── 04_skills/                (開発に必要なスキル・技術スタック)
│   │   └── 04_skills.md
│   ├── 05_structures/            (データ構造・クラス構成定義)
│   │   └── 05_structures.md
│   ├── 06_task_format/           (開発タスク管理・進捗状況)
│   │   └── 06_task_format.md
│   ├── 07_instructions/          (AIへの詳細指示書・開発ガイドライン)
│   │   └── 07_instructions.md
│   ├── 08_workflow/              (開発・デプロイ・アーカイブのワークフロー)
│   │   └── 08_workflow.md
│   ├── 09_tree/
│   │   └── 09_tree.md            (このファイル：プロジェクト構成図)
│   └── 10_問題/                  (既知のバグ・課題・優先度別問題管理)
│       ├── 01_致命的な問題.md
│       ├── 02_修正推奨の問題.md
│       └── 03_軽微な問題.md
├── scripts/                      (開発・デプロイ用補助スクリプト)
├── 01_build_dev.bat              (開発ビルド・テスト実行)
├── main.manifest                 (Windowsビジュアルスタイル用マニフェスト)
└── go.mod / go.sum               (Goモジュール依存関係管理)
============================================================