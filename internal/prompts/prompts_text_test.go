package prompts

import (
	"strings"
	"testing"
)

func TestPlanPromptReadOnlyAndNumbered(t *testing.T) {
	body, ok := systemBody("plan")
	if !ok {
		t.Fatal("plan prompt not found")
	}
	if strings.Contains(body, "go list") {
		t.Error("plan prompt should not include go list among read-only commands")
	}
	if strings.Contains(body, "`todo`") {
		t.Error("plan prompt should not mention the removed todo tool")
	}
	if !strings.Contains(body, "numbered list") {
		t.Error("plan prompt should ask for the plan as a numbered list")
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
