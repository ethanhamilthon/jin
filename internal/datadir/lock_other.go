//go:build !unix

package datadir

// Hold, Release, Exclusive and CheckAlone need flock; elsewhere they do nothing.
func Hold() error { return nil }

func Release() {}

func Exclusive() error { return nil }

func CheckAlone() error { return nil }
