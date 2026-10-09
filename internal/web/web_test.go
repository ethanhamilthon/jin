package web

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseArgs(t *testing.T) {
	opt, err := ParseArgs([]string{"--no-open", "--port", "8123", "--cwd", "/tmp"})
	if err != nil || opt.Port != 8123 || !opt.NoOpen || opt.Cwd != "/tmp" {
		t.Fatalf("opt = %+v, err = %v", opt, err)
	}
	for _, bad := range [][]string{{"--port"}, {"--port", "x"}, {"--port", "70000"}, {"--what"}} {
		if _, err := ParseArgs(bad); err == nil {
			t.Errorf("ParseArgs(%v) accepted", bad)
		}
	}
}

func TestListenFallsBackWhenTheDefaultIsBusy(t *testing.T) {
	busy, err := net.Listen("tcp", address(defaultPort))
	if err != nil {
		t.Skip("default port in use by another program")
	}
	defer busy.Close()
	l, err := listen(0)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	if l.Addr().(*net.TCPAddr).Port == defaultPort {
		t.Fatal("took the busy port")
	}
	if _, err := listen(defaultPort); err == nil {
		t.Fatal("an explicit busy port must be an error")
	}
}

func TestGuard(t *testing.T) {
	db := testDB(t)
	g := newGuard("7373", []string{"mac.tail.ts.net"}, db)
	_, dev, _ := db.AddDevice("test")
	ok := g.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	cases := []struct {
		name, host, origin, cookie, query string
		want                              int
	}{
		{"no token", "127.0.0.1:7373", "", "", "", http.StatusUnauthorized},
		{"token in url", "127.0.0.1:7373", "", "", "?token=" + g.token, http.StatusSeeOther},
		{"cookie", "localhost:7373", "", dev, "", http.StatusTeapot},
		{"wrong cookie", "127.0.0.1:7373", "", "x", "", http.StatusUnauthorized},
		{"rebound host", "evil.test:7373", "", dev, "", http.StatusForbidden},
		{"other origin", "127.0.0.1:7373", "http://evil.test", dev, "", http.StatusForbidden},
		{"same origin", "127.0.0.1:7373", "http://127.0.0.1:7373", dev, "", http.StatusTeapot},
		{"allowed host", "mac.tail.ts.net", "https://mac.tail.ts.net", dev, "", http.StatusTeapot},
		{"allowed host, http origin", "mac.tail.ts.net", "http://mac.tail.ts.net", dev, "", http.StatusForbidden},
		{"allowed host, other origin", "mac.tail.ts.net", "https://evil.test", dev, "", http.StatusForbidden},
		{"unlisted host", "other.ts.net", "", dev, "", http.StatusForbidden},
	}
	for _, c := range cases {
		r := httptest.NewRequest("POST", "/api/x"+c.query, nil)
		r.Host = c.host
		if c.origin != "" {
			r.Header.Set("Origin", c.origin)
		}
		if c.cookie != "" {
			r.AddCookie(&http.Cookie{Name: cookieName, Value: c.cookie})
		}
		w := httptest.NewRecorder()
		ok.ServeHTTP(w, r)
		if w.Code != c.want {
			t.Errorf("%s: code %d, want %d", c.name, w.Code, c.want)
		}
	}
}
