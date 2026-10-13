package web

import (
	"io"
	"net/http"
	"testing"
)

func TestSoundSelectionAndAudio(t *testing.T) {
	server, _ := newTestServer(t)
	var result any
	body := `{"enabled":true,"only_blur":true,"volume":25,"tone":"chime"}`
	if status := call(t, server, "POST", "/api/settings/sound", body, &result); status != 200 {
		t.Fatalf("save = %d", status)
	}
	var state struct{ Config configView }
	call(t, server, "GET", "/api/state", "", &state)
	if state.Config.Sound.Tone != "chime" || state.Config.Sound.Volume != 25 || !state.Config.Sound.OnlyBlur {
		t.Fatalf("sound = %+v", state.Config.Sound)
	}
	var list []struct{ ID, Name string }
	if status := call(t, server, "GET", "/api/sounds", "", &list); status != 200 || len(list) != 16 {
		t.Fatalf("catalog = %d: %+v", status, list)
	}
	response, err := server.Client().Get(server.URL + "/api/sounds/chime.wav")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	missing, err := server.Client().Get(server.URL + "/api/sounds/unknown.wav")
	if err != nil {
		t.Fatal(err)
	}
	missing.Body.Close()
	if missing.StatusCode != http.StatusNotFound {
		t.Fatalf("unknown sound = %d", missing.StatusCode)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil || response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "audio/wav" || len(data) < 44 || string(data[:4]) != "RIFF" {
		t.Fatalf("audio = %d, %d bytes, err = %v", response.StatusCode, len(data), err)
	}
}
