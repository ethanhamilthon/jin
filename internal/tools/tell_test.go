package tools

import (
	"context"
	"testing"
)

func TestTellReachesTheUser(t *testing.T) {
	var gotMode, gotText string
	ctx := WithTeller(context.Background(), func(mode, text string) { gotMode, gotText = mode, text })
	out, err := Tell{}.Run(ctx, `{"mode":"suggest","text":"  Run the migrations  "}`)
	if err != nil || out != "suggestion set" || gotMode != "suggest" || gotText != "Run the migrations" {
		t.Fatalf("out %q err %v got %s %q", out, err, gotMode, gotText)
	}
	for _, args := range []string{`{"mode":"shout","text":"x"}`, `{"mode":"message","text":" "}`, `nope`} {
		if _, err := (Tell{}).Run(ctx, args); err == nil {
			t.Errorf("%s: expected an error", args)
		}
	}
	if _, err := (Tell{}).Run(context.Background(), `{"mode":"message","text":"x"}`); err == nil {
		t.Error("expected an error without a teller")
	}
}
