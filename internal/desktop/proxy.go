//go:build desktop

package desktop

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
)

const maxBody = 64 << 20

var returnedHeaders = []string{"Content-Type", "Content-Length", "Content-Range", "Accept-Ranges", "Content-Disposition", "Cache-Control", "Etag"}

type proxy struct {
	to     atomic.Pointer[link]
	failed atomic.Pointer[string]
}

func (p *proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !shellOriginAllowed(r.Header.Get("Origin")) {
		http.Error(w, "desktop request refused", http.StatusForbidden)
		return
	}
	to := p.to.Load()
	if to == nil {
		message := "Jin is starting"
		if failure := p.failed.Load(); failure != nil {
			message = *failure
		}
		http.Error(w, message, http.StatusServiceUnavailable)
		return
	}
	req, err := upstreamRequest(r, to)
	if err != nil {
		http.Error(w, "invalid desktop request", http.StatusBadRequest)
		return
	}
	resp, err := to.client.Do(req)
	if err != nil {
		http.Error(w, "daemon request failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		http.Error(w, "daemon redirect refused", http.StatusBadGateway)
		return
	}
	for _, key := range returnedHeaders {
		if value := resp.Header.Get(key); value != "" {
			w.Header().Set(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "text/event-stream") {
		streamBody(w, resp.Body)
		return
	}
	_, _ = io.Copy(w, resp.Body)
}

func streamBody(w http.ResponseWriter, body io.Reader) {
	flusher, _ := w.(http.Flusher)
	buf := make([]byte, 8<<10)
	for {
		n, err := body.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if err != nil {
			return
		}
	}
}
