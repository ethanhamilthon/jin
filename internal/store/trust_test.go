package store

import "testing"

func TestHooksTrust(t *testing.T) {
	db := openTest(t)
	if trust, _ := db.HooksTrust("/repo"); trust != TrustUnknown {
		t.Fatalf("default = %v", trust)
	}
	_ = db.SaveHooksTrust("/repo", true)
	_ = db.SaveHooksTrust("/other", false)
	if trust, _ := db.HooksTrust("/repo"); trust != Trusted {
		t.Errorf("repo = %v", trust)
	}
	if trust, _ := db.HooksTrust("/other"); trust != Distrusted {
		t.Errorf("other = %v", trust)
	}
}
