package provider

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	retryBase = time.Millisecond
	_ = os.Unsetenv("JIN_DEBUG")
	os.Exit(m.Run())
}
