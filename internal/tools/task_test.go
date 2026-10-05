//go:build unix

package tools

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestTaskToolStartCheckStop(t *testing.T) {
	ctx, m := backgroundCtx(t, nil)
	tool := Task{dir: t.TempDir()}
	out, err := tool.Run(ctx, `{"action":"start","command":"echo hi; sleep 30"}`)
	if err != nil || !strings.HasPrefix(out, "started task ") {
		t.Fatalf("start: %q %v", out, err)
	}
	id := m.Running("s1")[0].ID
	time.Sleep(200 * time.Millisecond)
	out, err = tool.Run(ctx, `{"action":"check","id":"`+id+`","limit":100}`)
	if err != nil || !strings.Contains(out, "running") || !strings.Contains(out, "hi") {
		t.Fatalf("check: %q %v", out, err)
	}
	if out, _ = tool.Run(ctx, `{"action":"list"}`); !strings.Contains(out, id) {
		t.Fatalf("list: %q", out)
	}
	if _, err = tool.Run(ctx, `{"action":"stop","id":"`+id+`"}`); err != nil {
		t.Fatalf("stop: %v", err)
	}
}

func TestTaskToolValidates(t *testing.T) {
	ctx, _ := backgroundCtx(t, nil)
	for _, args := range []string{`{"action":"start"}`, `{"action":"check"}`, `{"action":"jump"}`, `{"action":"start","command":"x","dir":"nope"}`} {
		if _, err := (Task{dir: t.TempDir()}).Run(ctx, args); err == nil {
			t.Errorf("%s: expected an error", args)
		}
	}
	if _, err := (Task{}).Run(context.Background(), `{"action":"list"}`); err == nil {
		t.Error("expected an error without background tasks")
	}
}
