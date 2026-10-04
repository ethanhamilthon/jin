package headless

import (
	"bytes"
	"testing"

	"jin/internal/core"
	"jin/internal/store"
)

func TestOnlyErrorsThatEndTheTurnFailTheRun(t *testing.T) {
	cases := []struct {
		name string
		done core.Update
		code int
	}{
		{"non-fatal error then answer", core.Update{Kind: core.UpdateDone, Final: true}, exitOK},
		{"error ends the turn", core.Update{Kind: core.UpdateDone}, exitError},
	}
	for _, c := range cases {
		var out, errOut bytes.Buffer
		r := &runState{out: newWriter("text", &out, &errOut)}
		var o outcome
		var usage store.Usage
		r.apply(core.Update{Kind: core.UpdateError, Text: "Auto-compaction failed: boom"}, &o, &usage)
		r.apply(c.done, &o, &usage)
		if code := r.finish(o, result{Text: "answer"}); code != c.code {
			t.Errorf("%s: code %d, stderr %q", c.name, code, errOut.String())
		}
		if c.code == exitOK && out.String() != "answer\n" {
			t.Errorf("%s: stdout %q", c.name, out.String())
		}
	}
}
