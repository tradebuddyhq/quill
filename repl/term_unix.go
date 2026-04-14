//go:build !windows

package repl

import (
	"os"
	"syscall"
	"unsafe"
)

var origTermios syscall.Termios
var termSet bool

func enableRawMode() bool {
	fd := int(os.Stdin.Fd())
	var t syscall.Termios
	if !getTermios(fd, &t) {
		return false
	}
	origTermios = t
	t.Lflag &^= syscall.ECHO | syscall.ICANON
	t.Cc[syscall.VMIN] = 1
	t.Cc[syscall.VTIME] = 0
	if setTermios(fd, &t) {
		termSet = true
		return true
	}
	return false
}

func disableRawMode() {
	if termSet {
		fd := int(os.Stdin.Fd())
		setTermios(fd, &origTermios)
		termSet = false
	}
}

func getTermios(fd int, t *syscall.Termios) bool {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlReadTermios, uintptr(unsafe.Pointer(t)))
	return errno == 0
}

func setTermios(fd int, t *syscall.Termios) bool {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), ioctlWriteTermios, uintptr(unsafe.Pointer(t)))
	return errno == 0
}
