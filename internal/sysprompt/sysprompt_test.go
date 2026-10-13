package sysprompt

import (
	"jin/internal/provider"
	"os"
	"strings"
	"testing"
)

func home(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
}

func write(t *testing.T, text string) {
	t.Helper()
	path, _ := Path()
	if err := os.MkdirAll(strings.TrimSuffix(path, "/system-prompt.md"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestTemplateIdentityIsInterfaceNeutral(t *testing.T) {
	if got := Defaults().System; !strings.Contains(got, "running in the user's environment") || strings.Contains(got, "running in the user's terminal") {
		t.Fatalf("identity is not interface-neutral: %q", got[:strings.Index(got, "\n")])
	}
}

func TestCustomFileIsIgnoredAndPreserved(t *testing.T) {
	home(t)
	write(t, "# system\nCUSTOM_SENTINEL {{touch /tmp/jin-custom-prompt-ran}}\n# compact\ncustom compact\n# handoff\ncustom handoff\n")
	path, _ := Path()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil || got.Custom || got != Defaults() {
		t.Fatalf("Load = %+v, %v, want built-in defaults", got, err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Fatalf("custom file changed: %q, %v", after, err)
	}
}

func TestNoFileMeansDefaults(t *testing.T) {
	home(t)
	got, err := Load()
	if err != nil || got.Custom {
		t.Fatalf("Load = %+v, %v", got, err)
	}
	if got.System != Defaults().System || got.Compact == "" || got.Handoff == "" {
		t.Errorf("defaults not used: %+v", got)
	}
	for _, want := range []string{"{{jin docs}}", "{{jin hooks render}}", "AGENTS.md", provider.CacheBreak, "{{date +%F}}"} {
		if !strings.Contains(got.System, want) {
			t.Errorf("the default system prompt lacks %q:\n%s", want, got.System)
		}
	}
	if strings.Index(got.System, "AGENTS.md") > strings.Index(got.System, provider.CacheBreak) || strings.Index(got.System, "{{date") < strings.Index(got.System, provider.CacheBreak) {
		t.Errorf("stable text must come before the cache break, live data after it:\n%s", got.System)
	}
}

func TestCreateNeverCreatesLegacyFile(t *testing.T) {
	home(t)
	path, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Create created legacy file: %v", err)
	}
}

func TestLegacyFileNeverAffectsLoadedTemplates(t *testing.T) {
	home(t)
	write(t, "ignored preamble\n\n# system\nCUSTOM_SENTINEL\n# compact\ncustom compact\n# handoff\ncustom handoff\n")
	got, err := Load()
	if err != nil || got != Defaults() {
		t.Fatalf("Load = %+v, %v, want built-ins", got, err)
	}
}

func TestEmptyLegacySectionsDoNotChangeDefaults(t *testing.T) {
	home(t)
	write(t, "# system\n\n   \n# compact\ncustom compact\n# handoff\n")
	got, _ := Load()
	if got != Defaults() {
		t.Errorf("got %+v", got)
	}
}

func TestCustomFileRemainsUnchangedAfterLoad(t *testing.T) {
	home(t)
	write(t, "# system\nmine\n")
	path, _ := Path()
	before, _ := os.ReadFile(path)
	if _, err := Load(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Errorf("custom file changed: %q, %v", after, err)
	}
}

func TestLegacySectionNamesCannotOverrideBuiltIns(t *testing.T) {
	home(t)
	write(t, "# system\nline one\n# systematic\nline two\n#system\nline three\n")
	got, _ := Load()
	if got != Defaults() {
		t.Errorf("got %+v", got)
	}
}

func TestLegacyFileFormattingDoesNotMatter(t *testing.T) {
	home(t)
	write(t, "# system  \r\nmine\r\n# compact\t\r\nmy compact\r\n")
	got, _ := Load()
	if got != Defaults() {
		t.Errorf("got %+v", got)
	}
}
