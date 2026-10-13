package sounds

import (
	"encoding/binary"
	"testing"
)

func TestCatalogAndWAV(t *testing.T) {
	list := Catalog()
	if len(list) != 16 {
		t.Fatalf("sounds = %d, want 16", len(list))
	}
	seen := map[string]bool{}
	for _, sound := range list {
		if sound.ID == "" || sound.Name == "" || seen[sound.ID] {
			t.Fatalf("invalid sound: %+v", sound)
		}
		seen[sound.ID] = true
		data := WAV(sound.ID)
		if string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
			t.Fatalf("invalid WAV: %s", sound.ID)
		}
		if int(binary.LittleEndian.Uint32(data[40:44])) != len(data)-44 {
			t.Fatalf("invalid data size: %s", sound.ID)
		}
		if len(data) <= 44 || len(data) > 100000 {
			t.Fatalf("invalid length: %s: %d", sound.ID, len(data))
		}
	}
	if Normalize("unknown") != Default || Normalize("") != Default || Normalize("pluck") != "pluck" {
		t.Fatal("invalid sound fallback")
	}
}
