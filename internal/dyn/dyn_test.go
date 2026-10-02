package dyn

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func expand(text string) Result {
	return Expand(context.Background(), text, Options{})
}

func TestCommandIsReplacedByItsOutput(t *testing.T) {
	got := expand("a {{echo one}} b {{printf 'x\\n\\n'}} c")
	if got.Text != "a one b x c" || len(got.Warnings) != 0 {
		t.Errorf("got %+v", got)
	}
}

func TestRunsInTheGivenDirectory(t *testing.T) {
	dir := t.TempDir()
	got := Expand(context.Background(), "{{pwd -P}}", Options{Dir: dir})
	if !strings.HasSuffix(got.Text, strings.TrimPrefix(dir, "/private")) {
		t.Errorf("pwd = %q, want %q", got.Text, dir)
	}
}

func TestEnvironmentIsTheOneGiven(t *testing.T) {
	got := Expand(context.Background(), "{{echo $JIN_X}}", Options{Env: []string{"JIN_X=yes", "PATH=/usr/bin:/bin"}})
	if got.Text != "yes" {
		t.Errorf("got %q", got.Text)
	}
}

func TestEscapedBracesStayLiteral(t *testing.T) {
	got := expand(`json: \{{"a": 1}} and {{echo ok}}`)
	if got.Text != `json: {{"a": 1}} and ok` {
		t.Errorf("got %q", got.Text)
	}
}

func TestEmptyAndUnclosedBracesAreLeftAlone(t *testing.T) {
	for _, text := range []string{"a {{}} b", "a {{  }} b", "open {{ never closed", "plain", "}} only"} {
		if got := expand(text); got.Text != text || len(got.Warnings) != 0 {
			t.Errorf("%q became %+v", text, got)
		}
	}
}

func TestFailingCommandLeavesAMarkerAndAWarning(t *testing.T) {
	got := expand("before {{echo oops >&2; exit 3}} after {{echo fine}}")
	if !strings.Contains(got.Text, "before [command failed: exit status 3: oops] after fine") {
		t.Errorf("text = %q", got.Text)
	}
	if len(got.Warnings) != 1 || !strings.Contains(got.Warnings[0], "exit status 3") {
		t.Errorf("warnings = %v", got.Warnings)
	}
}

func TestCommandThatDoesNotExistIsAFailureNotACrash(t *testing.T) {
	got := expand("{{definitely-not-a-command-xyz}}")
	if !strings.Contains(got.Text, "[command failed:") || len(got.Warnings) != 1 {
		t.Errorf("got %+v", got)
	}
}

func TestOutputHasNoLimit(t *testing.T) {
	got := expand("{{head -c 200000 /dev/zero | tr '\\0' a}}")
	if len(got.Text) != 200000 {
		t.Errorf("len = %d, want the whole output", len(got.Text))
	}
}

func TestTimeoutKillsTheCommandAndItsChildren(t *testing.T) {
	old := Timeout
	Timeout = 300 * time.Millisecond
	defer func() { Timeout = old }()
	marker := "jin-dyn-" + time.Now().Format("150405.000000")
	start := time.Now()
	got := expand("{{exec -a " + marker + " sleep 30 & wait}}")
	if time.Since(start) > 5*time.Second {
		t.Fatalf("took %v", time.Since(start))
	}
	if !strings.Contains(got.Text, "timed out") || len(got.Warnings) != 1 {
		t.Errorf("got %+v", got)
	}
	time.Sleep(200 * time.Millisecond)
	if out, _ := exec.Command("pgrep", "-f", marker).Output(); len(out) > 0 {
		t.Errorf("a child is still running: %s", out)
	}
}

func TestCancelStopsCommandsAndMarksThem(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Result, 1)
	go func() { done <- Expand(ctx, "{{sleep 30}} and {{sleep 30}}", Options{}) }()
	time.Sleep(200 * time.Millisecond)
	cancel()
	select {
	case got := <-done:
		if got.Text != "[command cancelled] and [command cancelled]" {
			t.Errorf("text = %q", got.Text)
		}
		if len(got.Warnings) != 0 {
			t.Errorf("a cancelled command is not a failure, got warnings %v", got.Warnings)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not stop the commands")
	}
}

func TestCommandsOfOneTextRunInParallel(t *testing.T) {
	start := time.Now()
	got := expand("{{sleep 0.5; echo a}}{{sleep 0.5; echo b}}{{sleep 0.5; echo c}}{{sleep 0.5; echo d}}")
	if got.Text != "abcd" {
		t.Errorf("text = %q", got.Text)
	}
	if time.Since(start) > 1500*time.Millisecond {
		t.Errorf("four half-second commands took %v: not parallel", time.Since(start))
	}
}

func TestHas(t *testing.T) {
	for text, want := range map[string]bool{"{{ls}}": true, "plain": false, `\{{ls}}`: false, "{{}}": false, "a {{x": false} {
		if got := Has(text); got != want {
			t.Errorf("Has(%q) = %v, want %v", text, got, want)
		}
	}
}
