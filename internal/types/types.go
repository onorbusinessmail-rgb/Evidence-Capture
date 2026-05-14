// Package types はすべてのパッケージから参照される共有データ型を定義します。
// このパッケージは外部ライブラリや他の internal パッケージに依存しません。
package types

import (
	"image"

	"github.com/lxn/walk"
)

// =========================================================================
// アプリケーション設定構造体
// =========================================================================

// AppConfig はツール全体の動作設定を集約する構造体です。
type AppConfig struct {
	// Evidence関連 (config_evidence.ini)
	// --- [1_基本動作設定] ---
	// --- 撮影トリガー設定 ---
	EnableTriggerTopLeft  bool // 撮影トリガー：画面左上へのマウス移動
	EnableTriggerTopRight bool // 撮影トリガー：画面右上へのマウス移動
	EnableTriggerRightDbl bool // 撮影トリガー：画面右端でのダブルクリック

	// --- 終了トリガー設定 ---
	EnableExitShortcut       bool // 終了トリガー：Ctrl+Shift+Xキーによる強制終了
	EnableExitWinKey         bool // 終了トリガー：Windowsキー押下による中断
	EnableExitLeftDbl        bool // 終了トリガー：画面左端でのダブルクリック
	EnableAutoTimeout        bool // 終了トリガー：タイムアウト
	AutoSwitchTimeoutMinutes int  // 環境設定：タイムアウトまでの時間（分）
	Language                 string // 環境設定：言語 (Japanese / English)

	// --- 数値入力・パス設定 ---
	WaitTimeSeconds     float64 // 撮影制御：撮影実行後の待機時間（秒）
	EdgeSensitivity     float64 // 環境設定：画面端の判定感度（0.01単位など）
	EditModeBehavior    int     // 環境設定：Excel編集モード検知時の動作
	DefaultCaptureRange int     // 撮影範囲：デフォルトの範囲設定
	EnableFullWindowCapture bool // 撮影範囲：画面外・裏側ウィンドウの完全キャプチャを有効にする

	ExcelOutputPath string // 環境設定：出力先となるExcelファイルのフルパス

	// --- その他の動作設定 ---
	EnableBeep       bool // スクショ取得有無トリガー：通知音を鳴らす
	MinimizeExcel    bool // 環境設定：起動時にExcelを最小化する
	EnableTopMostTOC bool // 環境設定：目次管理メニューを常に最前面に表示する
	EnableMutex      int  // 環境設定：ツール起動時のExcel二重起動チェック
	AutoSaveExcel    int  // 環境設定：撮影ごとの自動保存 (1:無効, 2:有効)
	MaxCellSearchRow int  // 貼り付け位置の最大検索行数
	EnableLogOutput  bool // 環境設定：ログ出力を有効にする
	LogOutputPath    string // 環境設定：ログ出力先フォルダ
	LogMaxSizeMB     int    // 環境設定：ログフォルダの最大サイズ（警告閾値）

	// Layout関連 (config_layout.ini)
	// --- [1_Excelレイアウト-目次設定] ---
	EnableAutoToc      bool    // 作成トリガー：目次シート
	TocSheetName       string  // 目次シートの名前
	EnableTocLink      bool    // 作成トリガー：目次遷移リンク
	ReturnButtonText   string  // 戻りボタン表示文字
	EnableWindowFreeze bool    // ウィンドウ固定を有効
	FreezePaneCell     string  // 固定するセル
	DefaultRowHeight   float64 // 新規シートの行の高さ（px）
	DefaultColWidth    float64 // 新規シートの列の幅（px）
	ZoomPercent        int     // 表示倍率：拡縮率（％）
	HideGridlines      bool    // 表示トリガー：枠線（目盛線）を表示しない
	PageBreakPreview   bool    // 表示トリガー：ページレイアウト：改ページプレビュー

	// --- [2_Excelレイアウト-インデックス設定] ---
	MakeIndex         bool     // 作成トリガー：インデックスシート自動作成
	IndexSheetName    string   // インデックスシート名
	StatusDefinitions []string // ステータス定義_1~10

	// --- [3_Excelレイアウト-Excel画像挿入設定] ---
	EnableTimestamp      bool    // 入力トリガー：タイムスタンプ
	TimestampFormat      int     // タイムスタンプ：書式選択
	TimestampCell        string  // タイムスタンプ：初期記入セル
	ImageInsertStartCell string  // 画像配置：初期貼り付けセル
	Margin               int     // 画像配置：画像同士の上下余白
	ImageScale           float64 // 画像配置：画像の縮小率

	// --- [4_画像ローカル-保存設定] ---
	LocalSaveOn     bool   // 保存トリガー：画像をパソコン内に保存
	LocalSaveFormat int    // ローカル保存：保存形式
	LocalSaveFolder string // ローカル保存：保存先フォルダ
	PasteMode       int    // 貼り付け方式：画像データ管理

	// --- [5_画像ローカル-保存詳細設定] ---
	AutoOrganizeWarn    bool    // 圧縮トリガー：自動整理警告を表示
	LimitCount          int     // 制限枚数
	LimitSizeMB         int     // 制限容量MB
	LimitDays           int     // 制限日数
	CompTiming          int     // 圧縮実行タイミング
	CompTool            int     // 圧縮ツール選択
	EnableCompPass      bool    // 圧縮トリガー：画像圧縮パスワード設定
	CompPass            string  // 圧縮パスワード（画像圧縮）
	EnableAutoDelZip    bool    // 削除トリガー：画像圧縮ファイル自動削除
	DelZipDays          int     // 経過日数
	DiskFreeThresholdGB float64 // 空き容量警告閾値（GB）

	// License関連 (config_license.ini)
	// --- [1_ライセンス管理] ---
	RemainingCount    int    // 残り利用可能回数
	StartDate         string // ツール利用開始日
	EndDate           string // ツール利用終了日
	LastRunDate       string // 前回利用年月日
	NextAvailableTime string // 次回利用可能時刻
	CheckSum          string // 整合性コード
	LicenseMode       string // ライセンスモード
	LicenseKey        string // ライセンス認証キー (v2.0.0で追加予定のため保持)
	MachineID         string // マシン固有ID（照合用）
	Version           string // ツールバージョン

	// ランタイム状態（保存されない）
	IsRestricted      bool // 制限モードフラグ
	SessionImageCount int  // 現在の制限モードセッションでの撮影枚数

	// 状態管理
}

// LicenseDat はシステム暗号化ファイルに保存されるデータ構造体です。
type LicenseDat struct {
	StartDate         string
	EndDate           string
	LastRunDate       string
	RemainingCount    int
	NextAvailableTime string
	SessionImageCount int
}

// ImageTreeItem はTreeViewの各ノード（シートまたは画像）を表す構造体です
type ImageTreeItem struct {
	Parent   *ImageTreeItem
	Text     string           // 表示名
	IsSheet  bool             // シートノードかどうか
	IsSystem bool             // 目次・インデックスシートかどうか
	Path     string           // 画像ファイルの場合のフルパス
	Children []*ImageTreeItem // 子ノードのリスト
}

// SheetInfo に画像数を保持するフィールドを追加（必要に応じて既存構造体を拡張）
type SheetInfo struct {
	Name           string
	IsVisible      bool
	ImageCount     int
	ExistingImages []string
}

// =========================================================================
// UI共有型
// =========================================================================

// ComboItem はコンボボックスの表示名と内部値をセットで保持します。
type ComboItem struct {
	Name  string
	Value int
}

// ComboModel は walk.ComboBox に構造体リストを表示するためのモデルです。
type ComboModel struct {
	walk.ListModelBase
	Items []ComboItem
}

// ItemCount は ComboModel のアイテム数を返します。
func (m *ComboModel) ItemCount() int {
	return len(m.Items)
}

// Value は ComboModel の指定インデックスの表示名を返します。
func (m *ComboModel) Value(index int) interface{} {
	return m.Items[index].Name
}

// IndexStatus はインデックスシートのステータス定義を保持します。
type IndexStatus struct {
	ID    string
	Name  string
	Color string
}

// ValidationError は設定項目の検証エラー情報を保持します。
type ValidationError struct {
	FieldName string
	Message   string
}

// =========================================================================
// キャプチャ・保存フロー用インターフェース
// =========================================================================

// ImageSource は画像データの取得元を抽象化します。
type ImageSource interface {
	Fetch() (image.Image, error)
}

// ImageDestination は画像データの保存・出力先を抽象化します。
type ImageDestination interface {
	Store(img image.Image, path string) error
}

// ClipboardProvider はクリップボードへの出力を抽象化し、循環参照を回避します。
type ClipboardProvider interface {
	SetImage(img image.Image) error
}
