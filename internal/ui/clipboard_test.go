package ui

import "testing"

func TestImageExt(t *testing.T) {
	cases := map[string]string{"\x89PNG\r\n": ".png", "\xff\xd8\xff\xe0": ".jpg", "GIF89a": ".gif", "RIFF0000WEBPVP8": ".webp", "BM00": ".bmp"}
	for data, want := range cases {
		if got := imageExt([]byte(data)); got != want {
			t.Errorf("imageExt(%q) = %s, want %s", data, got, want)
		}
	}
}

func TestLinuxClipboardNeedsADisplay(t *testing.T) {
	t.Setenv("WAYLAND_DISPLAY", "")
	t.Setenv("DISPLAY", "")
	if _, ok := linuxClipboard(); ok {
		t.Fatal("no display, no clipboard tool")
	}
}
