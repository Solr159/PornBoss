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
		pszDisplayName: &displayName[0],
		lpszTitle:      title,
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
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
)

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
	procSHBrowseForFolderW   = modshell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = modshell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree        = modole32.NewProc("CoTaskMemFree")
)
