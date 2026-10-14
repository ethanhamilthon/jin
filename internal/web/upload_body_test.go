package web

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func uploadRequest(contentType, body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/images", strings.NewReader(body))
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	return r
}

func TestReadUploadRawBody(t *testing.T) {
	data, err := readUpload(uploadRequest("image/png", "bytes"), 100)
	if err != nil || string(data) != "bytes" {
		t.Fatalf("got %q, %v", data, err)
	}
}

func TestReadUploadJSONBase64(t *testing.T) {
	body := `{"data":"` + base64.StdEncoding.EncodeToString([]byte("bytes")) + `"}`
	data, err := readUpload(uploadRequest("application/json", body), 100)
	if err != nil || string(data) != "bytes" {
		t.Fatalf("got %q, %v", data, err)
	}
}

func TestReadUploadRejectsEmpty(t *testing.T) {
	for name, r := range map[string]*http.Request{
		"raw":  uploadRequest("image/png", ""),
		"json": uploadRequest("application/json", `{"data":""}`),
	} {
		if _, err := readUpload(r, 100); err == nil || err.Error() != "empty upload" {
			t.Errorf("%s: err = %v", name, err)
		}
	}
}

func TestReadUploadRejectsBadJSON(t *testing.T) {
	for _, body := range []string{`{"data":"!!!"}`, `not json`} {
		if _, err := readUpload(uploadRequest("application/json", body), 100); err == nil {
			t.Errorf("%q: want an error", body)
		}
	}
}

func TestReadUploadLimit(t *testing.T) {
	big := strings.Repeat("a", 101)
	if _, err := readUpload(uploadRequest("image/png", big), 100); err == nil {
		t.Error("raw body over the limit passed")
	}
	body := `{"data":"` + base64.StdEncoding.EncodeToString([]byte(big)) + `"}`
	if _, err := readUpload(uploadRequest("application/json", body), 100); err == nil {
		t.Error("decoded body over the limit passed")
	}
}
