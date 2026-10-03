package provider

import (
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	retryBase = time.Millisecond
	os.Exit(m.Run())
}
