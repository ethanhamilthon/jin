package cliproxy

import (
	"crypto/subtle"
	"net/http"
)

func (b *broker) notify() {
	select {
	case b.wake <- struct{}{}:
	default:
	}
}
func (b *broker) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /client", b.register)
	mux.HandleFunc("GET /lease", b.lease)
	mux.HandleFunc("POST /command", b.command)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("Authorization")), []byte("Bearer "+b.keys.Control)) != 1 {
			http.Error(w, "unauthorized", 401)
			return
		}
		mux.ServeHTTP(w, r)
	})
}
