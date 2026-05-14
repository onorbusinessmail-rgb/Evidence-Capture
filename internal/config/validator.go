package config

import (
	"Evidence-Capture/internal/types"
	"Evidence-Capture/internal/winapi"

	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"
)

// ConfigValidator は設定値のバリデーションと復旧を行います。
type ConfigValidator struct {
	ModifiedMessages []string
}

// NewConfigValidator は新しいバリデーターを作成します。
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		ModifiedMessages: make([]string, 0),
	}
}

// GetValidInt は整数値の型・範囲チェックを行い、不正な場合は補正します。
func (v *ConfigValidator) GetValidInt(sec *ini.Section, keyName string, defVal, minVal, maxVal int) int {
	if !sec.HasKey(keyName) {
		return defVal
	}
	val, err := sec.Key(keyName).Int()
	if err != nil || val < minVal || val > maxVal {
		v.logAndStoreWarning(keyName, fmt.Sprintf("%d", defVal))
		return defVal
	}
	return val
}

// GetValidFloat は浮動小数点数の型・範囲チェックを行い、不正な場合は補正します。
func (v *ConfigValidator) GetValidFloat(sec *ini.Section, keyName string, defVal, minVal, maxVal float64) float64 {
	if !sec.HasKey(keyName) {
		return defVal
	}
	val, err := sec.Key(keyName).Float64()
	if err != nil || val < minVal || val > maxVal {
		v.logAndStoreWarning(keyName, fmt.Sprintf("%f", defVal))
		return defVal
	}
	return val
}

// GetValidBool はON/OFFの文字列チェックを行い、不正な場合は補正します。
func (v *ConfigValidator) GetValidBool(sec *ini.Section, keyName string, defVal bool) bool {
	if !sec.HasKey(keyName) {
		return defVal
	}
	valStr := strings.ToUpper(strings.TrimSpace(sec.Key(keyName).String()))
	if valStr != "ON" && valStr != "OFF" {
		defStr := "OFF"
		if defVal {
			defStr = "ON"
		}
		v.logAndStoreWarning(keyName, defStr)
		return defVal
	}
	return valStr == "ON"
}

// GetValidPath はパス文字列を取得し、絶対パスに解決します。
// 空白文字が含まれていたり、書き込み権限がない場合は実行ディレクトリを返します。
func (v *ConfigValidator) GetValidPath(sec *ini.Section, keyName string, defVal string) string {
	// 1. キーが存在しない場合は、引数で渡されたデフォルト値を返す
	if !sec.HasKey(keyName) {
		return defVal
	}

	val := strings.TrimSpace(sec.Key(keyName).String())

	// 2. 実行ファイルパスを基点としたパス解決の準備
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)

	// 3. 値が空（現場で「保存先が空」になった際）の処理
	if val == "" {
		// 実行ディレクトリをデフォルトとするロジックの完遂
		// 必要に応じて "storage" や "output" などのサブフォルダを足しても良い
		return exeDir
	}

	// 4. パスの正規化と絶対パス化
	cleanPath := filepath.Clean(val)
	if !filepath.IsAbs(cleanPath) {
		// 相対パスの場合は実行ディレクトリを起点にする（ポータブル性の担保）
		cleanPath = filepath.Join(exeDir, cleanPath)
	}

	// 5. 権限・存在チェック（オプション）
	// 現場の制約（レジストリ禁止・管理者権限なし）を考慮し、[cite: 1]
	// 書き込み権限がない場所が指定された場合に実行ディレクトリへフォールバックする処理を入れるとより堅牢です。
	if err := os.MkdirAll(cleanPath, 0755); err != nil {
		v.logAndStoreWarning(keyName, exeDir)
		return exeDir
	}

	return cleanPath
}

// GetValidString は文字列を取得します（トリム処理のみ）。
func (v *ConfigValidator) GetValidString(sec *ini.Section, keyName string, defVal string) string {
	if !sec.HasKey(keyName) {
		return defVal
	}
	return strings.TrimSpace(sec.Key(keyName).String())
}

// logAndStoreWarning は警告メッセージの保存とログ出力を行います。
func (v *ConfigValidator) logAndStoreWarning(keyName, defValStr string) {
	msg := fmt.Sprintf("設定.iniが手動操作され、誤った記載があったため、[%s] を初期化しました。", keyName)
	v.ModifiedMessages = append(v.ModifiedMessages, msg)
	Log("WARN", "[Config] %s の値が無効なためデフォルト値 %s に復旧しました", "Invalid value for "+keyName, keyName, defValStr)
}

// ShowWarningsAndSave は警告が存在する場合にダイアログを表示し、正常な値で設定ファイルを上書き保存します。
func (v *ConfigValidator) ShowWarningsAndSave(cfg types.AppConfig) {
	if len(v.ModifiedMessages) > 0 {
		msg := strings.Join(v.ModifiedMessages, "\n")
		// 表示領域を考慮し、10件を超える場合は省略表示にする
		if len(v.ModifiedMessages) > 10 {
			msg = strings.Join(v.ModifiedMessages[:10], "\n") + "\n...他多数"
		}
		winapi.ShowDialog("設定ファイルの自動復旧", msg, winapi.MB_ICONWARNING)

		// 復旧した値を設定ファイルに上書き保存し、以後の起動を正常にする
		SaveAppConfig(cfg)
	}
}

// HasWarnings は復旧による警告メッセージが存在するか判定します。
func (v *ConfigValidator) HasWarnings() bool {
	return len(v.ModifiedMessages) > 0
}

// GetWarningMessage は警告メッセージを連結して返します。
// ※ダイアログの表示や保存の実行は呼び出し側のUI層で行います。
func (v *ConfigValidator) GetWarningMessage() string {
	if len(v.ModifiedMessages) == 0 {
		return ""
	}
	msg := strings.Join(v.ModifiedMessages, "\n")
	if len(v.ModifiedMessages) > 10 {
		msg = strings.Join(v.ModifiedMessages[:10], "\n") + "\n...他多数"
	}
	return msg
}
