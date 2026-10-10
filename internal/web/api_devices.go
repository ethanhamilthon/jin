package web

import (
	"encoding/base64"
	"errors"
	"net/http"

	"github.com/skip2/go-qrcode"

	"jin/internal/store"
)

type deviceView struct {
	store.Device
	Online  bool `json:"online"`
	Current bool `json:"current"`
}

func (s *server) deviceRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/devices", api(func(r *http.Request) (any, error) {
		devices, err := s.db.Devices()
		if err != nil {
			return nil, err
		}
		online, views := s.hub.online(), []deviceView{}
		for _, d := range devices {
			views = append(views, deviceView{Device: d, Online: online[d.ID], Current: d.ID == deviceID(r)})
		}
		host := s.remoteHost
		if s.remote != nil {
			state := s.remote.state()
			host = state.URL
		}
		return map[string]any{"remote": host != "", "devices": views}, nil
	}))
	mux.HandleFunc("POST /api/devices/pair", api(s.newPairing))
	mux.HandleFunc("PATCH /api/devices/{id}", api(func(r *http.Request) (any, error) {
		var body struct{ Name string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		return done(s.db.RenameDevice(r.PathValue("id"), body.Name))
	}))
	mux.HandleFunc("DELETE /api/devices/{id}", api(func(r *http.Request) (any, error) {
		id := r.PathValue("id")
		if err := s.db.RemoveDevice(id); err != nil {
			return nil, err
		}
		s.hub.kick(id)
		return nil, nil
	}))
}

// newPairing makes a one-time address for a phone and its QR code.
func (s *server) newPairing(r *http.Request) (any, error) {
	host := s.remoteHost
	if s.remote != nil {
		s.remote.mu.Lock()
		defer s.remote.mu.Unlock()
		host = s.remote.host
	}
	if host == "" || s.guard == nil {
		return nil, errors.New("remote access is off; enable it in Settings / Remote access")
	}
	code, expires := s.guard.codes.issue()
	link := "https://" + host + "/pair?code=" + code
	png, err := qrcode.Encode(link, qrcode.Medium, 280)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"url": link, "expires": expires.Unix(),
		"qr": "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
	}, nil
}
