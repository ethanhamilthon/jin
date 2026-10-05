package prompts

import (
	"strings"
	"testing"
)

func TestSubagentsLaunchRecipeFileRedirection(t *testing.T) {
	body, ok := systemBody("subagents")
	if !ok {
		t.Fatal("subagents prompt not found")
	}
	wantCmd := `jin -p --no-session --model <id> --timeout 20m < /tmp/jin-task-<name>.md`
	if !strings.Contains(body, wantCmd) || !strings.Contains(body, "`task` tool, action `start`") {
		t.Errorf("subagents missing the task launch with file redirection:\n%s", body)
	}
	if strings.Contains(body, "jin async") {
		t.Error("subagents still mention jin async")
	}
}

func TestSubagentsModelSelection(t *testing.T) {
	body, ok := systemBody("subagents")
	if !ok {
		t.Fatal("subagents prompt not found")
	}
	if !strings.Contains(body, "Use the model the user named or the model of this chat without asking") {
		t.Error("subagents should use named or current chat model without asking")
	}
	if !strings.Contains(body, "Ask with `ask_user` only once when the user asked to choose") {
		t.Error("subagents should ask only once when user requested choice")
	}
}

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
