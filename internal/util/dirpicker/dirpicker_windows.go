package dirpicker

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func pickDirectoryPlatform(ctx context.Context) (string, error) {
	if ctx != nil {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	inited := false
	if err := windows.CoInitializeEx(0, windows.COINIT_APARTMENTTHREADED); err == nil {
		inited = true
	} else if errno, ok := err.(syscall.Errno); ok && errno == syscall.Errno(windows.S_FALSE) {
		inited = true
	}
	if inited {
		defer windows.CoUninitialize()
	}

	title, err := windows.UTF16PtrFromString("Select Directory")
	if err != nil {
		return "", fmt.Errorf("encode dialog title: %w", err)
	}
	displayName := make([]uint16, windows.MAX_PATH)
	bi := browseInfo{
		hwndOwner:      foregroundWindow(),
		pszDisplayName: &displayName[0],
		lpszTitle:      title,
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
		lpfn:           browseCallback,
	}

	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", ErrDirPickerCanceled
	}
	defer procCoTaskMemFree.Call(pidl)

	pathBuf := make([]uint16, windows.MAX_PATH)
	ret, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&pathBuf[0])))
	if ret == 0 {
		return "", errors.New("resolve selected directory failed")
	}
	path := windows.UTF16ToString(pathBuf)
	if path == "" {
		return "", ErrDirPickerCanceled
	}
	return path, nil
}

const (
	bifReturnOnlyFSDirs = 0x0001
	bifNewDialogStyle   = 0x0040

	bffmInitialized = 1

	swpNoSize     = 0x0001
	swpNoMove     = 0x0002
	swpShowWindow = 0x0040
)

var hwndTopmost = ^uintptr(0)

type browseInfo struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

var (
	modshell32               = windows.NewLazySystemDLL("shell32.dll")
	modole32                 = windows.NewLazySystemDLL("ole32.dll")
	moduser32                = windows.NewLazySystemDLL("user32.dll")
	procSHBrowseForFolderW   = modshell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = modshell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree        = modole32.NewProc("CoTaskMemFree")
	procGetForegroundWindow  = moduser32.NewProc("GetForegroundWindow")
	procSetForegroundWindow  = moduser32.NewProc("SetForegroundWindow")
	procSetWindowPos         = moduser32.NewProc("SetWindowPos")

	browseCallback = windows.NewCallback(browseForFolderCallback)
)

func foregroundWindow() uintptr {
	hwnd, _, _ := procGetForegroundWindow.Call()
	return hwnd
}

func browseForFolderCallback(hwnd uintptr, msg uint32, lParam uintptr, lpData uintptr) uintptr {
	if msg != bffmInitialized || hwnd == 0 {
		return 0
	}
	procSetWindowPos.Call(
		hwnd,
		hwndTopmost,
		0,
		0,
		0,
		0,
		swpNoMove|swpNoSize|swpShowWindow,
	)
	procSetForegroundWindow.Call(hwnd)
	return 0
}
