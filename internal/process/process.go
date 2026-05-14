package process

import (
	"Evidence-Capture/internal/config"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"Evidence-Capture/internal/winapi"
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sys/windows"
)

// IsProcessRunning は指定した実行ファイル名（自分自身を除く）が実行中か判定します。
func IsProcessRunning(exeName string) (bool, error) {
	myPid := int32(os.Getpid())
	processes, err := process.Processes()
	if err != nil {
		return false, err
	}

	for _, p := range processes {
		name, _ := p.Name()
		if strings.EqualFold(name, exeName) && p.Pid != myPid {
			return true, nil
		}
	}
	return false, nil
}

// KillProcessByName は指定した実行ファイル名（自分自身を除く）をすべて終了させます。
func KillProcessByName(exeName string) error {
	myPid := int32(os.Getpid())
	processes, err := process.Processes()
	if err != nil {
		config.Log("ERROR", "プロセス一覧の取得に失敗しました: %v", "Failed to get process list: %v", err)
		return err
	}

	found := false
	for _, p := range processes {
		name, _ := p.Name()
		if strings.EqualFold(name, exeName) && p.Pid != myPid {
			found = true
			if err := p.Kill(); err != nil {
				config.Log("ERROR", "プロセス %s (PID:%d) の強制終了に失敗しました: %v", "Failed to forcibly terminate process %s (PID:%d): %v", name, p.Pid, err)
				return err
			} else {
				config.Log("INFO", "プロセス %s (PID:%d) を正常に強制終了しました。", "Successfully forcibly terminated process %s (PID:%d)", name, p.Pid)
			}
		}
	}

	if !found {
		config.Log("INFO", "終了対象のプロセス %s は見つかりませんでした。", "No process named %s found to kill", exeName)
	}

	return nil
}

// CleanupTemporaryFolders は OS の一時フォルダ内に残っている本アプリの残骸を削除します。
func CleanupTemporaryFolders() {
	tempBase := config.AppTempDir

	// 1. ec_images_* フォルダの削除処理
	files, err := os.ReadDir(tempBase)
	if err == nil {
		for _, f := range files {
			// ディレクトリであり、かつ名称が "ec_images_" で始まるものを対象とする
			if f.IsDir() && strings.HasPrefix(f.Name(), "ec_images_") {
				targetPath := filepath.Join(tempBase, f.Name())
				err := os.RemoveAll(targetPath)
				if err != nil {
					config.Log("WARN", "一時フォルダの削除に失敗しました: %s", "Failed to cleanup temp dir: %s", targetPath)
				} else {
					config.Log("INFO", "古い一時フォルダを清掃しました: %s", "Temporary folder cleaned: %s", targetPath)
				}
			}
		}
	}

	// 2. excel_insert_*.png ファイルの削除処理（filepath.Globを使用）
	pattern := filepath.Join(tempBase, "excel_insert_*.png")
	pngFiles, err := filepath.Glob(pattern)
	if err == nil {
		for _, targetFile := range pngFiles {
			err := os.Remove(targetFile)
			if err != nil {
				// OS側でロックされているなど、削除に失敗しても警告のみで続行
				config.Log("WARN", "一時ファイルの削除に失敗しました: %s", "Failed to cleanup temp file: %s", targetFile)
			} else {
				config.Log("INFO", "古い一時ファイルを清掃しました: %s", "Temporary file cleaned: %s", targetFile)
			}
		}
	}
}

func ForceKillProcessByName(targetExe string) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(snapshot)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))

	err = windows.Process32First(snapshot, &pe)
	if err != nil {
		return err
	}

	for {
		name := windows.UTF16ToString(pe.ExeFile[:])

		if strings.EqualFold(name, targetExe) {
			hProcess, err := windows.OpenProcess(
				windows.PROCESS_TERMINATE,
				false,
				pe.ProcessID,
			)

			if err == nil {
				windows.TerminateProcess(hProcess, 1)
				windows.CloseHandle(hProcess)
			}
		}

		err = windows.Process32Next(snapshot, &pe)
		if err != nil {
			break
		}
	}

	return nil
}
// RestartSelf は現在のプロセスを終了し、新しくプロセスを立ち上げます。
func RestartSelf() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	err = winapi.RestartSelf(exe)
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}
