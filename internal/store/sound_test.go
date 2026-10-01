package store

import "testing"

func TestParseSound(t *testing.T) {
	cases := []struct {
		values map[string]string
		want   Sound
	}{
		{nil, Sound{SoundOn, 70}},
		{map[string]string{keyOldMute: "1"}, Sound{SoundOff, 70}},
		{map[string]string{keySoundMode: SoundBlur, keyVolume: "30"}, Sound{SoundBlur, 30}},
		{map[string]string{keySoundMode: "junk", keyVolume: "900"}, Sound{SoundOn, 70}},
	}
	for _, c := range cases {
		if got := parseSound(c.values); got != c.want {
			t.Errorf("parseSound(%v) = %+v, want %+v", c.values, got, c.want)
		}
	}
}
