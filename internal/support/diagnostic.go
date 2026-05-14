package support

import (
	"Evidence-Capture/internal/config"
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
)

// CreateDiagnosticPackage は診断用情報をZIPにまとめます
func CreateDiagnosticPackage(destPath string) error {
	// 1. ZIPファイルの作成
	zipFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("診断ファイルの作成に失敗しました: %w", err)
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	// 2. 既存のファイルをZIPに追加
	logPath := config.GetCurrentLogPath()
	filesToInclude := []string{
		config.IniFile1,
		config.IniFile2,
		config.IniFile3,
		logPath,
	}

	for _, file := range filesToInclude {
		if file == "" {
			continue
		}
		// ログファイルの場合はコピーを経由する（ファイルロック対策）
		if file == logPath {
			if err := addLockedFileToZip(archive, file); err != nil {
				log.Printf("[Warning] ログファイルの追加に失敗しました: %v", err)
			}
			continue
		}

		if err := addFileToZip(archive, file); err != nil {
			log.Printf("[Warning] ファイルの追加をスキップしました (%s): %v", file, err)
		}
	}

	// 3. システム情報の生成と追加
	envInfo := generateEnvInfo()
	w, err := archive.Create("env_info.txt")
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(envInfo))

	return err
}

// addLockedFileToZip はロックされている可能性のあるファイルを一時コピーしてZIPに追加します
func addLockedFileToZip(archive *zip.Writer, filename string) error {
	// 1. オリジナルファイルが存在するか確認
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}

	// 2. 一時ファイルの作成
	tempFile, err := os.CreateTemp("", "diag_log_*.tmp")
	if err != nil {
		return err
	}
	tempPath := tempFile.Name()
	defer func() {
		tempFile.Close()
		os.Remove(tempPath) // 処理が終わったら削除
	}()

	// 3. 読み取り専用で開いてコピーを試みる
	// Windowsの os.Open は共有読み取りを許容する場合が多いですが、
	// 確実にコピーするために一時ファイルへストリーム出力します。
	src, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("ログファイルを開けません: %w", err)
	}
	defer src.Close()

	if _, err := io.Copy(tempFile, src); err != nil {
		return fmt.Errorf("ログのコピーに失敗しました: %w", err)
	}
	tempFile.Seek(0, 0) // ポインタを先頭に戻す

	// 4. 一時ファイルをZIPに追加
	info, _ := tempFile.Stat()
	header, _ := zip.FileInfoHeader(info)
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, tempFile)
	return err
}

func addFileToZip(archive *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filename)
	header.Method = zip.Deflate

	writer, err := archive.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}

func generateEnvInfo() string {
	info := "=== Evidence-Capture 診断レポート ===\n"

	// OS詳細情報の取得
	if hInfo, err := host.Info(); err == nil {
		info += fmt.Sprintf("OS: %s (%s)\n", hInfo.Platform, hInfo.OS)
		info += fmt.Sprintf("OS Version: %s\n", hInfo.PlatformVersion)
	} else {
		info += fmt.Sprintf("OS: %s\n", runtime.GOOS)
	}

	info += fmt.Sprintf("Arch: %s\n", runtime.GOARCH)
	info += fmt.Sprintf("CPU Cores: %d\n", runtime.NumCPU())

	// Excelの有無およびバージョンチェック (OLE使用)
	excelStatus := "未インストール"
	ole.CoInitialize(0)
	defer ole.CoUninitialize()
	if unknown, err := oleutil.CreateObject("Excel.Application"); err == nil {
		defer unknown.Release()
		excelApp, err := unknown.QueryInterface(ole.IID_IDispatch)
		if err == nil {
			defer excelApp.Release()
			if v, err := excelApp.GetProperty("Version"); err == nil {
				excelStatus = fmt.Sprintf("インストール済み (Version: %v)", v.Value())
			} else {
				excelStatus = "インストール済み"
			}
		} else {
			excelStatus = "インストール済み"
		}
	}
	info += fmt.Sprintf("Excel状態: %s\n", excelStatus)

	// ディスク空き容量 (設定ディレクトリのドライブ)
	if usage, err := disk.Usage(config.ConfigDir); err == nil {
		info += fmt.Sprintf("ディスク空き容量: %.2f GB / %.2f GB\n",
			float64(usage.Free)/1024/1024/1024,
			float64(usage.Total)/1024/1024/1024)
	}

	return info
}
