package web

import (
	"encoding/json"
	"errors"
	"net/http"
)

type handler func(r *http.Request) (any, error)

// api turns a handler's value into JSON and its error into a 400 with
// {"error": text}.
func api(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		value, err := h(r)
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		if value == nil {
			value = map[string]bool{"ok": true}
		}
		_ = json.NewEncoder(w).Encode(value)
	}
}

// decode reads a JSON body into v; an empty body leaves v as it is.
func decode(r *http.Request, v any) error {
	if r.ContentLength == 0 {
		return nil
	}
	if err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 8<<20)).Decode(v); err != nil {
		return errors.New("bad request body: " + err.Error())
	}
	return nil
}

func done(err error) (any, error) { return nil, err }
