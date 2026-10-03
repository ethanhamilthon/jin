package provider

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugOffByDefault(t *testing.T) {
	if NewClient(Config{}).debug != nil {
		t.Fatal("debug must be opt-in")
	}
}

func TestDebugLogsMetadataOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "d", "log.jsonl")
	t.Setenv("JIN_DEBUG", path)
	client, _ := responsesServer(t, responsesDone)
	client.Configure(Config{Kind: KindResponses, BaseURL: client.Config().BaseURL, APIKey: "sk-super-secret"})
	history := []Message{{Role: "system", Content: "SYSTEM-PROMPT"}, {Role: "user", Content: "USER-PROMPT"}}
	for range 2 {
		if _, err := client.Stream(t.Context(), "m", "high", history, json.RawMessage(testToolsSchema), func(StreamEvent) {}); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"sk-super-secret", "SYSTEM-PROMPT", "USER-PROMPT", "ENC-BLOB", `\"path\"`, "hello"} {
		if strings.Contains(string(data), secret) {
			t.Fatalf("log leaks %q", secret)
		}
	}
	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("mode %v", info.Mode())
	}
	events := map[string][]map[string]any{}
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		var line map[string]any
		if err := json.Unmarshal(scanner.Bytes(), &line); err != nil {
			t.Fatalf("bad line %s", scanner.Text())
		}
		events[line["event"].(string)] = append(events[line["event"].(string)], line)
	}
	for _, name := range []string{"attempt_start", "request", "http_headers", "first_byte", "raw_usage", "attempt_end"} {
		if len(events[name]) != 2 {
			t.Fatalf("event %s count %d: %s", name, len(events[name]), data)
		}
	}
	if events["request"][1]["prefix_unchanged"] != true || events["attempt_start"][0]["silence_timeout_ms"] != float64(300000) {
		t.Fatalf("request metadata: %v %v", events["request"][1], events["attempt_start"][0])
	}
	if events["attempt_end"][1]["cache_miss_requests"] != float64(2) {
		t.Fatalf("cache counters: %v", events["attempt_end"][1])
	}
}

func TestDebugRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target, link := filepath.Join(dir, "target"), filepath.Join(dir, "link")
	if err := os.WriteFile(target, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Skip(err)
	}
	t.Setenv("JIN_DEBUG", link)
	NewClient(Config{}).Debug("x", map[string]any{})
	if data, _ := os.ReadFile(target); len(data) != 0 {
		t.Fatal("symlink must not be followed")
	}
}
