package ui

import (
	"strconv"

	"jin/internal/store"
)

var soundModes = []option{
	{label: "On", detail: "always", value: store.SoundOn},
	{label: "Off", detail: "silent", value: store.SoundOff},
	{label: "Only in blur", detail: "only while the window is not focused", value: store.SoundBlur},
}

var volumes = []int{100, 75, 50, 25, 10}

func (a *app) saveSound(sound store.Sound) error {
	if err := a.store.SaveSound(sound); err != nil {
		return err
	}
	a.cfg.Sound = sound
	return nil
}

func (a *app) openSoundFlow() {
	a.openList("Sound", soundModes, a.cfg.Sound.Mode, func(mode string) error {
		sound := a.cfg.Sound
		sound.Mode = mode
		return a.saveSound(sound)
	})
}

// openVolumeFlow lists the levels; the chosen one is played as a preview.
func (a *app) openVolumeFlow() {
	options := make([]option, len(volumes))
	for i, volume := range volumes {
		label := strconv.Itoa(volume) + "%"
		options[i] = option{label: label, value: strconv.Itoa(volume)}
	}
	a.openList("Volume", options, strconv.Itoa(a.cfg.Sound.Volume), func(value string) error {
		sound := a.cfg.Sound
		sound.Volume, _ = strconv.Atoi(value)
		if err := a.saveSound(sound); err != nil {
			return err
		}
		a.ring()
		return nil
	})
}
