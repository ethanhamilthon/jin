package ui

import "strconv"

var volumes = []int{10, 25, 50, 75, 100}

// openSoundFlow shows the sound settings as rows whose value changes with
// Left and Right. Every change is saved at once.
func (a *app) openSoundFlow() {
	sound := a.cfg.Sound
	labels := make([]string, len(volumes))
	for i, volume := range volumes {
		labels[i] = strconv.Itoa(volume) + "%"
	}
	options := []option{
		{label: "Toggle", value: "toggle", choices: []string{"On", "Off"}, chosen: indexOf(!sound.Enabled)},
		{label: "When", value: "when", choices: []string{"Always", "On blur"}, chosen: indexOf(sound.OnlyBlur)},
		{label: "Volume", value: "volume", choices: labels, chosen: nearestVolume(sound.Volume)},
	}
	sel := a.openList("Sound", options, "", func(row string) error {
		if row == "volume" {
			a.ring()
		}
		return nil
	})
	sel.hint = "←/→ change · Enter hear · / search"
	sel.keepOpen = true
	sel.onChoice = a.changeSound
}

func (a *app) changeSound(row string, chosen int) error {
	sound := a.cfg.Sound
	switch row {
	case "toggle":
		sound.Enabled = chosen == 0
	case "when":
		sound.OnlyBlur = chosen == 1
	case "volume":
		sound.Volume = volumes[chosen]
	}
	if err := a.store.SaveSound(sound); err != nil {
		return err
	}
	a.cfg.Sound = sound
	if row == "volume" {
		a.ring()
	}
	return nil
}

func indexOf(second bool) int {
	if second {
		return 1
	}
	return 0
}

func nearestVolume(volume int) int {
	best := 0
	for i, level := range volumes {
		if abs(level-volume) < abs(volumes[best]-volume) {
			best = i
		}
	}
	return best
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
