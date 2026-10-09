package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
)

const cookieName = "jin_token"

// guard lets in only this page: the Host must be our loopback address, a
// browser request from another origin is refused, and every request needs
// the token, which the URL brings once and a cookie keeps.
type guard struct {
	token  string
	hosts  map[string]bool
	remote map[string]bool
}

// newGuard accepts the loopback addresses over http and every name in
// allowed over https, the way a reverse proxy such as `tailscale serve`
// brings the page.
func newGuard(port string, allowed []string) guard {
	var buf [24]byte
	rand.Read(buf[:])
	g := guard{token: hex.EncodeToString(buf[:]), remote: map[string]bool{}, hosts: map[string]bool{
		"127.0.0.1:" + port: true, "localhost:" + port: true,
	}}
	for _, name := range allowed {
		g.remote[name] = true
	}
	return g
}

func (g guard) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !g.hosts[r.Host] && !g.remote[r.Host] {
			http.Error(w, "unknown host", http.StatusForbidden)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && !g.sameOrigin(origin) {
			http.Error(w, "cross-origin request refused", http.StatusForbidden)
			return
		}
		if token := r.URL.Query().Get("token"); token != "" && g.valid(token) {
			http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true, Secure: g.remote[r.Host], SameSite: http.SameSiteStrictMode})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		if cookie, err := r.Cookie(cookieName); err != nil || !g.valid(cookie.Value) {
			http.Error(w, "Open the address that jin web printed; it carries the access token.", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (g guard) valid(token string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(g.token)) == 1
}

func (g guard) sameOrigin(origin string) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return u.Scheme == "http" && g.hosts[u.Host] || u.Scheme == "https" && g.remote[u.Host]
}
