package ui

const (
	soundOn  = "On"
	soundOff = "Off"
)

// openSoundFlow switches the bell that rings on a final answer.
func (a *app) openSoundFlow() {
	current := soundOn
	if a.cfg.Mute {
		current = soundOff
	}
	options := []option{
		{label: soundOn, detail: "bell on a final answer", value: soundOn},
		{label: soundOff, detail: "silent", value: soundOff},
	}
	a.openList("Sound", options, current, func(value string) error {
		mute := value == soundOff
		if err := a.store.SaveMute(mute); err != nil {
			return err
		}
		a.cfg.Mute = mute
		return nil
	})
}
