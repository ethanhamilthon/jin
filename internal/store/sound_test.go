package store

import "testing"

func TestParseSound(t *testing.T) {
	cases := []struct {
		values map[string]string
		want   Sound
	}{
		{nil, Sound{true, false, 75, "bell"}},
		{map[string]string{keyOldMute: "1"}, Sound{false, false, 75, "bell"}},
		{map[string]string{keyOldMode: "blur", keyVolume: "30"}, Sound{true, true, 30, "bell"}},
		{map[string]string{keyOldMode: "off", keyEnabled: "1", keyWhen: WhenBlur}, Sound{true, true, 75, "bell"}},
		{map[string]string{keyEnabled: "0", keyVolume: "900"}, Sound{false, false, 75, "bell"}},
		{map[string]string{keyTone: "pluck"}, Sound{true, false, 75, "pluck"}},
		{map[string]string{keyTone: "unknown"}, Sound{true, false, 75, "bell"}},
	}
	for _, c := range cases {
		if got := parseSound(c.values); got != c.want {
			t.Errorf("parseSound(%v) = %+v, want %+v", c.values, got, c.want)
		}
	}
}
