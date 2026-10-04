package voice

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFalEndpoint(t *testing.T) {
	cases := []struct{ base, model, want string }{
		{"https://queue.fal.run/fal-ai/wizper", "wizper", "https://fal.run/fal-ai/wizper"},
		{"https://fal.run/fal-ai/whisper/", "x", "https://fal.run/fal-ai/whisper"},
		{"https://fal.run", "wizper", "https://fal.run/fal-ai/wizper"},
		{"https://fal.run", "fal-ai/whisper", "https://fal.run/fal-ai/whisper"},
	}
	for _, c := range cases {
		if got := falEndpoint(c.base, c.model); got != c.want {
			t.Errorf("falEndpoint(%q, %q) = %q, want %q", c.base, c.model, got, c.want)
		}
	}
	if !isFal("https://queue.fal.run/x") || isFal("https://api.groq.com/openai/v1") || isFal("https://notfal.run") {
		t.Error("isFal mismatch")
	}
}

func TestFalRequestBody(t *testing.T) {
	c := Client{BaseURL: "https://fal.run/fal-ai/wizper", APIKey: "k", Model: "wizper", Language: "ru"}
	req, err := c.falRequest(context.Background(), tone(0.5, 3000))
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(req.Body)
	if req.Header.Get("Authorization") != "Key k" || req.URL.String() != "https://fal.run/fal-ai/wizper" {
		t.Errorf("auth=%q url=%s", req.Header.Get("Authorization"), req.URL)
	}
	if !strings.Contains(string(data), `"audio_url":"data:audio/x-wav;base64,UklGR`) || !strings.Contains(string(data), `"language":"ru"`) {
		t.Errorf("body = %.120s", data)
	}
}

func TestErrorTextReadsFalDetail(t *testing.T) {
	if got := errorText([]byte(`{"detail":"Authentication is required"}`)); got != "Authentication is required" {
		t.Errorf("got %q", got)
	}
}

func TestCheckSendsSilence(t *testing.T) {
	var size int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseMultipartForm(1 << 20)
		f, _, _ := r.FormFile("file")
		data, _ := io.ReadAll(f)
		size = len(data)
		_, _ = w.Write([]byte(`{"text":""}`))
	}))
	defer srv.Close()
	if err := (Client{BaseURL: srv.URL, APIKey: "k", Model: "m"}).Check(context.Background()); err != nil || size != 44+bytesPerSecond {
		t.Fatalf("err = %v, size = %d", err, size)
	}
}

func TestLoudness(t *testing.T) {
	if loudness(0, 0) != 0 || loudness(0, 10) != 0 {
		t.Error("silence should be 0")
	}
	if got := loudness(8000*8000*100, 100); got != 1 {
		t.Errorf("loud = %v", got)
	}
	if quiet, loud := loudness(500*500*100, 100), loudness(4000*4000*100, 100); !(quiet > 0 && quiet < loud && loud < 1) {
		t.Errorf("quiet=%v loud=%v", quiet, loud)
	}
}

func TestCheckRejectsQueueAcknowledgement(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"status":"IN_QUEUE","request_id":"x"}`))
	}))
	defer srv.Close()
	if err := (Client{BaseURL: srv.URL, APIKey: "k", Model: "m"}).Check(context.Background()); err == nil {
		t.Fatal("a queue acknowledgement is not a transcription result")
	}
}
