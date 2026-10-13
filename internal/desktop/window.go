package desktop

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type windowState struct{ Width, Height int }

const (
	defaultWidth, defaultHeight = 1200, 800
	minWidth, minHeight         = 640, 480
)

func windowFile() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Jin", "window.json")
}

func loadWindow(path string) windowState {
	state := windowState{defaultWidth, defaultHeight}
	data, err := os.ReadFile(path)
	if err != nil {
		return state
	}
	var saved windowState
	if json.Unmarshal(data, &saved) == nil && saved.Width >= minWidth && saved.Height >= minHeight {
		return saved
	}
	return state
}

func saveWindow(ctx context.Context, path string) {
	if path == "" {
		return
	}
	width, height := runtime.WindowGetSize(ctx)
	if width < minWidth || height < minHeight {
		return
	}
	data, _ := json.Marshal(windowState{width, height})
	if os.MkdirAll(filepath.Dir(path), 0o700) == nil {
		_ = os.WriteFile(path, data, 0o600)
	}
}
