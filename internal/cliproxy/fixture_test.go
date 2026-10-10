package cliproxy

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if len(os.Args) == 3 && os.Args[1] == "--internal-cliproxy" {
		if err := Serve(os.Args[2]); err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}
	if len(os.Args) == 4 && os.Args[1] == "-config" {
		fakeProxy(os.Args[2])
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func fakeProxy(path string) {
	var cfg struct {
		Host       string
		Port       int
		APIKeys    []string `json:"api-keys"`
		Management struct {
			Key string `json:"secret-key"`
		} `json:"remote-management"`
	}
	if readJSON(path, &cfg) != nil {
		os.Exit(2)
	}
	version := filepath.Base(filepath.Dir(os.Args[0]))
	if version == "v8.0.25" && !strings.Contains(path, ".probe-") {
		os.Exit(4)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v0/management/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CPA-VERSION", version)
		if r.Header.Get("Authorization") != "Bearer "+cfg.Management.Key {
			w.WriteHeader(401)
			return
		}
		fakeManagement(w, r)
	})
	mux.HandleFunc("/v1/models", func(w http.ResponseWriter, r *http.Request) {
		if len(cfg.APIKeys) == 0 || r.Header.Get("Authorization") != "Bearer "+cfg.APIKeys[0] {
			w.WriteHeader(401)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]string{{"id": "codex/shared"}, {"id": "claude/shared"}}})
	})
	mux.HandleFunc("/v1/test-exit", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+cfg.APIKeys[0] {
			w.WriteHeader(401)
			return
		}
		w.WriteHeader(200)
		w.(http.Flusher).Flush()
		go func() { time.Sleep(50 * time.Millisecond); os.Exit(1) }()
	})
	if err := http.ListenAndServe(cfg.Host+":"+itoa(cfg.Port), mux); err != nil {
		os.Exit(3)
	}
}

func installFixture(t *testing.T, root, version string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(binaryPath(root, version)), 0o700); err != nil {
		t.Fatal(err)
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(exe, binaryPath(root, version)); err != nil {
		t.Fatal(err)
	}
	if err = writeJSON(filepath.Join(root, "installed.json"), Installation{Version: version}); err != nil {
		t.Fatal(err)
	}
}
