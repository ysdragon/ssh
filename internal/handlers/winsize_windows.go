//go:build windows
// +build windows

package handlers

import (
	"os"
	"syscall"
	"unsafe"
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")
var procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
var procSetConsoleScreenBufferSize = kernel32.NewProc("SetConsoleScreenBufferSize")
var procSetConsoleWindowInfo = kernel32.NewProc("SetConsoleWindowInfo")

// Console screen buffer info structure
type consoleScreenBufferInfo struct {
	dwSize              coord
	dwCursorPosition    coord
	wAttributes         uint16
	srWindow            smallRect
	dwMaximumWindowSize coord
}

type coord struct {
	x, y int16
}

type smallRect struct {
	left, top, right, bottom int16
}

func setWinsize(f *os.File, w, h int) {
	// Get the console handle from the file descriptor
	handle := syscall.Handle(f.Fd())

	// Get current console info
	var csbi consoleScreenBufferInfo
	ret, _, _ := procGetConsoleScreenBufferInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&csbi)),
	)
	if ret == 0 {
		return // error getting console info, just return
	}

	// Set the new buffer size
	newSize := coord{
		x: int16(w),
		y: int16(h),
	}
	ret, _, _ = procSetConsoleScreenBufferSize.Call(
		uintptr(handle),
		uintptr(*(*int32)(unsafe.Pointer(&newSize))),
	)
	if ret == 0 {
		return // error setting buffer size, just return
	}

	// Set the window size
	windowRect := smallRect{
		left:   0,
		top:    0,
		right:  int16(w - 1),
		bottom: int16(h - 1),
	}
	ret, _, _ = procSetConsoleWindowInfo.Call(
		uintptr(handle),
		1, // absolute coordinates
		uintptr(unsafe.Pointer(&windowRect)),
	)
	if ret == 0 {
		// If setting window info fails, try to at least set buffer size
		// This ensures the buffer is at least the requested size
	}
}
