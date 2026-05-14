// Package config のINIセクション・キー定数定義です。
package config

// =========================================================================
// config_evidence.ini 定数
// =========================================================================

const (
	SecBaseTrigger = "1-1_証跡取得-トリガー設定"
	SecBaseCapture = "1-2_証跡取得-撮影環境設定"
	SecBaseFile    = "1-3_証跡取得-Excelファイル制御設定"

	KeyTriggerTopLeft           = "撮影トリガー：画面左上へのマウス移動"
	KeyTriggerTopRight          = "撮影トリガー：画面右上へのマウス移動"
	KeyTriggerRightDbl          = "撮影トリガー：右ダブルクリック"
	KeyExitLeftDbl              = "終了トリガー：画面左端でのダブルクリック"
	KeyExitCtrlShiftX           = "終了トリガー：Ctrl+Shift+Xキー"
	KeyExitWinKey               = "終了トリガー：Windowsキー"
	KeyExitTimeout              = "終了トリガー：タイムアウト"
	KeyAutoSwitchTimeoutMinutes = "環境設定：タイムアウトまでの時間（分）"
	KeyLanguage                 = "環境設定：言語"
	KeyWaitTimeSeconds          = "撮影制御：撮影後の待機時間（秒）"
	KeyEdgeSensitivity          = "環境設定：画面端の判定感度"
	KeyExcelOutputPath          = "環境設定：証跡格納先Excelファイルパス"
	KeyMinimizeExcel            = "環境設定：起動時にExcelを最小化する"
	KeyEditModeBehavior         = "環境設定：編集モード検知時の動作"
	KeyEnableBeepSound          = "スクショ取得有無トリガー：通知音を鳴らす"
	KeyDefaultCaptureRange      = "撮影範囲：デフォルト撮影範囲"
	KeyEnableFullWindowCapture = "証跡取得：画面外・裏側の完全キャプチャ"
	KeyTopMostTOC               = "環境設定：目次管理メニューを常に最前面に表示する"
	KeyEnableMutex              = "環境設定：ツール起動時のExcel二重起動チェック"
	KeyAutoSaveExcel            = "環境設定：撮影ごとの自動保存"
	KeyEnableLogOutput          = "環境設定：ログ出力を有効にする"
	KeyLogOutputPath            = "環境設定：ログ出力先フォルダ"
	KeyLogMaxSizeMB             = "環境設定：ログフォルダ警告サイズ(MB)"

	KeyExcelOutputFolderPath = "納品物Excel保存先設定：保存先フォルダ"
)

// =========================================================================
// config_layout.ini 定数
// =========================================================================

const (
	SecLayoutToc    = "1_Excelレイアウト-目次設定"
	SecLayoutIndex  = "2_Excelレイアウト-インデックス設定"
	SecLayoutImage  = "3_Excelレイアウト-Excel画像挿入設定"
	SecLayoutLocal  = "4_画像ローカル-保存設定"
	SecLayoutDetail = "5_画像ローカル-保存詳細設定"

	KeyTocEnable     = "作成トリガー：目次シート"
	KeyTocText       = "目次シートの名前"
	KeyTocLinkEnable = "作成トリガー：目次遷移リンク"
	KeyTocLinkText   = "戻りボタン表示文字"
	KeyTocWinFix     = "ウィンドウ固定を有効"
	KeyTocFixCell    = "固定するセル"
	KeyTocRowHeight  = "新規シートの行の高さ（px）"
	KeyTocColWidth   = "新規シートの列の幅（px）"
	KeyTocZoom       = "表示倍率：拡縮率（％）"
	KeyTocHideGrid   = "表示トリガー：枠線（目盛線）を表示しない"
	KeyTocPageBreak  = "表示トリガー：ページレイアウト：改ページプレビュー"

	KeyIndexEnable     = "作成トリガー：インデックスシート自動作成"
	KeyIndexName       = "インデックスシート名"
	KeyStatusDef1      = "ステータス定義_1"
	KeyStatusDef2      = "ステータス定義_2"
	KeyStatusDef3      = "ステータス定義_3"
	KeyStatusDef4      = "ステータス定義_4"
	KeyStatusDef5      = "ステータス定義_5"
	KeyStatusDef6      = "ステータス定義_6"
	KeyStatusDef7      = "ステータス定義_7"
	KeyStatusDef8      = "ステータス定義_8"
	KeyStatusDef9      = "ステータス定義_9"
	KeyStatusDef10     = "ステータス定義_10"

	KeyImgTimestampOn = "入力トリガー：タイムスタンプ"
	KeyImgTimeFormat  = "タイムスタンプ：書式選択"
	KeyImgTimeCell    = "タイムスタンプ：初期記入セル"
	KeyImgStartCell   = "画像配置：初期貼り付けセル"
	KeyImgMargin      = "画像配置：画像同士の上下余白"
	KeyImgScale       = "画像配置：画像の縮小率"

	KeyLocSaveOn    = "保存トリガー：画像をローカルに保存"
	KeyLocFormat    = "ローカル保存：保存形式"
	KeyLocFolder    = "ローカル保存：保存先フォルダ"
	KeyLocPasteMode = "貼り付け方式：画像データ管理"

	KeyDetailWarn      = "圧縮トリガー：自動整理警告を表示"
	KeyDetailLimitCnt  = "制限枚数"
	KeyDetailLimitSize = "制限容量MB"
	KeyDetailLimitDay  = "制限日数"
	KeyDetailCompTime  = "圧縮実行タイミング"
	KeyDetailCompTool  = "圧縮ツール選択"
	KeyDetailPassOn    = "圧縮トリガー：画像圧縮パスワード設定"
	KeyDetailPassText  = "圧縮パスワード（画像圧縮）"
	KeyDetailAutoDel   = "削除トリガー：画像圧縮ファイル自動削除"
	KeyDetailDelDay    = "経過日数"
	KeyDetailDiskWarn  = "空き容量警告閾値（GB）"
)

// =========================================================================
// config_license.ini 定数
// =========================================================================

const (
	SecLicenseManagement = "1_ライセンス管理"
	KeyRemainingCount    = "残り利用可能回数"
	KeyStartDate         = "ツール利用開始日"
	KeyEndDate           = "ツール利用終了日"
	KeyLastRunDate       = "前回利用年月日"
	KeyNextAvailableTime = "次回利用可能時刻"
	KeyCheckSum          = "整合性コード"
	KeyLicenseMode       = "ライセンスモード"
	KeyLicenseKey        = "ライセンス認証キー"
	KeyMachineID         = "マシン固有ID（照合用）"
	KeySessionImageCount = "現在の撮影枚数"
	KeyVersion           = "ツールバージョン"
)

// =========================================================================
// UI・色・表示関連定数
// =========================================================================

const (
	// Excel ヘッダー背景色（目次・インデックス共通）
	ColorHeaderBgToc   = 13434879 // 薄い水色 (0x00CCFF)
	ColorHeaderBgIndex = 13421823 // 薄い紫 (0x00CC99)

	// ステータス無効時のテキスト色（グレーアウト）
	ColorDisabledText = "A0A0A0"

	// 16進数カラーパターン（正規表現用）
	RegexHexColor6 = `^[0-9A-Fa-f]{6}$`

	// ステータス定義の区切り文字
	StatusDelimiter = "|"

	// ステータス定義の要素数
	StatusPartsMin    = 3
	StatusPartsWithOn = 4

	// ステータス有効フラグ
	StatusEnabled  = "ON"
	StatusDisabled = "OFF"
	StatusLabelOn  = "ON"
	StatusLabelOff = "OFF"
)
