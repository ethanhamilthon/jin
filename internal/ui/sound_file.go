package ui

import (
	"os"

	"jin/internal/sounds"
)

func soundFile(id string) (string, func(), error) {
	file, err := os.CreateTemp("", "jin-sound-*.wav")
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	if _, err = file.Write(sounds.WAV(id)); err != nil {
		file.Close()
		os.Remove(path)
		return "", func() {}, err
	}
	if err = file.Close(); err != nil {
		os.Remove(path)
		return "", func() {}, err
	}
	return path, func() { _ = os.Remove(path) }, nil
}
