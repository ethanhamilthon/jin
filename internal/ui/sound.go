package ui

import (
	"os/exec"
	"runtime"
	"strconv"

	"jin/internal/store"
)

// shouldRing checks mute, volume and focus before a notification.
func shouldRing(sound store.Sound, focused bool) bool {
	return sound.Enabled && sound.Volume > 0 && (!sound.OnlyBlur || !focused)
}

// ring plays the notification. macOS gets a real sound with the chosen
// volume; elsewhere the terminal bell is the only option and has no volume.
func (a *app) ring() {
	if a.cfg.Sound.Volume <= 0 {
		return
	}
	if runtime.GOOS == "darwin" {
		path, cleanup, err := soundFile(a.cfg.Sound.Tone)
		if err != nil {
			return
		}
		cmd := exec.Command("afplay", "-v", strconv.FormatFloat(float64(a.cfg.Sound.Volume)/100, 'f', 2, 64), path)
		if cmd.Start() == nil {
			go func() {
				defer cleanup()
				_ = cmd.Wait()
			}()
		} else {
			cleanup()
		}
		return
	}
	_ = a.screen.Beep()
}
