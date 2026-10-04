package headless

import (
	"strings"
	"sync/atomic"
	"testing"
)

func TestMaxCostNeedsAKnownPrice(t *testing.T) {
	var hits atomic.Int32
	h := newHarness(t, loopingServer(&hits, "true", 1))
	h.env["JIN_MODEL"] = "no-price"
	if code := h.run(t, "-p", "--max-cost", "1", "go"); code != exitError || !strings.Contains(h.errOut.String(), `model "no-price" is not in the price catalogue`) {
		t.Fatalf("code %d stderr %q", code, h.errOut.String())
	}
	if hits.Load() != 0 {
		t.Errorf("a request was sent")
	}
}

func TestBudgetFlagsAreValidated(t *testing.T) {
	for _, args := range [][]string{
		{"--max-cost", "0"}, {"--max-cost", "-1"}, {"--max-cost", "x"}, {"--max-cost", "NaN"},
		{"--max-turns", "0"}, {"--max-turns", "1.5"}, {"--max-turns", "x"},
	} {
		if _, err := ParseArgs(append([]string{"-p"}, append(args, "hi")...)); err == nil {
			t.Errorf("%v accepted", args)
		}
	}
	opt, err := ParseArgs([]string{"-p", "--max-cost", "0.5", "--max-turns", "3", "hi"})
	if err != nil || opt.MaxCost != 0.5 || opt.MaxTurns != 3 {
		t.Errorf("opt %+v err %v", opt, err)
	}
}
