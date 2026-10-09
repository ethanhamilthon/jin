package web

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

const pairTTL = 5 * time.Minute

// pairCodes are one-time codes: a scanned QR code turns into a device once.
type pairCodes struct {
	mu    sync.Mutex
	codes map[string]time.Time
}

func newPairCodes() *pairCodes { return &pairCodes{codes: map[string]time.Time{}} }

func (p *pairCodes) issue() (string, time.Time) {
	var raw [16]byte
	rand.Read(raw[:])
	code, expires := hex.EncodeToString(raw[:]), time.Now().Add(pairTTL)
	p.mu.Lock()
	defer p.mu.Unlock()
	for old, at := range p.codes {
		if time.Now().After(at) {
			delete(p.codes, old)
		}
	}
	p.codes[code] = expires
	return code, expires
}

// take burns a code and says whether it was still good.
func (p *pairCodes) take(code string) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	expires, ok := p.codes[code]
	delete(p.codes, code)
	return ok && time.Now().Before(expires)
}

func (g *guard) pair(w http.ResponseWriter, r *http.Request, known bool) {
	if !g.codes.take(r.URL.Query().Get("code")) {
		http.Error(w, "This QR code has expired or was already used. Scan a new one in jin web.", http.StatusUnauthorized)
		return
	}
	g.admit(w, r, known)
}

// admit gives the browser a device cookie, unless it already has one, and
// sends it to the start page.
func (g *guard) admit(w http.ResponseWriter, r *http.Request, known bool) {
	if !known {
		_, token, err := g.db.AddDevice(deviceName(r.UserAgent()))
		if err != nil {
			http.Error(w, "could not register the device", http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{Name: cookieName, Value: token, Path: "/", HttpOnly: true,
			Secure: g.remote[r.Host], SameSite: http.SameSiteStrictMode, MaxAge: int(cookieAge.Seconds())})
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
