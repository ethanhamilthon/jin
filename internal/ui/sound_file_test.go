package ui

import (
	"os"
	"testing"
)

func TestSoundFileCleanup(t *testing.T) {
	path, cleanup, err := soundFile("chime")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	data, err := os.ReadFile(path)
	if err != nil || len(data) <= 44 || string(data[:4]) != "RIFF" {
		t.Fatalf("sound file: %d bytes, %v", len(data), err)
	}
	cleanup()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("file remains after cleanup: %v", err)
	}
}
