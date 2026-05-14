package process

import (
	"syscall"
	"unsafe"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutex = kernel32.NewProc("CreateMutexW")
	procCloseHandle = kernel32.NewProc("CloseHandle")
)

const ERROR_ALREADY_EXISTS = 183

// CreateSingletonMutex は Windows Mutex を作成し、既に存在するかどうかを判定します。
// 戻り値: (ハンドル, 既に存在したか, エラー)
func CreateSingletonMutex(name string) (uintptr, bool, error) {
	ptr, err := syscall.UTF16PtrFromString(name)
	if err != nil {
		return 0, false, err
	}

	handle, _, _ := procCreateMutex.Call(0, 0, uintptr(unsafe.Pointer(ptr)))

	if syscall.GetLastError() == syscall.Errno(ERROR_ALREADY_EXISTS) {
		return handle, true, nil
	}

	return handle, false, nil
}

// EnsureSingleInstance は二重起動をチェックし、すでに起動している場合は
func EnsureSingleInstance(name string) syscall.Handle {
	namePtr, _ := syscall.UTF16PtrFromString(name)

	handle, _, _ := procCreateMutex.Call(
		0,
		0,
		uintptr(unsafe.Pointer(namePtr)),
	)

	if handle == 0 {
		panic("Mutex作成失敗")
	}

	errCode := syscall.GetLastError()
	if errCode == syscall.ERROR_ALREADY_EXISTS {
		panic("既に起動しています")
	}

	return syscall.Handle(handle)
}

func ReleaseMutex(handle syscall.Handle) {
	if handle != 0 {
		procCloseHandle.Call(uintptr(handle))
	}
}
