package cliproxy

import "testing"

func TestPatchOfOrdersCompatibleVersions(t *testing.T) {
	if patchOf("v8.0.30") <= patchOf("v8.0.23") || patchOf(DefaultVersion) != 23 {
		t.Fatal("patch numbers must compare as numbers")
	}
}
