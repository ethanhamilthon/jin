package store

import "strconv"

const (
	SoundOn   = "on"
	SoundOff  = "off"
	SoundBlur = "blur"

	defaultVolume = 70
	keySoundMode  = "sound.mode"
	keyVolume     = "sound.volume"
	keyOldMute    = "sound.mute"
)

// Sound is the notification played on a final answer. Blur means only while
// the terminal window is not focused.
type Sound struct {
	Mode   string
	Volume int
}

func parseSound(values map[string]string) Sound {
	sound := Sound{Mode: values[keySoundMode], Volume: defaultVolume}
	switch sound.Mode {
	case SoundOn, SoundOff, SoundBlur:
	default:
		sound.Mode = SoundOn
		if values[keyOldMute] == "1" {
			sound.Mode = SoundOff
		}
	}
	if volume, err := strconv.Atoi(values[keyVolume]); err == nil && volume >= 0 && volume <= 100 {
		sound.Volume = volume
	}
	return sound
}

func (db *DB) SaveSound(sound Sound) error {
	return db.setSettings(map[string]string{keySoundMode: sound.Mode, keyVolume: strconv.Itoa(sound.Volume)})
}
