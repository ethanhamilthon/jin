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

func TestCreateWritesAllThreeSectionsOnce(t *testing.T) {
	home(t)
	path, err := Create()
	if err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	for _, want := range []string{"# system\n", "# compact\n", "# handoff\n"} {
		if !strings.Contains(string(data), want) {
			t.Errorf("file lacks %q", want)
		}
	}
	loaded, _ := Load()
	if !loaded.Custom || loaded.System != Defaults().System || loaded.Compact != Defaults().Compact || loaded.Handoff != Defaults().Handoff {
		t.Errorf("a fresh file must read back as the defaults: %+v", loaded)
	}
	write(t, "# system\n\nmine\n")
	if _, err := Create(); err != nil {
		t.Fatal(err)
	}
	if again, _ := Load(); again.System != "mine" {
		t.Error("Create must not overwrite an existing file")
	}
	if info, _ := os.Stat(path); info.Mode().Perm() != 0o600 {
		t.Errorf("mode = %v", info.Mode().Perm())
	}
}

func TestSectionsAreReadAndMissingOnesFallBack(t *testing.T) {
	home(t)
	write(t, "ignored preamble\n\n# system\n\nYou are mine.\n\n## A heading stays in the section\nmore\n\n# handoff\nbrief it\n")
	got, _ := Load()
	if !got.Custom {
		t.Error("Custom must be set")
	}
	if got.System != "You are mine.\n\n## A heading stays in the section\nmore" {
		t.Errorf("system = %q", got.System)
	}
	if got.Handoff != "brief it" {
		t.Errorf("handoff = %q", got.Handoff)
	}
	if got.Compact != Defaults().Compact {
		t.Error("a section that is not in the file must use the default")
	}
}

func TestEmptySectionFallsBackToTheDefault(t *testing.T) {
	home(t)
	write(t, "# system\n\n   \n# compact\ncustom compact\n# handoff\n")
	got, _ := Load()
	if got.System != Defaults().System || got.Handoff != Defaults().Handoff || got.Compact != "custom compact" {
		t.Errorf("got %+v", got)
	}
}

func TestDeletedFileResets(t *testing.T) {
	home(t)
	write(t, "# system\nmine\n")
	path, _ := Path()
	os.Remove(path)
	if got, _ := Load(); got.Custom || got.System != Defaults().System {
		t.Errorf("got %+v", got)
	}
}

func TestSectionLineMustMatchExactly(t *testing.T) {
	home(t)
	write(t, "# system\nline one\n# systematic\nline two\n#system\nline three\n")
	got, _ := Load()
	if got.System != "line one\n# systematic\nline two\n#system\nline three" {
		t.Errorf("system = %q", got.System)
	}
}

func TestWindowsLineEndingsAndTrailingSpaces(t *testing.T) {
	home(t)
	write(t, "# system  \r\nmine\r\n# compact\t\r\nmy compact\r\n")
	got, _ := Load()
	if got.System != "mine" || got.Compact != "my compact" {
		t.Errorf("got %+v", got)
	}
}
