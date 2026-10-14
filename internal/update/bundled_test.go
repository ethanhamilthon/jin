package update

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUpdateRefusesInsideTheDesktopApp(t *testing.T) {
	target := filepath.Join(t.TempDir(), "Jin.app", "Contents", "Resources", "jin")
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old"), 0o755); err != nil {
		t.Fatal(err)
	}
	old := executable
	defer func() { executable = old }()
	executable = func() (string, error) { return target, nil }
	for _, args := range [][]string{nil, {"--check"}, {"--force"}} {
		var out, errOut strings.Builder
		if code := Main(t.Context(), args, "v0.0.1", Control{}, &out, &errOut); code != 1 {
			t.Fatalf("%v: code %d", args, code)
		}
		if !strings.Contains(errOut.String(), "belongs to the Jin desktop app") {
			t.Fatalf("%v: message = %q", args, errOut.String())
		}
	}
	if data, _ := os.ReadFile(target); string(data) != "old" {
		t.Fatalf("binary changed: %q", data)
	}
}
