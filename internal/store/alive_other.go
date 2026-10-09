//go:build !unix && !windows

package store

func processAlive(pid int) bool { return pid > 0 }
