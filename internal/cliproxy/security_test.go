package cliproxy

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestPrivateKeysAndConfiguration(t *testing.T) {
	root := t.TempDir()
	keys, err := loadKeys(root)
	if err != nil {
		t.Fatal(err)
	}
	if keys.Client == keys.Management || keys.Client == keys.Control || keys.Management == keys.Control {
		t.Fatal("keys reused")
	}
	again, err := loadKeys(root)
	if err != nil || again != keys {
		t.Fatal("keys not stable")
	}
	info, _ := os.Stat(root + "/keys.json")
	if info.Mode().Perm() != 0o600 {
		t.Fatal("keys not private")
	}
	config := configuration(root, 8317, keys)
	if config["host"] != "127.0.0.1" || config["force-model-prefix"] != true {
		t.Fatal("unsafe listener/routing")
	}
	data, _ := json.Marshal(config)
	if !strings.Contains(string(data), `"disable-claude-cloak-mode":true`) {
		t.Fatal("prompt cloaking not disabled")
	}
	if config["plugins"].(map[string]any)["enabled"] != false {
		t.Fatal("plugins enabled")
	}
}

func TestProxyEnvironmentDoesNotImportOtherCredentialStores(t *testing.T) {
	t.Setenv("PG_DSN", "private")
	t.Setenv("GITHUB_TOKEN", "private")
	t.Setenv("ANTHROPIC_API_KEY", "private")
	for _, entry := range processEnv(t.TempDir()) {
		if strings.Contains(entry, "private") {
			t.Fatal("inherited credential store")
		}
	}
}
