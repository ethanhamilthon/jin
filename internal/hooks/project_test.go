package hooks

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectHooksNeedTrust(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path, err := CreateProject(dir, "lint")
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(path, []byte("Run make lint."), 0o644)
	if got := ActiveIn(dir, nil, false); len(got) != 0 {
		t.Fatalf("untrusted project hooks loaded: %+v", got)
	}
	got := ActiveIn(dir, nil, true)
	if len(got) != 1 || !got[0].Project || got[0].Body != "Run make lint." {
		t.Fatalf("got %+v", got)
	}
	if got := ActiveIn(dir, []string{ProjectKey(dir, "lint")}, true); len(got) != 0 {
		t.Fatalf("disabled project hook loaded: %+v", got)
	}
}

func TestAddFromURLAndPath(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Use searchctl."))
	}))
	defer server.Close()
	path, err := Add(context.Background(), server.URL+"/x/searchctl.md", "", "")
	if err != nil || filepath.Base(path) != "searchctl.md" {
		t.Fatalf("url: %s %v", path, err)
	}
	if _, err := Add(context.Background(), server.URL+"/x/searchctl.md", "", ""); err == nil {
		t.Fatal("must refuse to overwrite")
	}
	project := t.TempDir()
	var out, errOut bytes.Buffer
	if code := Main(context.Background(), []string{"add", path, "--name", "web", "--project"}, project, &out, &errOut); code != 0 {
		t.Fatalf("code %d: %s", code, errOut.String())
	}
	if data, _ := os.ReadFile(filepath.Join(project, ".jin", "hooks", "web.md")); string(data) != "Use searchctl." {
		t.Fatalf("project copy = %q", data)
	}
	out.Reset()
	Main(context.Background(), []string{"list"}, project, &out, &errOut)
	if !strings.Contains(out.String(), "searchctl\n") || !strings.Contains(out.String(), "web\t(project)") {
		t.Fatalf("list:\n%s", out.String())
	}
}
