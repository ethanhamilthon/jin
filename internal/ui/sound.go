package ui

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"jin/internal/store"
)

const macSound = "/System/Library/Sounds/Glass.aiff"

// shouldRing decides whether a final answer makes a sound right now.
func shouldRing(sound store.Sound, focused bool) bool {
	return sound.Enabled && (!sound.OnlyBlur || !focused)
}

// ring plays the notification. macOS gets a real sound with the chosen
// volume; elsewhere the terminal bell is the only option and has no volume.
func (a *app) ring() {
	if runtime.GOOS == "darwin" && a.cfg.Sound.Volume > 0 {
		if _, err := os.Stat(macSound); err == nil {
			cmd := exec.Command("afplay", "-v", strconv.FormatFloat(float64(a.cfg.Sound.Volume)/100, 'f', 2, 64), macSound)
			if cmd.Start() == nil {
				go cmd.Wait()
				return
			}
		}
	}
	if runtime.GOOS != "darwin" {
		_ = a.screen.Beep()
	}
}
