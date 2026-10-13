package desktop

import (
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
)

var tokenPattern = regexp.MustCompile(`^[a-f0-9]{48}$`)

const shellOrigin = "wails://wails"

func parseUpstream(raw string) (*url.URL, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.User != nil || u.Fragment != "" || u.Path != "/" || u.RawPath != "" {
		return nil, errors.New("daemon returned an invalid local service URL")
	}
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil || (host != "127.0.0.1" && host != "localhost") || port == "" || len(u.Query()) != 1 || !tokenPattern.MatchString(u.Query().Get("token")) {
		return nil, errors.New("daemon returned an invalid local service URL")
	}
	return u, nil
}

func shellOriginAllowed(origin string) bool { return origin == "" || origin == shellOrigin }

var passedHeaders = []string{"Accept", "Content-Type", "Last-Event-ID", "Range", "If-None-Match"}

func upstreamRequest(r *http.Request, to *link) (*http.Request, error) {
	u := *r.URL
	u.Scheme, u.Host = to.base.Scheme, to.base.Host
	q := u.Query()
	q.Del("token")
	u.RawQuery = q.Encode()
	body := r.Body
	if body != nil {
		body = http.MaxBytesReader(nil, body, maxBody)
	}
	req, err := http.NewRequestWithContext(r.Context(), r.Method, u.String(), body)
	if err != nil {
		return nil, err
	}
	req.ContentLength = r.ContentLength
	for _, key := range passedHeaders {
		if value := r.Header.Get(key); value != "" {
			req.Header.Set(key, value)
		}
	}
	req.Header.Set("Cookie", "jin_device="+to.cookie)
	req.Header.Set("Origin", to.base.Scheme+"://"+to.base.Host)
	return req, nil
}
