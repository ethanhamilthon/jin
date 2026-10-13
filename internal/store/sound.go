package store

import (
	"jin/internal/sounds"
	"strconv"
)

const (
	WhenAlways = "always"
	WhenBlur   = "blur"

	DefaultVolume = 75
	keyEnabled    = "sound.enabled"
	keyWhen       = "sound.when"
	keyVolume     = "sound.volume"
	keyTone       = "sound.tone"
	keyOldMode    = "sound.mode"
	keyOldMute    = "sound.mute"
)

// Sound controls notifications for questions, answers and errors.
type Sound struct {
	Enabled  bool
	OnlyBlur bool
	Volume   int
	Tone     string
}

// parseSound reads the current keys and falls back to the single "mode" and
// "mute" settings that earlier versions saved.
func parseSound(values map[string]string) Sound {
	sound := Sound{Enabled: true, Volume: DefaultVolume, Tone: sounds.Normalize(values[keyTone])}
	switch values[keyOldMode] {
	case "off":
		sound.Enabled = false
	case "blur":
		sound.OnlyBlur = true
	}
	if values[keyOldMute] == "1" {
		sound.Enabled = false
	}
	if value, ok := values[keyEnabled]; ok {
		sound.Enabled = value == "1"
	}
	if value, ok := values[keyWhen]; ok {
		sound.OnlyBlur = value == WhenBlur
	}
	if volume, err := strconv.Atoi(values[keyVolume]); err == nil && volume >= 0 && volume <= 100 {
		sound.Volume = volume
	}
	return sound
}

func (db *DB) SaveSound(sound Sound) error {
	enabled, when := "0", WhenAlways
	if sound.Enabled {
		enabled = "1"
	}
	if sound.OnlyBlur {
		when = WhenBlur
	}
	return db.setSettings(map[string]string{keyEnabled: enabled, keyWhen: when, keyVolume: strconv.Itoa(sound.Volume), keyTone: sounds.Normalize(sound.Tone)})
}
