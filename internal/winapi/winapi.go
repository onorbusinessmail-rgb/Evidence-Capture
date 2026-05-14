// Package winapi は Windows API のバインディング、定数、および基本的な
// Win32 ダイアログ操作を提供します。
// このパッケージは walk ライブラリと標準ライブラリのみに依存します。
package winapi

import (
	"syscall"
	"time"
	"unsafe"

	"github.com/lxn/walk"
	"golang.org/x/sys/windows"
)

// =========================================================================
// Windows API 定義・定数
// =========================================================================
const (
	// DWMWA_EXTENDED_FRAME_BOUNDS は、ウィンドウの影を除いた正確な矩形を取得するための属性
	DWMWA_EXTENDED_FRAME_BOUNDS = 9
)

var (
	Moddwmapi            = syscall.NewLazyDLL("dwmapi.dll")
	ProcDwmGetWindowAttr = Moddwmapi.NewProc("DwmGetWindowAttribute")
)

var (
	shell32           = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32.NewProc("ShellExecuteW")
)

var (
	User32   = syscall.NewLazyDLL("user32.dll")
	Gdi32    = syscall.NewLazyDLL("gdi32.dll")
	Kernel32 = syscall.NewLazyDLL("kernel32.dll")

	// --- user32.dll から取得 ---
	GetCursorPos               = User32.NewProc("GetCursorPos")
	GetAsyncKeyState           = User32.NewProc("GetAsyncKeyState")
	GetSystemMetrics           = User32.NewProc("GetSystemMetrics")
	MessageBox                 = User32.NewProc("MessageBoxW")
	MessageBeep                = User32.NewProc("MessageBeep")
	KeybdEvent                 = User32.NewProc("keybd_event")
	PostMessage                = User32.NewProc("PostMessageW")
	FindWindow                 = User32.NewProc("FindWindowW")
	IsClipboardFormatAvailable = User32.NewProc("IsClipboardFormatAvailable")
	SetForegroundWindow        = User32.NewProc("SetForegroundWindow")
	GetClipboardSequenceNumber = User32.NewProc("GetClipboardSequenceNumber")
	OpenClipboard              = User32.NewProc("OpenClipboard")
	CloseClipboard             = User32.NewProc("CloseClipboard")
	GetClipboardData           = User32.NewProc("GetClipboardData")
	SetWindowPos               = User32.NewProc("SetWindowPos")
	SendMessage                = User32.NewProc("SendMessageW")
	SendMessageTimeout         = User32.NewProc("SendMessageTimeoutW")
	ShowWindow                 = User32.NewProc("ShowWindow")
	MonitorFromPoint           = User32.NewProc("MonitorFromPoint")
	GetMonitorInfo             = User32.NewProc("GetMonitorInfoW")
	CreateDIBSection           = Gdi32.NewProc("CreateDIBSection")

	// --- kernel32.dll から取得 ---
	GetDiskFreeSpaceEx = Kernel32.NewProc("GetDiskFreeSpaceExW")

	// ウィンドウ情報の取得用
	GetForegroundWindow = User32.NewProc("GetForegroundWindow")
	GetWindowRect       = User32.NewProc("GetWindowRect")
	GetDC               = User32.NewProc("GetDC")
	ReleaseDC           = User32.NewProc("ReleaseDC")
	PrintWindow         = User32.NewProc("PrintWindow")
	EmptyClipboard      = User32.NewProc("EmptyClipboard")
	SetClipboardData    = User32.NewProc("SetClipboardData")

	// DPI 関連
	SetProcessDpiAwarenessContext = User32.NewProc("SetProcessDpiAwarenessContext")
	GetDpiForWindow               = User32.NewProc("GetDpiForWindow")

	// gdi32.dll から取得
	// 画面キャプチャ（ビットマップ操作）用
	CreateCompatibleDC     = Gdi32.NewProc("CreateCompatibleDC")
	CreateCompatibleBitmap = Gdi32.NewProc("CreateCompatibleBitmap")
	SelectObject           = Gdi32.NewProc("SelectObject")
	BitBlt                 = Gdi32.NewProc("BitBlt")
	DeleteDC               = Gdi32.NewProc("DeleteDC")
	DeleteObject           = Gdi32.NewProc("DeleteObject")
	GetObject              = Gdi32.NewProc("GetObjectW")
	GetDIBits              = Gdi32.NewProc("GetDIBits")
	AllowSetForegroundWindow = User32.NewProc("AllowSetForegroundWindow")
)

const (
	// --- モニター・システム情報 (GetSystemMetrics用) ---
	SM_CXSCREEN = 0 // プライマリモニターの画面幅（ピクセル）
	SM_CYSCREEN = 1 // プライマリモニターの画面高さ（ピクセル）

	// --- 仮想キーコード (Virtual-Key Codes) ---
	VK_SHIFT    = 0x10 // SHIFT キー
	VK_CONTROL  = 0x11 // CTRL キー
	VK_MENU     = 0x12 // ALT キー
	VK_RETURN   = 0x0D // ENTER キー
	VK_ESCAPE   = 0x1B // ESC キー
	VK_SNAPSHOT = 0x2C // PRINT SCREEN キー
	VK_X        = 0x58 // X キー
	VK_LWIN     = 0x5B // 左Windowsキー
	VK_RWIN     = 0x5C // 右Windowsキー
	VK_LBUTTON  = 0x01 // マウス左ボタン
	VK_RBUTTON  = 0x02 // マウス右ボタン

	// --- キーボードイベントフラグ (keybd_event用) ---
	KEYEVENTF_KEYUP = 0x0002 // キーを離す動作を指定（指定しない場合は「押す」）

	// --- クリップボード形式 (Clipboard Formats) ---
	CF_BITMAP = 2 // ビットマップ形式の画像データ

	// --- Windows メッセージ (Window Messages) ---
	WM_NULL    = 0      // 接続確認用メッセージ
	WM_KEYDOWN = 0x0100 // 非システムキーが押された時のメッセージ
	WM_KEYUP   = 0x0101 // 非システムキーが離された時のメッセージ
	WM_CLOSE   = 0x0010 // ウィンドウを閉じるメッセージ
	WM_USER    = 0x0400 // ユーザー定義メッセージの開始番号

	// --- タイムアウト・フラグ ---
	SMTO_ABORTIFHUNG = 0x0002 // 応答なし時に即時復帰

	// --- ダイアログ：ボタンスタイル (MessageBox用) ---
	MB_OK          = 0x00000000 // [OK] ボタンを表示
	MB_YESNO       = 0x00000004 // [はい] [いいえ] ボタンを表示
	MB_YESNOCANCEL = 0x00000003 // [はい] [いいえ] [キャンセル] ボタンを表示

	// --- ダイアログ：アイコン種類 ---
	MB_ICONERROR       = 0x00000010 // エラーアイコン（×）を表示
	MB_ICONQUESTION    = 0x00000020 // 問い合わせアイコン（？）を表示
	MB_ICONWARNING     = 0x00000030 // 警告アイコン（！）を表示
	MB_ICONINFORMATION = 0x00000040 // 情報アイコン（i）を表示

	// --- ダイアログ：表示オプション ---
	MB_SYSTEMMODAL   = 0x00001000 // 全てのウィンドウの最前面に固定
	MB_SETFOREGROUND = 0x00010000 // ダイアログをフォアグラウンドウィンドウにする
	MB_TOPMOST       = 0x00040000 // ダイアログを常に最前面に表示する

	// --- ダイアログ：戻り値 (ユーザーの選択結果) ---
	IDYES    = 6 // [はい] がクリックされた
	IDNO     = 7 // [いいえ] がクリックされた
	IDCANCEL = 2 // [キャンセル] がクリックされた

	// --- ウィンドウ位置・状態制御 (SetWindowPos / ShowWindow用) ---
	HWND_TOP       = uintptr(0)
	HWND_TOPMOST   = ^uintptr(0) // -1
	HWND_NOTOPMOST = ^uintptr(1) // -2
	SWP_NOSIZE     = 0x0001
	SWP_NOMOVE     = 0x0002
	SWP_SHOWWINDOW = 0x0040
	SW_RESTORE     = 9
	SW_SHOWNORMAL  = 1
	SRCCOPY        = 0x00CC0020

	// --- GDI 定数 ---
	DIB_RGB_COLORS = 0

	// --- OLE/Excel関連 ---
	S_FALSE          = 1
	MsoPicture       = 13
	MsoLinkedPicture = 11

	DPI_AWARENESS_CONTEXT_PER_MONITOR_AWARE_V2 = uintptr(0xfffffffc) // -4
	HGDI_ERROR                                 = ^uintptr(0)          // (HGDIOBJ)-1
)

// POINT はマウス座標を格納するための構造体定義です。
type POINT struct {
	X, Y int32
}

// RECT は矩形領域を表す構造体定義です。
type RECT struct {
	Left, Top, Right, Bottom int32
}

// MONITORINFO はモニターの情報を表す構造体定義です。
type MONITORINFO struct {
	CbSize    uint32
	RcMonitor RECT
	RcWork    RECT
	DwFlags   uint32
}

// =========================================================================
// ダイアログ・通知関数
// =========================================================================

// ShowDialog は Windows 標準のメッセージボックスを表示し、
// ユーザーがクリックしたボタンの ID を返します。
func ShowDialog(title, msg string, icon uintptr) int {
	MessageBeep.Call(uintptr(icon))

	t, _ := syscall.UTF16PtrFromString(title)
	m, _ := syscall.UTF16PtrFromString(msg)

	ret, _, _ := MessageBox.Call(
		0,
		uintptr(unsafe.Pointer(m)),
		uintptr(unsafe.Pointer(t)),
		icon|0x40000,
	)
	return int(ret)
}

// ShowAutoCloseDialog は指定秒数後に自動的に閉じる最前面メッセージボックスを表示します。
func ShowAutoCloseDialog(title, msg string, seconds int) {
	go func() {
		if seconds > 0 {
			time.AfterFunc(time.Duration(seconds)*time.Second, func() {
				t, _ := syscall.UTF16PtrFromString(title)
				hwnd, _, _ := FindWindow.Call(0, uintptr(unsafe.Pointer(t)))
				if hwnd != 0 {
					PostMessage.Call(hwnd, uintptr(WM_CLOSE), 0, 0)
				}
			})
		}
		ShowDialog(title, msg, MB_ICONWARNING|MB_SYSTEMMODAL)
	}()
}

// IsClipboardImageAvailable は現在クリップボードにビットマップ画像が存在するかを確認します。
func IsClipboardImageAvailable() bool {
	ret, _, _ := IsClipboardFormatAvailable.Call(uintptr(CF_BITMAP))
	return ret != 0
}

// PlayBeep は標準의システム通知音を鳴らします。
func PlayBeep() {
	MessageBeep.Call(uintptr(0))
}

// SetAlwaysOnTop は walk.Form（Dialog/MainWindow等）を常に最前面に固定するかどうかを制御します。
func SetAlwaysOnTop(form walk.Form, alwaysOnTop bool) {
	if form == nil || form.Handle() == 0 {
		return
	}

	target := HWND_NOTOPMOST
	if alwaysOnTop {
		target = HWND_TOPMOST
	}

	SetWindowPos.Call(
		uintptr(form.Handle()),
		target,
		0, 0, 0, 0,
		SWP_NOMOVE|SWP_NOSIZE|SWP_SHOWWINDOW,
	)
}

// GetScreenResolution は現在のメインモニタの解像度を動的に取得します
func GetScreenResolution() (int32, int32) {
	w, _, _ := GetSystemMetrics.Call(uintptr(SM_CXSCREEN))
	h, _, _ := GetSystemMetrics.Call(uintptr(SM_CYSCREEN))
	return int32(w), int32(h)
}

// GetPrimaryMonitorRect を定数キャッシュではなく、都度実行するように修正（もしキャッシュしていた場合）
func GetPrimaryMonitorRect() RECT {
	w, h := GetScreenResolution()
	return RECT{Left: 0, Top: 0, Right: w, Bottom: h}
}

const EM_SETCUEBANNER = 0x1501

// SetCueBanner はテキストボックスにプレースホルダー（Cue Banner）を設定します。
func SetCueBanner(hwnd uintptr, text string) {
	ptr, _ := syscall.UTF16PtrFromString(text)
	SendMessage.Call(
		hwnd,
		uintptr(EM_SETCUEBANNER),
		1,
		uintptr(unsafe.Pointer(ptr)),
	)
}

// GetDiskFreeSpaceGB は指定ドライブの空き容量をGB単位で取得します。
func GetDiskFreeSpaceGB(drive string) float64 {
	var freeBytes, totalBytes, totalFreeBytes int64
	ptr, _ := syscall.UTF16PtrFromString(drive)
	ret, _, _ := GetDiskFreeSpaceEx.Call(
		uintptr(unsafe.Pointer(ptr)),
		uintptr(unsafe.Pointer(&freeBytes)),
		uintptr(unsafe.Pointer(&totalBytes)),
		uintptr(unsafe.Pointer(&totalFreeBytes)),
	)
	if ret == 0 {
		return 0
	}
	return float64(freeBytes) / 1024 / 1024 / 1024
}

func ShowFileInExplorer(path string) error {
	ptrArgs, err := syscall.UTF16PtrFromString("/select,\"" + path + "\"")
	if err != nil {
		return err
	}

	ret, _, callErr := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("explorer.exe"))),
		uintptr(unsafe.Pointer(ptrArgs)),
		0,
		1,
	)

	if ret <= 32 {
		return callErr
	}

	return nil
}

func OpenFolder(path string) error {
	ret, _, callErr := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("open"))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(path))),
		0,
		0,
		1,
	)

	if ret <= 32 {
		return callErr
	}

	return nil
}

func RestartSelf(exePath string) error {
	var si windows.StartupInfo
	var pi windows.ProcessInformation

	cmdLine, err := windows.UTF16PtrFromString(`"` + exePath + `"`)
	if err != nil {
		return err
	}

	si.Cb = uint32(unsafe.Sizeof(si))
	si.Flags = windows.STARTF_USESHOWWINDOW
	si.ShowWindow = windows.SW_SHOWNORMAL

	err = windows.CreateProcess(
		nil,
		cmdLine,
		nil,
		nil,
		false,
		0,
		nil,
		nil,
		&si,
		&pi,
	)

	if err != nil {
		return err
	}

	// 新しいプロセスにフォーカス設定の権限を委譲する
	AllowSetForegroundWindow.Call(uintptr(pi.ProcessId))

	windows.CloseHandle(pi.Process)
	windows.CloseHandle(pi.Thread)

	return nil
}
