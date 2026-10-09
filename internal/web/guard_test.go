package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"jin/internal/store"
)

func testDB(t *testing.T) *store.DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func get(g *guard, host, target string, cookie *http.Cookie) *httptest.ResponseRecorder {
	r := httptest.NewRequest("GET", target, nil)
	r.Host = host
	r.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0) AppleWebKit Version/18.0 Mobile Safari/604.1")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	g.wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })).ServeHTTP(w, r)
	return w
}

func TestPairingCodeWorksOnce(t *testing.T) {
	db := testDB(t)
	g := newGuard("7373", []string{"mac.tail.ts.net"}, db)
	code, _ := g.codes.issue()
	w := get(g, "mac.tail.ts.net", "/pair?code="+code, nil)
	cookies := w.Result().Cookies()
	if w.Code != http.StatusSeeOther || len(cookies) != 1 || !cookies[0].Secure {
		t.Fatalf("first use: code %d, cookies %v", w.Code, cookies)
	}
	if got := get(g, "mac.tail.ts.net", "/api/x", cookies[0]).Code; got != http.StatusTeapot {
		t.Fatalf("device cookie: code %d", got)
	}
	if got := get(g, "mac.tail.ts.net", "/pair?code="+code, nil).Code; got != http.StatusUnauthorized {
		t.Fatalf("second use: code %d", got)
	}
	devices, _ := db.Devices()
	if len(devices) != 1 || devices[0].Name != "iPhone, Safari" {
		t.Fatalf("devices: %+v", devices)
	}
}

func TestRevokedDeviceIsRefused(t *testing.T) {
	db := testDB(t)
	g := newGuard("7373", nil, db)
	w := get(g, "127.0.0.1:7373", "/?token="+g.token, nil)
	cookie := w.Result().Cookies()[0]
	devices, _ := db.Devices()
	_ = db.RemoveDevice(devices[0].ID)
	if got := get(g, "127.0.0.1:7373", "/api/x", cookie).Code; got != http.StatusUnauthorized {
		t.Fatalf("revoked: code %d", got)
	}
}

func TestStartTokenInOldBrowserAddsNoDevice(t *testing.T) {
	db := testDB(t)
	g := newGuard("7373", nil, db)
	cookie := get(g, "127.0.0.1:7373", "/?token="+g.token, nil).Result().Cookies()[0]
	get(g, "127.0.0.1:7373", "/?token="+g.token, cookie)
	if devices, _ := db.Devices(); len(devices) != 1 {
		t.Fatalf("devices: %d, want 1", len(devices))
	}
}
