// Package config は設定ファイル（INI）の読み書きロジックを提供します。
// このパッケージは walk や winapi に依存せず、error を返すことで
// ダイアログ表示の責務を呼び出し側（cmd/app や internal/ui）に委譲します。
package config

import (
	"errors"
	"fmt"

	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"Evidence-Capture/internal/multilang"
	"Evidence-Capture/internal/types"

	"gopkg.in/ini.v1"
)

// =========================================================================
// パッケージレベル変数
// =========================================================================

var (
	// 正規表現を事前にコンパイル（関数外に定義）
	cellRegex = regexp.MustCompile(`^([A-Z]{1,3})([1-9][0-9]{0,6})$`)

	// Replacerを事前に定義（関数外に定義）
	cellReplacer = strings.NewReplacer(
		"０", "0", "１", "1", "２", "2", "３", "3", "４", "4",
		"５", "5", "６", "6", "７", "7", "８", "8", "９", "9",
		"Ａ", "A", "Ｂ", "B", "Ｃ", "C", "Ｄ", "D", "Ｅ", "E",
		"Ｆ", "F", "Ｇ", "G", "Ｈ", "H", "Ｉ", "I", "Ｊ", "J",
		"Ｋ", "K", "Ｌ", "L", "Ｍ", "M", "Ｎ", "N", "Ｏ", "O",
		"Ｐ", "P", "Ｑ", "Q", "Ｒ", "R", "Ｓ", "S", "Ｔ", "T",
		"Ｕ", "U", "Ｖ", "V", "Ｗ", "W", "Ｘ", "X", "Ｙ", "Y", "Ｚ", "Z",
	)

	// CurrentConfig はアプリ全体で共有される現在の設定値です。
	CurrentConfig types.AppConfig

	// Version はビルド時に -ldflags "-X ..." で注入されるバージョン情報です。
	Version string
)

// INIファイルのパス（init で初期化）
var (
	ConfigDir string
	AppTempDir string
	IniFile1  string
	IniFile2  string
	IniFile3  string
)

func init() {
	// 実行ファイル（.exe）の絶対パスを取得し、常に同じ場所を基準にする
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		ConfigDir = filepath.Join(exeDir, "setting") // 絶対パス化
		AppTempDir = filepath.Join(exeDir, "temp")    // 絶対パス化
	} else {
		ConfigDir = "setting"
		AppTempDir = "temp"
	}

	// 各INIファイルの絶対パスを構築
	IniFile1 = filepath.Join(ConfigDir, "config_evidence.ini")
	IniFile2 = filepath.Join(ConfigDir, "config_layout.ini")
	IniFile3 = filepath.Join(ConfigDir, "config_license.ini")

	// INIファイル保存時の自動パディング（揃え）を無効化
	ini.PrettyFormat = false
	ini.PrettyEqual = false
}

// =========================================================================
// 設定読み込みメイン処理
// =========================================================================

// LoadAppConfig は複数のINIファイルから設定値を読み込み、AppConfig構造体に展開します。
func LoadAppConfig() (types.AppConfig, error) {
	opts := ini.LoadOptions{
		IgnoreContinuation: true, // '\'を行継続文字として扱わない
	}

	cfgEvidence, err := ini.LoadSources(opts, IniFile1)
	if err != nil {
		Log("ERROR", "config_evidence.iniの読み込みに失敗しました: %v", "config_evidence.ini loading failed: %v", err)
		return types.AppConfig{}, fmt.Errorf("config_evidence.ini の読み込みに失敗しました: %w", err)
	}
	cfgLayout, err := ini.LoadSources(opts, IniFile2)
	if err != nil {
		Log("ERROR", "config_layout.iniの読み込みに失敗しました: %v", "config_layout.ini loading failed: %v", err)
		return types.AppConfig{}, fmt.Errorf("config_layout.ini の読み込みに失敗しました: %w", err)
	}
	cfgLicense, err := ini.LoadSources(opts, IniFile3)
	if err != nil {
		Log("ERROR", "config_license.iniの読み込みに失敗しました: %v", "config_license.ini loading failed: %v", err)
		return types.AppConfig{}, fmt.Errorf("config_license.ini の読み込みに失敗しました: %w", err)
	}

	// 1. まずデフォルト値で初期化
	config := types.AppConfig{
		EnableExitShortcut:       true,
		EnableAutoTimeout:        true,
		WaitTimeSeconds:          3.0,
		EnableBeep:               true,
		EditModeBehavior:         1,
		DefaultCaptureRange:      1,
		EnableFullWindowCapture:  false,
		EdgeSensitivity:          0.01,
		MinimizeExcel:            true,
		AutoSwitchTimeoutMinutes: 15,
		Language:                 "",
		EnableMutex:              1,
		MaxCellSearchRow:         100000,
		LogMaxSizeMB:             50,

		EnableAutoToc:      true,
		TocSheetName:       "目次",
		EnableTocLink:      true,
		ReturnButtonText:   "← 目次へ",
		EnableWindowFreeze: true,
		FreezePaneCell:     "B2",
		DefaultRowHeight:   30.0,
		DefaultColWidth:    30.0,

		Margin:               5,
		ImageInsertStartCell: "B4",
		ImageScale:           0.8,
		EnableTopMostTOC:     true,
	}

	validator := NewConfigValidator()

	// --- config_evidence.ini ---
	if cfgEvidence != nil {
		// 1-1_証跡取得-トリガー設定
		secTrigger := cfgEvidence.Section(SecBaseTrigger)
		config.EnableTriggerTopLeft = validator.GetValidBool(secTrigger, KeyTriggerTopLeft, true)
		config.EnableTriggerTopRight = validator.GetValidBool(secTrigger, KeyTriggerTopRight, true)
		config.EnableTriggerRightDbl = validator.GetValidBool(secTrigger, KeyTriggerRightDbl, true)
		config.EnableExitShortcut = validator.GetValidBool(secTrigger, KeyExitCtrlShiftX, true)
		config.EnableExitWinKey = validator.GetValidBool(secTrigger, KeyExitWinKey, true)
		config.EnableExitLeftDbl = validator.GetValidBool(secTrigger, KeyExitLeftDbl, true)
		config.EnableAutoTimeout = validator.GetValidBool(secTrigger, KeyExitTimeout, true)
		config.AutoSwitchTimeoutMinutes = validator.GetValidInt(secTrigger, KeyAutoSwitchTimeoutMinutes, 15, 1, 60)

		// 1-2_証跡取得-撮影環境設定
		secCapture := cfgEvidence.Section(SecBaseCapture)
		config.Language = validator.GetValidString(secCapture, KeyLanguage, "")
		config.DefaultCaptureRange = validator.GetValidInt(secCapture, KeyDefaultCaptureRange, 1, 1, 2)
		config.EnableFullWindowCapture = validator.GetValidBool(secCapture, KeyEnableFullWindowCapture, false)
		config.EnableBeep = validator.GetValidBool(secCapture, KeyEnableBeepSound, true)
		config.EdgeSensitivity = validator.GetValidFloat(secCapture, KeyEdgeSensitivity, 0.01, 0.001, 0.1)
		config.WaitTimeSeconds = validator.GetValidFloat(secCapture, KeyWaitTimeSeconds, 3.0, 0.1, 10.0)

		// 言語設定を同期
		multilang.SetLanguage(config.Language)

		// 1-3_証跡取得-Excelファイル制御設定
		secFile := cfgEvidence.Section(SecBaseFile)
		config.ExcelOutputPath = validator.GetValidString(secFile, KeyExcelOutputPath, "")
		config.MinimizeExcel = validator.GetValidBool(secFile, KeyMinimizeExcel, true)
		config.EditModeBehavior = validator.GetValidInt(secFile, KeyEditModeBehavior, 1, 1, 4)
		config.EnableTopMostTOC = validator.GetValidBool(secFile, KeyTopMostTOC, true)
		config.EnableMutex = validator.GetValidInt(secFile, KeyEnableMutex, 1, 1, 3)
		config.AutoSaveExcel = validator.GetValidInt(secFile, KeyAutoSaveExcel, 1, 1, 2)
		config.EnableLogOutput = validator.GetValidBool(secFile, KeyEnableLogOutput, true)
		config.LogOutputPath = validator.GetValidString(secFile, KeyLogOutputPath, "")
		config.LogMaxSizeMB = validator.GetValidInt(secFile, KeyLogMaxSizeMB, 50, 1, 10000)

		// ログ出力先が空（未設定）の場合は、デフォルトで実行ファイル階層の logs フォルダをターゲットにする
		if config.LogOutputPath == "" {
			exePath, _ := os.Executable()
			config.LogOutputPath = filepath.Join(filepath.Dir(exePath), "logs")
		} else {
			// パスが指定されている場合は絶対パスに解決
			if !filepath.IsAbs(config.LogOutputPath) {
				exePath, _ := os.Executable()
				config.LogOutputPath = filepath.Join(filepath.Dir(exePath), config.LogOutputPath)
			}
		}
	}
	Log("INFO", "設定ファイルの読み込み完了: %s", "setting file loading completed: %s", IniFile1)

	// --- config_layout.ini ---
	if cfgLayout != nil {
		sec1 := cfgLayout.Section(SecLayoutToc)
		config.EnableAutoToc = validator.GetValidBool(sec1, KeyTocEnable, true)
		config.TocSheetName = validator.GetValidString(sec1, KeyTocText, "目次")
		config.EnableTocLink = validator.GetValidBool(sec1, KeyTocLinkEnable, true)
		config.ReturnButtonText = validator.GetValidString(sec1, KeyTocLinkText, "← 目次へ")
		config.EnableWindowFreeze = validator.GetValidBool(sec1, KeyTocWinFix, true)
		config.FreezePaneCell = validator.GetValidString(sec1, KeyTocFixCell, "B2")
		config.DefaultRowHeight = validator.GetValidFloat(sec1, KeyTocRowHeight, 30.0, 1, 100)
		config.DefaultColWidth = validator.GetValidFloat(sec1, KeyTocColWidth, 30.0, 1, 100)
		config.ZoomPercent = validator.GetValidInt(sec1, KeyTocZoom, 80, 1, 100)
		config.HideGridlines = validator.GetValidBool(sec1, KeyTocHideGrid, true)
		config.PageBreakPreview = validator.GetValidBool(sec1, KeyTocPageBreak, true)

		sec2 := cfgLayout.Section(SecLayoutIndex)
		config.MakeIndex = validator.GetValidBool(sec2, KeyIndexEnable, true)
		config.IndexSheetName = validator.GetValidString(sec2, KeyIndexName, "インデックス")

		config.StatusDefinitions = []string{}
		for i := 1; ; i++ {
			keyName := fmt.Sprintf("ステータス定義_%d", i)
			if !sec2.HasKey(keyName) {
				break
			}
			config.StatusDefinitions = append(config.StatusDefinitions, sec2.Key(keyName).String())
		}

		sec3 := cfgLayout.Section(SecLayoutImage)
		config.EnableTimestamp = validator.GetValidBool(sec3, KeyImgTimestampOn, true)
		config.TimestampFormat = validator.GetValidInt(sec3, KeyImgTimeFormat, 1, 1, 3)
		config.TimestampCell = validator.GetValidString(sec3, KeyImgTimeCell, "B3")
		config.ImageInsertStartCell = validator.GetValidString(sec3, KeyImgStartCell, "C4")
		config.Margin = validator.GetValidInt(sec3, KeyImgMargin, 5, 0, 100)
		config.ImageScale = validator.GetValidFloat(sec3, KeyImgScale, 0.8, 0.1, 2.0)

		sec4 := cfgLayout.Section(SecLayoutLocal)
		config.LocalSaveOn = validator.GetValidBool(sec4, KeyLocSaveOn, false)
		config.LocalSaveFormat = validator.GetValidInt(sec4, KeyLocFormat, 1, 1, 3)
		config.LocalSaveFolder = validator.GetValidString(sec4, KeyLocFolder, "")
		config.PasteMode = validator.GetValidInt(sec4, KeyLocPasteMode, 1, 1, 3)
	}
	Log("INFO", "設定ファイルの読み込み完了: %s", "setting file loading completed: %s", IniFile2)

	// --- config_license.ini ---
	if cfgLicense != nil {
		sec := cfgLicense.Section(SecLicenseManagement)
		config.RemainingCount = sec.Key(KeyRemainingCount).MustInt(0)
		config.StartDate = sec.Key(KeyStartDate).String()
		config.EndDate = sec.Key(KeyEndDate).String()
		config.LastRunDate = sec.Key(KeyLastRunDate).String()
		config.NextAvailableTime = sec.Key(KeyNextAvailableTime).String()
		config.CheckSum = sec.Key(KeyCheckSum).String()
		config.LicenseMode = sec.Key(KeyLicenseMode).String()
		config.LicenseKey = sec.Key(KeyLicenseKey).String()
		config.MachineID = sec.Key(KeyMachineID).String()
		config.SessionImageCount = sec.Key(KeySessionImageCount).MustInt(0)
		config.Version = sec.Key(KeyVersion).String()

		// INIにバージョンがない、または実行バイナリと異なる場合は現在のバイナリのバージョンを優先
		if config.Version == "" && Version != "" {
			config.Version = Version
		}
	}
	Log("INFO", "設定ファイルの読み込み完了: %s", "setting file loading completed: %s", IniFile3)

	return config, nil
}

// =========================================================================
// ファイル管理・バリデーション・補助機能
// =========================================================================

// EnsureConfigFiles は設定フォルダとデフォルトINIファイルの存在を保証します。
// 失敗した場合は error を返します（ダイアログ表示は呼び出し側の責務）。
func EnsureConfigFiles() error {
	// 1. 設定保存用ディレクトリの作成
	if err := os.MkdirAll(ConfigDir, 0755); err != nil {
		Log("ERROR", "設定フォルダの作成に失敗しました: %v", "Failed to create config directory: %v", err)
		return fmt.Errorf("設定フォルダ(setting)の作成に失敗しました: %w", err)
	}

	// 2. 一時保存用ディレクトリの作成
	if err := os.MkdirAll(AppTempDir, 0755); err != nil {
		Log("ERROR", "一時フォルダの作成に失敗しました: %v", "Failed to create temp directory: %v", err)
		return fmt.Errorf("一時フォルダ(temp)の作成に失敗しました: %w", err)
	}

	// 2. 各INIファイルの存在チェックとデフォルト書き出し
	templates := map[string]string{
		IniFile1: DefaultIni1,
		IniFile2: DefaultIni2,
		IniFile3: DefaultIni3,
	}

	for path, template := range templates {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte(template), 0666); err != nil {
				Log("ERROR", "設定ファイル(%s)の作成に失敗しました。", "Failed to create config file: %v", path)
				return fmt.Errorf("%s の作成に失敗しました。", path)
			}
		}
	}
	return nil
}

// CheckWritePermission は実行ディレクトリに書き込み権限があるかチェックします。
func CheckWritePermission() error {
	exePath, err := os.Executable()
	if err != nil {
		Log("ERROR", "実行ファイルパスの取得失敗: %v", "Failed to get executable path: %v", err)
		return err
	}
	exeDir := filepath.Dir(exePath)

	testFile := filepath.Join(exeDir, ".write_test")
	f, err := os.Create(testFile)
	if err != nil {
		Log("ERROR", "書き込み権限エラー: %v", "Write permission error: %v", err)
		return fmt.Errorf("フォルダ（%s）に書き込み権限がありません: %w", exeDir, err)
	}
	f.Close()
	os.Remove(testFile)
	return nil
}

// ValidateDate は文字列が「YYYY-MM-DD」形式の有効な日付であるかチェックします。
func ValidateDate(input string) bool {
	_, err := time.Parse("2006-01-02", strings.TrimSpace(input))
	return err == nil
}

// ValidateRange は指定された数値が許容範囲内（最小値～最大値）に収まっているか判定します。
func ValidateRange(val float64, min, max float64) bool {
	return val >= min && val <= max
}

// IsON はON/OFFを判定します。
func IsON(val string) bool {
	return strings.ToUpper(strings.TrimSpace(val)) == "ON"
}

// ValidateAndNormalizeCell はセル形式を検証し、正規化します。
func ValidateAndNormalizeCell(input string) (string, bool) {
	if input == "" {
		Log("ERROR", "セル検証: 空白", "Cell validation: empty")
		return "", false
	}

	// 1. 正規化（全角→半角、空白除去、大文字化）
	res := strings.ToUpper(strings.TrimSpace(cellReplacer.Replace(input)))

	// 2. 正規表現で「形式」をチェックしつつ、列と行を分離
	matches := cellRegex.FindStringSubmatch(res)
	if len(matches) != 3 {
		Log("ERROR", "セル検証: 形式不正 %s", "Cell validation: invalid format %s", input)
		return input, false
	}

	columnPart := matches[1]
	rowPart := matches[2]

	// 3. 行番号がExcelの最大値（1,048,576）を超えていないかチェック
	rowNum, err := strconv.Atoi(rowPart)
	if err != nil || rowNum > 1048576 {
		Log("ERROR", "セル検証: 行番号不正 %s", "Cell validation: invalid row number %s", input)
		return input, false
	}

	// 4. 列名が最大（XFD）を超えていないかチェック
	if len(columnPart) == 3 && columnPart > "XFD" {
		Log("ERROR", "セル検証: 列番号不正 %s", "Cell validation: invalid column number %s", input)
		return input, false
	}

	return res, true
}

func SaveLicenseConfig(c types.AppConfig) error {
	ini.PrettyFormat = false
	ini.PrettyEqual = false
	cfg, err := ini.Load(IniFile3)
	if err != nil {
		return err
	}
	sec := cfg.Section(SecLicenseManagement)
	sec.Key(KeyRemainingCount).SetValue(strconv.Itoa(c.RemainingCount))
	sec.Key(KeyStartDate).SetValue(c.StartDate)
	sec.Key(KeyEndDate).SetValue(c.EndDate)
	sec.Key(KeyLastRunDate).SetValue(c.LastRunDate)
	sec.Key(KeyNextAvailableTime).SetValue(c.NextAvailableTime)
	sec.Key(KeyCheckSum).SetValue(c.CheckSum)
	sec.Key(KeyLicenseMode).SetValue(c.LicenseMode)
	sec.Key(KeyLicenseKey).SetValue(c.LicenseKey)
	sec.Key(KeyMachineID).SetValue(c.MachineID)
	sec.Key(KeySessionImageCount).SetValue(strconv.Itoa(c.SessionImageCount))
	sec.Key(KeyVersion).SetValue(c.Version)

	return cfg.SaveTo(IniFile3)
}

// SaveAppConfig は AppConfig 構造体の内容を各 INI ファイルに保存します。
// ファイル保存が成功した直後、メモリ上の CurrentConfig も新しい値で上書きします。
func SaveAppConfig(c types.AppConfig) error {
	ini.PrettyFormat = false
	ini.PrettyEqual = false
	opts := ini.LoadOptions{
		IgnoreContinuation: true,
	}

	// 1. config_evidence.ini の保存
	if err := saveEvidenceConfig(c, opts); err != nil {
		return err
	}

	// 2. config_layout.ini の保存
	if err := saveLayoutConfig(c, opts); err != nil {
		return err
	}

	// 3. config_license.ini の保存
	if err := SaveLicenseConfig(c); err != nil {
		return err
	}

	// 4. ファイル保存成功後、メモリ上の CurrentConfig を更新
	// これにより呼び出し側での再読み込み（LoadAppConfig）を不要にする
	CurrentConfig = c
	Log("INFO", "設定をファイルに保存し、メモリを更新しました", "Config saved and memory updated")

	return nil
}

func saveEvidenceConfig(c types.AppConfig, opts ini.LoadOptions) error {
	ini.PrettyFormat = false
	ini.PrettyEqual = false
	cfg, err := ini.LoadSources(opts, IniFile1)
	if err != nil {
		return err
	}

	// 1-1_証跡取得-トリガー設定
	secTrigger := cfg.Section(SecBaseTrigger)
	secTrigger.Key(KeyTriggerTopLeft).SetValue(boolToONOFF(c.EnableTriggerTopLeft))
	secTrigger.Key(KeyTriggerTopRight).SetValue(boolToONOFF(c.EnableTriggerTopRight))
	secTrigger.Key(KeyTriggerRightDbl).SetValue(boolToONOFF(c.EnableTriggerRightDbl))
	secTrigger.Key(KeyExitCtrlShiftX).SetValue(boolToONOFF(c.EnableExitShortcut))
	secTrigger.Key(KeyExitWinKey).SetValue(boolToONOFF(c.EnableExitWinKey))
	secTrigger.Key(KeyExitLeftDbl).SetValue(boolToONOFF(c.EnableExitLeftDbl))
	secTrigger.Key(KeyExitTimeout).SetValue(boolToONOFF(c.EnableAutoTimeout))
	secTrigger.Key(KeyAutoSwitchTimeoutMinutes).SetValue(strconv.Itoa(c.AutoSwitchTimeoutMinutes))

	// 1-2_証跡取得-撮影環境設定
	secCapture := cfg.Section(SecBaseCapture)
	secCapture.Key(KeyLanguage).SetValue(c.Language)
	secCapture.Key(KeyDefaultCaptureRange).SetValue(strconv.Itoa(c.DefaultCaptureRange))
	secCapture.Key(KeyEnableFullWindowCapture).SetValue(boolToONOFF(c.EnableFullWindowCapture))
	secCapture.Key(KeyEnableBeepSound).SetValue(boolToONOFF(c.EnableBeep))
	secCapture.Key(KeyEdgeSensitivity).SetValue(fmt.Sprintf("%.3f", c.EdgeSensitivity))
	secCapture.Key(KeyWaitTimeSeconds).SetValue(fmt.Sprintf("%.1f", c.WaitTimeSeconds))

	// 1-3_証跡取得-Excelファイル制御設定
	secFile := cfg.Section(SecBaseFile)
	secFile.Key(KeyExcelOutputPath).SetValue(c.ExcelOutputPath)
	secFile.Key(KeyMinimizeExcel).SetValue(boolToONOFF(c.MinimizeExcel))
	secFile.Key(KeyEditModeBehavior).SetValue(strconv.Itoa(c.EditModeBehavior))
	secFile.Key(KeyTopMostTOC).SetValue(boolToONOFF(c.EnableTopMostTOC))
	secFile.Key(KeyEnableMutex).SetValue(strconv.Itoa(c.EnableMutex))
	secFile.Key(KeyAutoSaveExcel).SetValue(strconv.Itoa(c.AutoSaveExcel))
	secFile.Key(KeyEnableLogOutput).SetValue(boolToONOFF(c.EnableLogOutput))
	secFile.Key(KeyLogOutputPath).SetValue(c.LogOutputPath)
	secFile.Key(KeyLogMaxSizeMB).SetValue(strconv.Itoa(c.LogMaxSizeMB))

	return cfg.SaveTo(IniFile1)
}

// NewMultiLangError は現在の言語設定に基づいてエラーを生成します。
func NewMultiLangError(jp, en string) error {
	if CurrentConfig.Language == "English" {
		return errors.New(en)
	}
	return errors.New(jp)
}

func saveLayoutConfig(c types.AppConfig, opts ini.LoadOptions) error {
	ini.PrettyFormat = false
	ini.PrettyEqual = false
	cfg, err := ini.LoadSources(opts, IniFile2)
	if err != nil {
		return err
	}

	// [1_Excelレイアウト-目次設定]
	sec1 := cfg.Section(SecLayoutToc)
	sec1.Key(KeyTocEnable).SetValue(boolToONOFF(c.EnableAutoToc))
	sec1.Key(KeyTocText).SetValue(c.TocSheetName)
	sec1.Key(KeyTocLinkEnable).SetValue(boolToONOFF(c.EnableTocLink))
	sec1.Key(KeyTocLinkText).SetValue(c.ReturnButtonText)
	sec1.Key(KeyTocWinFix).SetValue(boolToONOFF(c.EnableWindowFreeze))
	sec1.Key(KeyTocFixCell).SetValue(c.FreezePaneCell)
	sec1.Key(KeyTocRowHeight).SetValue(fmt.Sprintf("%.1f", c.DefaultRowHeight))
	sec1.Key(KeyTocColWidth).SetValue(fmt.Sprintf("%.1f", c.DefaultColWidth))
	sec1.Key(KeyTocZoom).SetValue(strconv.Itoa(c.ZoomPercent))
	sec1.Key(KeyTocHideGrid).SetValue(boolToONOFF(c.HideGridlines))
	sec1.Key(KeyTocPageBreak).SetValue(boolToONOFF(c.PageBreakPreview))

	// [2_Excelレイアウト-インデックス設定]
	sec2 := cfg.Section(SecLayoutIndex)
	sec2.Key(KeyIndexEnable).SetValue(boolToONOFF(c.MakeIndex))
	sec2.Key(KeyIndexName).SetValue(c.IndexSheetName)

	// ステータス定義のクリアと再設定
	for _, keyName := range sec2.KeyStrings() {
		if strings.HasPrefix(keyName, "ステータス定義_") {
			sec2.DeleteKey(keyName)
		}
	}
	for i, def := range c.StatusDefinitions {
		keyName := fmt.Sprintf("ステータス定義_%d", i+1)
		sec2.Key(keyName).SetValue(def)
	}

	// [3_Excelレイアウト-Excel画像挿入設定]
	sec3 := cfg.Section(SecLayoutImage)
	sec3.Key(KeyImgTimestampOn).SetValue(boolToONOFF(c.EnableTimestamp))
	sec3.Key(KeyImgTimeFormat).SetValue(strconv.Itoa(c.TimestampFormat))
	sec3.Key(KeyImgTimeCell).SetValue(c.TimestampCell)
	sec3.Key(KeyImgStartCell).SetValue(c.ImageInsertStartCell)
	sec3.Key(KeyImgMargin).SetValue(strconv.Itoa(c.Margin))
	sec3.Key(KeyImgScale).SetValue(fmt.Sprintf("%.2f", c.ImageScale))

	// [4_画像ローカル-保存設定]
	sec4 := cfg.Section(SecLayoutLocal)
	sec4.Key(KeyLocSaveOn).SetValue(boolToONOFF(c.LocalSaveOn))
	sec4.Key(KeyLocFormat).SetValue(strconv.Itoa(c.LocalSaveFormat))
	// 行継続文字バグ回避のため末尾のスラッシュを除去
	folder := strings.TrimRight(strings.TrimSpace(c.LocalSaveFolder), `\/`)
	sec4.Key(KeyLocFolder).SetValue(folder)
	sec4.Key(KeyLocPasteMode).SetValue(strconv.Itoa(c.PasteMode))

	// [5_画像ローカル-保存詳細設定]
	sec5 := cfg.Section(SecLayoutDetail)
	sec5.Key(KeyDetailWarn).SetValue(boolToONOFF(c.AutoOrganizeWarn))
	sec5.Key(KeyDetailLimitCnt).SetValue(strconv.Itoa(c.LimitCount))
	sec5.Key(KeyDetailLimitSize).SetValue(strconv.Itoa(c.LimitSizeMB))
	sec5.Key(KeyDetailLimitDay).SetValue(strconv.Itoa(c.LimitDays))
	sec5.Key(KeyDetailCompTime).SetValue(strconv.Itoa(c.CompTiming))
	sec5.Key(KeyDetailCompTool).SetValue(strconv.Itoa(c.CompTool))
	sec5.Key(KeyDetailPassOn).SetValue(boolToONOFF(c.EnableCompPass))
	sec5.Key(KeyDetailPassText).SetValue(c.CompPass)
	sec5.Key(KeyDetailAutoDel).SetValue(boolToONOFF(c.EnableAutoDelZip))
	sec5.Key(KeyDetailDelDay).SetValue(strconv.Itoa(c.DelZipDays))
	sec5.Key(KeyDetailDiskWarn).SetValue(fmt.Sprintf("%.2f", c.DiskFreeThresholdGB))

	return cfg.SaveTo(IniFile2)
}

func boolToONOFF(b bool) string {
	if b {
		return "ON"
	}
	return "OFF"
}

// T は現在の言語設定に応じて日本語または英語の文字列を返します。
func T(jp, en string) string {
	return multilang.T(jp, en)
}
