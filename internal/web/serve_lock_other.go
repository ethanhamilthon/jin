//go:build !unix

package web

import "context"

type serveLock struct{}

func lockServeEntry(context.Context) (*serveLock, error) { return &serveLock{}, nil }

func (l *serveLock) release() {}
