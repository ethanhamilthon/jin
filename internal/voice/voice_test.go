package voice

import (
	"context"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func tone(seconds float64, amp int16) []byte {
	pcm := make([]byte, int(seconds*bytesPerSecond))
	for i := 0; i+1 < len(pcm); i += 2 {
		v := amp
		if (i/2)%2 == 0 {
			v = -amp
		}
		binary.LittleEndian.PutUint16(pcm[i:], uint16(v))
	}
	return pcm
}

func TestWAVHeader(t *testing.T) {
	pcm := tone(0.1, 1000)
	wav := WAV(pcm)
	if len(wav) != 44+len(pcm) || string(wav[:4]) != "RIFF" || string(wav[8:16]) != "WAVEfmt " || string(wav[36:40]) != "data" {
		t.Fatalf("bad header: %q", wav[:44])
	}
	if got := binary.LittleEndian.Uint32(wav[24:]); got != SampleRate {
		t.Errorf("sample rate = %d", got)
	}
	if got := binary.LittleEndian.Uint32(wav[40:]); int(got) != len(pcm) {
		t.Errorf("data size = %d", got)
	}
}

func TestSpeechless(t *testing.T) {
	cases := []struct {
		name string
		pcm  []byte
		want bool
	}{
		{"empty", nil, true},
		{"too short", tone(0.1, 5000), true},
		{"silent", tone(1, 5), true},
		{"loud", tone(1, 3000), false},
	}
	for _, c := range cases {
		if got := Speechless(c.pcm); got != c.want {
			t.Errorf("%s: Speechless = %v, want %v", c.name, got, c.want)
		}
	}
}

func TestTranscribeSendsOpenAIRequest(t *testing.T) {
	var path, auth, model, lang string
	var size int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, auth = r.URL.Path, r.Header.Get("Authorization")
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Error(err)
		}
		model, lang = r.FormValue("model"), r.FormValue("language")
		f, _, err := r.FormFile("file")
		if err != nil {
			t.Error(err)
			return
		}
		data, _ := io.ReadAll(f)
		size = len(data)
		_, _ = w.Write([]byte(`{"text":" hello world "}`))
	}))
	defer srv.Close()
	c := Client{BaseURL: srv.URL + "/v1/", APIKey: "k", Model: "m", Language: "ru"}
	pcm := tone(0.5, 3000)
	text, err := c.Transcribe(context.Background(), pcm)
	if err != nil || text != "hello world" {
		t.Fatalf("text = %q, err = %v", text, err)
	}
	if path != "/v1/audio/transcriptions" || auth != "Bearer k" || model != "m" || lang != "ru" || size != 44+len(pcm) {
		t.Errorf("path=%q auth=%q model=%q lang=%q size=%d", path, auth, model, lang, size)
	}
}

func TestTranscribeReportsAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer srv.Close()
	_, err := Client{BaseURL: srv.URL, APIKey: "k", Model: "m"}.Transcribe(context.Background(), tone(0.5, 3000))
	if err == nil || !strings.Contains(err.Error(), "401") || !strings.Contains(err.Error(), "bad key") {
		t.Fatalf("err = %v", err)
	}
}
