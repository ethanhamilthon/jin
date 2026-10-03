package store

const (
	keyTheme  = "ui.theme"
	keyMotion = "ui.motion"
)

// Motion is how fast the input glow and the logo shimmer move.
const (
	MotionOff    = "off"
	MotionSlow   = "slow"
	MotionNormal = "normal"
	MotionFast   = "fast"
)

func parseMotion(value string) string {
	switch value {
	case MotionOff, MotionSlow, MotionFast:
		return value
	}
	return MotionNormal
}

// SaveTheme stores the name of the color theme.
func (db *DB) SaveTheme(name string) error {
	return db.setSettings(map[string]string{keyTheme: name})
}

// SaveMotion stores the animation speed, or MotionOff.
func (db *DB) SaveMotion(motion string) error {
	return db.setSettings(map[string]string{keyMotion: parseMotion(motion)})
}
