package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	files, _ := filepath.Glob("internal/ui/*.go")
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		
		// BOM削除 (UTF-8)
		if bytes.HasPrefix(data, []byte("\xef\xbb\xbf")) {
			data = data[3:]
		}
		
		content := string(data)
		// 改めてパッケージ名を上書き（先頭の package main を ui に）
		content = strings.Replace(content, "package main\r\n", "package ui\r\n", 1)
		content = strings.Replace(content, "package main\n", "package ui\n", 1)
		
		os.WriteFile(f, []byte(content), 0644)
	}
}
