package prompts

import (
	"strings"
	"testing"
)

func TestPlanPromptReadOnlyAndStatus(t *testing.T) {
	body, ok := systemBody("plan")
	if !ok {
		t.Fatal("plan prompt not found")
	}
	if strings.Contains(body, "go list") {
		t.Error("plan prompt should not include go list among read-only commands")
	}
	if !strings.Contains(body, "Existing items keep their status; every new item you create is `pending`") {
		t.Error("plan prompt should preserve existing item status while making new items pending")
	}
	if strings.Contains(body, "Read the current list: call `todo` without `items`") {
		t.Error("plan prompt should not require calling todo without items unconditionally")
	}
}

func TestReviewPromptConfidenceAndScope(t *testing.T) {
	body, ok := systemBody("review")
	if !ok {
		t.Fatal("review prompt not found")
	}
	if !strings.Contains(body, "Confidence rule") {
		t.Error("review prompt missing confidence rule")
	}
	if !strings.Contains(body, "`ready` within the reviewed scope") {
		t.Error("review prompt should define ready within the reviewed scope")
	}
}
