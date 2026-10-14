package web

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
)

var errEmptyUpload = errors.New("empty upload")

// readUpload returns the bytes of an uploaded file. The page sends them as
// JSON {"data": base64}, because the desktop app can lose a raw binary body;
// a raw body is still accepted. limit counts the decoded bytes.
func readUpload(r *http.Request, limit int64) ([]byte, error) {
	kind, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	var data []byte
	if kind == "application/json" {
		var body struct {
			Data string `json:"data"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, limit*4/3+1024)).Decode(&body); err != nil {
			return nil, errors.New("bad upload body: " + err.Error())
		}
		decoded, err := base64.StdEncoding.DecodeString(body.Data)
		if err != nil {
			return nil, errors.New("bad upload body: data is not base64")
		}
		data = decoded
	} else {
		raw, err := io.ReadAll(http.MaxBytesReader(nil, r.Body, limit))
		if err != nil {
			return nil, err
		}
		data = raw
	}
	if len(data) == 0 {
		return nil, errEmptyUpload
	}
	if int64(len(data)) > limit {
		return nil, errors.New("upload is too large")
	}
	return data, nil
}
