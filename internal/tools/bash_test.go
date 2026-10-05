package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestBashRunRejectsZeroTimeout(t *testing.T) {
	_, err := Bash{}.Run(context.Background(), `{"command":"echo hi","timeout":0}`)
	if err == nil {
		t.Fatal("expected an error for timeout: 0")
	}
}

func TestBashRunRejectsNegativeTimeout(t *testing.T) {
	_, err := Bash{}.Run(context.Background(), `{"command":"echo hi","timeout":-5}`)
	if err == nil {
		t.Fatal("expected an error for a negative timeout")
	}
}

func TestBashSummaryRejectsZeroTimeout(t *testing.T) {
	if _, ok := (Bash{}).Summary(`{"command":"echo hi","timeout":0}`); ok {
		t.Fatal("expected Summary to reject timeout: 0 before the tool ever runs")
	}
}

func TestBashSummaryAppendsTimeout(t *testing.T) {
	got, ok := (Bash{}).Summary(`{"command":"echo hi","timeout":5}`)
	if !ok || got != "echo hi (5s)" {
		t.Fatalf("expected explicit timeout appended, got %q ok=%v", got, ok)
	}
	got, ok = (Bash{}).Summary(`{"command":"echo hi"}`)
	if !ok || got != "echo hi (120s)" {
		t.Fatalf("expected default timeout appended, got %q ok=%v", got, ok)
	}
}

func TestBashRunUsesDefaultTimeoutWhenOmitted(t *testing.T) {
	out, err := Bash{}.Run(context.Background(), `{"command":"echo hi"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected command output, got %q", out)
	}
}

func TestBashRunHonorsExplicitTimeout(t *testing.T) {
	out, err := Bash{}.Run(context.Background(), `{"command":"sleep 2","timeout":1}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "[command timed out]") {
		t.Fatalf("expected the 1s timeout to fire before the 2s sleep finished, got %q", out)
	}
}

func TestBashTimeoutKillsChildProcesses(t *testing.T) {
	marker := fmt.Sprintf("jin-test-%d", time.Now().UnixNano())
	command := fmt.Sprintf("exec -a %s sleep 60 & wait", marker)
	start := time.Now()
	result := runBash(context.Background(), command, time.Second, "", true)
	if !strings.Contains(result, "timed out") {
		t.Fatalf("result = %q", result)
	}
	if time.Since(start) > 10*time.Second {
		t.Fatalf("took %v", time.Since(start))
	}
	time.Sleep(200 * time.Millisecond)
	if out, _ := exec.Command("pgrep", "-f", marker).Output(); len(out) > 0 {
		t.Fatalf("child still running: %s", out)
	}
}

func TestBashBackgroundProcessSurvivesUntilKillBackground(t *testing.T) {
	start := time.Now()
	marker := fmt.Sprintf("jin-bg-%d", time.Now().UnixNano())
	args := fmt.Sprintf(`{"command":"(exec -a %s sleep 30 > /dev/null 2>&1 < /dev/null &); echo started"}`, marker)
	out, err := Bash{}.Run(context.Background(), args)
	if err != nil || !strings.Contains(out, "started") {
		t.Fatalf("Run: %q %v", out, err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("bash call blocked on the background process: %v", time.Since(start))
	}
	if exec.Command("pgrep", "-f", marker).Run() != nil {
		t.Fatal("background process died with the bash call")
	}
	KillBackground()
	deadline := time.Now().Add(2 * time.Second)
	for exec.Command("pgrep", "-f", marker).Run() == nil {
		if time.Now().After(deadline) {
			t.Fatal("background process still running after KillBackground")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestBashAsyncLaunchRecipe(t *testing.T) {
	dir := t.TempDir()
	cmd := fmt.Sprintf(`d=%q; ( { sleep 1; echo done > "$d/1.txt"; } < /dev/null & echo $! > "$d/1.pid"; wait $!; echo $? > "$d/1.exit" ) > /dev/null 2>&1 &`, dir)
	args := fmt.Sprintf(`{"command":%q}`, cmd)
	start := time.Now()
	if _, err := (Bash{}).Run(context.Background(), args); err != nil {
		t.Fatal(err)
	}
	if time.Since(start) > 700*time.Millisecond {
		t.Fatalf("launch blocked for %v", time.Since(start))
	}
	deadline := time.Now().Add(4 * time.Second)
	for {
		if b, err := os.ReadFile(dir + "/1.exit"); err == nil && strings.TrimSpace(string(b)) == "0" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("1.exit never appeared")
		}
		time.Sleep(50 * time.Millisecond)
	}
	KillBackground()
}

func TestBashCannotUseTerminal(t *testing.T) {
	start := time.Now()
	out := runBash(context.Background(), "cat /dev/tty; echo status=$?", 20*time.Second, "", true)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("hung on /dev/tty for %v", time.Since(start))
	}
	if strings.Contains(out, "status=0") {
		t.Fatalf("reading /dev/tty succeeded: %q", out)
	}
}

func TestBashSetsNonInteractiveEnvironment(t *testing.T) {
	out := runBash(context.Background(), "echo $GIT_TERMINAL_PROMPT $GIT_EDITOR $GIT_PAGER $PAGER $DEBIAN_FRONTEND", time.Minute, "", true)
	if !strings.Contains(out, "0 true cat cat noninteractive") {
		t.Fatalf("environment = %q", out)
	}
}

func TestBashRunsInOwnProcessGroup(t *testing.T) {
	out := runBash(context.Background(), "ps -o pid=,pgid= -p $$", time.Minute, "", true)
	f := strings.Fields(out)
	if len(f) < 2 || f[0] != f[1] {
		t.Fatalf("pid and pgid differ: %q", out)
	}
}
