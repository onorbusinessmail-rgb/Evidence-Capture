package ui

import (
	"Evidence-Capture/internal/config"
	"path/filepath"
	"regexp"
	"strings"
)

// Excelのシート名として不適切な文字を置換し、31文字以内に収める
func sanitizeSheetName(name string) string {
	// 1. 禁止文字 [: \ / ? * [ ]] を "_" に置換
	re := regexp.MustCompile(`[:\\/\?\*\[\]]`)
	sanitized := re.ReplaceAllString(name, "_")

	// 2. 先頭または末尾のシングルクォートはエラーの元になるため削除
	sanitized = strings.Trim(sanitized, "'")

	// 3. 31文字制限
	runes := []rune(sanitized)
	if len(runes) > 31 {
		sanitized = string(runes[:31])
	}

	// 4. 空文字になった場合のフォールバック
	if sanitized == "" {
		sanitized = "Sheet"
	}

	return sanitized
}

// ヘルパー関数：パスからドライブ情報を取得する
func getDriveInfo(path string) string {
	vol := filepath.VolumeName(path)
	if strings.HasPrefix(vol, `\\`) {
		config.Log("INFO", "ネットワークドライブです: %s", "Network drive: %s", path)
		return "Network Share"
	}
	config.Log("INFO", "ローカルドライブです: %s", "Local drive: %s", path)
	return vol
}
