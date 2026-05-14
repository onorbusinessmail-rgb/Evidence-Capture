package config

import (
	"Evidence-Capture/internal/types"
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var currentLogPath string

// GetCurrentLogPath は現在書き込み中のログファイルの絶対パスを返します。
func GetCurrentLogPath() string {
	return currentLogPath
}

// dailyLogger は日付が変わると自動でログファイルを切り替える io.WriteCloser 実装です。
type dailyLogger struct {
	mu       sync.Mutex
	logDir   string
	currYear int // 現在開いているファイルの年
	currYDay int // 現在開いているファイルの通算日
	file     *os.File
}

// Write はログを書き込みます。日付が変わっていればファイルをローテーションします。
func (l *dailyLogger) Write(p []byte) (n int, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	year := now.Year()
	yday := now.YearDay()

	if year != l.currYear || yday != l.currYDay {
		// 日付が変わったため、現在のファイルを閉じて新しいファイルを開く
		if l.file != nil {
			l.file.Close()
		}

		today := now.Format("20060102")
		logFileName := today + ".log"
		logFilePath, _ := filepath.Abs(filepath.Join(l.logDir, logFileName))
		currentLogPath = logFilePath

		f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}
		l.file = f
		l.currYear = year
		l.currYDay = yday
	}

	return l.file.Write(p)
}

// Close は現在開いているログファイルを閉じます。
func (l *dailyLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// InitLogger は設定に基づいてロガーの出力先を初期化します。
// cfg.EnableLogOutput が false の場合は破棄、true の場合は日次ローテーション対応のロガーを返します。
func InitLogger(cfg types.AppConfig) (io.WriteCloser, error) {
	if !cfg.EnableLogOutput {
		log.SetOutput(io.Discard)
		currentLogPath = ""
		return nil, nil
	}

	// 出力先フォルダの準備
	logDir := cfg.LogOutputPath
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	now := time.Now()
	currentMonth := now.Format("200601")

	// 1. 起動時の月次アーカイブ処理
	entries, _ := os.ReadDir(logDir)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".log" && len(name) >= 6 {
			fileMonth := name[:6]
			isNumeric := true
			for i := 0; i < 6; i++ {
				if name[i] < '0' || name[i] > '9' {
					isNumeric = false
					break
				}
			}

			if isNumeric && fileMonth < currentMonth {
				archiveDir := filepath.Join(logDir, fileMonth)
				if err := os.MkdirAll(archiveDir, 0755); err == nil {
					oldPath := filepath.Join(logDir, name)
					newPath := filepath.Join(archiveDir, name)
					os.Rename(oldPath, newPath)
				}
			}
		}
	}

	// 2. 起動時の前年以前の自動ZIP圧縮処理
	archiveYearlyLogs(logDir, now.Year())

	// 3. 日次ローテーションロガーの初期化
	l := &dailyLogger{
		logDir:   logDir,
		currYear: now.Year(),
		currYDay: now.YearDay(),
	}

	// 初回ファイルオープン
	today := now.Format("20060102")
	logFileName := today + ".log"
	logFilePath, _ := filepath.Abs(filepath.Join(logDir, logFileName))
	currentLogPath = logFilePath

	f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	l.file = f

	log.SetOutput(l)
	return l, nil
}

// archiveYearlyLogs は前年以前の月別ログフォルダを年ごとにZIP圧縮して整理します。
func archiveYearlyLogs(logDir string, currentYear int) {
	entries, _ := os.ReadDir(logDir)
	yearToDirs := make(map[string][]string)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// フォルダ名が数値6桁（YYYYMM）かチェック
		if len(name) == 6 {
			yearStr := name[:4]
			if year, err := strconv.Atoi(yearStr); err == nil && year < currentYear {
				yearToDirs[yearStr] = append(yearToDirs[yearStr], name)
			}
		}
	}

	for year, dirs := range yearToDirs {
		zipFileName := fmt.Sprintf("logs_%s.zip", year)
		zipPath := filepath.Join(logDir, zipFileName)

		// 既にZIPが存在する場合は、追加書き込みではなくスキップ（または必要に応じて上書き）
		// ここでは単純化のため新規作成を試みる
		if err := createZipFromDirs(zipPath, logDir, dirs); err == nil {
			// 圧縮が成功した場合のみ、元の月別フォルダを削除
			for _, d := range dirs {
				os.RemoveAll(filepath.Join(logDir, d))
			}
		}
	}
}

// createZipFromDirs は指定されたディレクトリ群を一つのZIPファイルにまとめます。
func createZipFromDirs(zipPath string, baseDir string, targets []string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	for _, target := range targets {
		targetPath := filepath.Join(baseDir, target)
		err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			header, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}

			// ZIP内の相対パスを設定 (YYYYMM/YYYYMMDD.log 形式を維持)
			relPath, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			header.Name = filepath.ToSlash(relPath)
			header.Method = zip.Deflate

			writer, err := archive.CreateHeader(header)
			if err != nil {
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		})
		if err != nil {
			archive.Close() // ループ中断時も閉じる
			return err
		}
	}

	// 書き込みを完了し、エラーをチェックする
	if err := archive.Close(); err != nil {
		return err
	}

	return nil
}

// GetDirSize は指定したディレクトリ内の全ファイルの合計サイズ（バイト）を返します。
func GetDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// CleanupOldLogArchives はログフォルダ内の古いZIPアーカイブをすべて削除します。
func CleanupOldLogArchives(logDir string) error {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// logs_YYYY.zip の形式を削除対象とする
		if filepath.Ext(name) == ".zip" && strings.HasPrefix(name, "logs_") {
			path := filepath.Join(logDir, name)
			if err := os.Remove(path); err != nil {
				Log("ERROR", "古いログアーカイブの削除に失敗しました: %s", "Failed to remove old log archive: %s", name)
			}
		}
	}
	return nil
}

// HandleLogCapacity はログフォルダの容量をチェックし、超過している場合にプロンプトを表示してクリーンアップを実行します。
// prompter はユーザーに削除の可否を確認する関数（UI実装）を渡します。
func HandleLogCapacity(cfg types.AppConfig, prompter func(currentMB, maxMB int) bool) {
	if !cfg.EnableLogOutput || cfg.LogOutputPath == "" {
		return
	}

	size, err := GetDirSize(cfg.LogOutputPath)
	if err != nil {
		return
	}

	sizeMB := int(size / 1024 / 1024)
	if sizeMB <= cfg.LogMaxSizeMB {
		return
	}

	// 容量オーバー時、プロンプトを介してクリーンアップを実行
	if prompter != nil && prompter(sizeMB, cfg.LogMaxSizeMB) {
		if err := CleanupOldLogArchives(cfg.LogOutputPath); err == nil {
			Log("INFO", "容量超過のため古いログアーカイブをクリーンアップしました。", "Old logs cleaned up due to capacity")
		}
	}
}

// Log は現在の言語設定に基づいてメッセージを選択・出力します。
// 設定が未ロードの場合は日英を併記します。
func Log(level string, msgJP string, msgEN string, v ...interface{}) {
	if !CurrentConfig.EnableLogOutput && CurrentConfig.Language != "" {
		return
	}

	lang := CurrentConfig.Language

	// 設定未ロード（空文字）の場合は日英両方を出力
	if lang == "" {
		fJP := msgJP
		fEN := msgEN
		if len(v) > 0 {
			fJP = fmt.Sprintf(msgJP, v...)
			fEN = fmt.Sprintf(msgEN, v...)
		}
		log.Printf("[%s] %s / %s", level, fJP, fEN)
		return
	}

	// 言語に応じたメッセージ選択
	baseMsg := msgJP
	if lang == "English" {
		baseMsg = msgEN
	}

	// フォーマット処理
	formattedMsg := baseMsg
	if len(v) > 0 {
		// 注意: メッセージ側に %v 等が含まれていない場合、fmt.Sprintf は
		// 余った引数を %(EXTRA ...) として末尾に付加します。
		// これを避けるために呼び出し側でディレクティブを合わせる必要があります。
		formattedMsg = fmt.Sprintf(baseMsg, v...)
	}

	log.Printf("[%s] %s", level, formattedMsg)
}
