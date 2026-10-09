package web

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"time"

	"jin/internal/store"
)

const (
	cookieName = "jin_device"
	cookieAge  = 400 * 24 * time.Hour
)

type deviceKey struct{}

// deviceID is the device a request came from; empty when the request did
// not pass the guard, as in tests.
func deviceID(r *http.Request) string {
	id, _ := r.Context().Value(deviceKey{}).(string)
	return id
}

// guard lets in only this page: the Host must be one of our addresses, a
// browser request from another origin is refused, and every request needs
// the cookie of a device. The start token in the printed address and a
// pairing code both turn into such a cookie once.
type guard struct {
	token  string
	hosts  map[string]bool
	remote map[string]bool
	db     *store.DB
	codes  *pairCodes
}

// newGuard accepts the loopback addresses over http and every name in
// allowed over https, the way a reverse proxy such as `tailscale serve`
// brings the page.
func newGuard(port string, allowed []string, db *store.DB) *guard {
	var buf [24]byte
	rand.Read(buf[:])
	g := &guard{token: hex.EncodeToString(buf[:]), db: db, codes: newPairCodes(), remote: map[string]bool{}, hosts: map[string]bool{
		"127.0.0.1:" + port: true, "localhost:" + port: true,
	}}
	for _, name := range allowed {
		g.remote[name] = true
	}
	return g
}

func (g *guard) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !g.hosts[r.Host] && !g.remote[r.Host] {
			http.Error(w, "unknown host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && !g.sameOrigin(origin) {
			http.Error(w, "cross-origin request refused", http.StatusForbidden)
			return
		}
		device, known := g.device(r)
		if r.URL.Path == "/pair" {
			g.pair(w, r, known)
			return
		}
		if token := r.URL.Query().Get("token"); token != "" && g.valid(token) {
			g.admit(w, r, known)
			return
		}
		if !known {
			http.Error(w, "Open the address that jin web printed, or scan a new QR code in jin web.", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), deviceKey{}, device.ID)))
	})
}

func (g *guard) device(r *http.Request) (store.Device, bool) {
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return store.Device{}, false
	}
	return g.db.DeviceByToken(cookie.Value)
}

func (g *guard) valid(token string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(g.token)) == 1
}

func (g *guard) sameOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Scheme == "http" && g.hosts[u.Host] || u.Scheme == "https" && g.remote[u.Host]
}
