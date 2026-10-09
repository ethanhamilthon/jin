//go:build windows

package store

import "syscall"

const (
	queryLimited = 0x1000 // PROCESS_QUERY_LIMITED_INFORMATION
	stillActive  = 259
)

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	handle, err := syscall.OpenProcess(queryLimited, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)
	var code uint32
	return syscall.GetExitCodeProcess(handle, &code) == nil && code == stillActive
}
