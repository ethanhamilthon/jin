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
	wantCmd := `jin async run "jin -p --no-session --model <id> --timeout 20m < /tmp/jin-task-<name>.md" --session <your session id>`
	if !strings.Contains(body, wantCmd) {
		t.Errorf("subagents missing safe file redirection command:\n%s", body)
	}
	if strings.Contains(body, "Tell the sub-agent to finish by running `jin async run \"echo") {
		t.Error("subagents should not instruct echo on finish")
	}
	if !strings.Contains(body, "Echo is only for a blocking question") {
		t.Error("subagents should limit echo to blocking questions")
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
