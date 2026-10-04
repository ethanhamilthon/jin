package headless

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

// TestMain doubles as the fake live owner: a child of the test binary claims
// the session named in JIN_TEST_OWNS and holds it until its stdin closes.
func TestMain(m *testing.M) {
	if id := os.Getenv("JIN_TEST_OWNS"); id != "" {
		db, err := store.Open()
		if err == nil {
			err = db.SetRunning(id, true)
		}
		if err != nil {
			fmt.Println("error", err)
			os.Exit(1)
		}
		fmt.Println("ready")
		bufio.NewReader(os.Stdin).ReadString('\n')
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func startOwner(t *testing.T, id string) (pid int, release func()) {
	t.Helper()
	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), "JIN_TEST_OWNS="+id)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	line, _ := bufio.NewReader(stdout).ReadString('\n')
	if strings.TrimSpace(line) != "ready" {
		t.Fatalf("owner said %q", line)
	}
	release = func() { stdin.Close(); cmd.Wait() }
	t.Cleanup(release)
	return cmd.Process.Pid, release
}

func TestSessionOfALiveProcessIsRefusedAndLeftUntouched(t *testing.T) {
	var hits int
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { hits++; sse(w, answerChunk) })
	if err := h.db.TouchProvider("s1", h.dir, "m", "", "t", ""); err != nil {
		t.Fatal(err)
	}
	call := provider.ToolCall{ID: "c1", Type: "function"}
	call.Function.Name, call.Function.Arguments = "bash", `{}`
	pending := provider.Message{Role: "assistant", ToolCalls: []provider.ToolCall{call}}
	if err := h.db.AppendMessage("s1", pending); err != nil {
		t.Fatal(err)
	}
	pid, release := startOwner(t, "s1")
	want := fmt.Sprintf("session s1 is in use by process %d", pid)
	for _, args := range [][]string{{"-p", "--session", "s1", "x"}, {"-p", "-c", "x"}} {
		if code := h.run(t, args...); code != exitError || !strings.Contains(h.errOut.String(), want) {
			t.Fatalf("%v: code %d stderr %q", args, code, h.errOut.String())
		}
	}
	if msgs, _ := h.db.LoadMessages("s1"); len(msgs) != 1 || hits != 0 {
		t.Errorf("the busy session was touched: %d messages, %d requests", len(msgs), hits)
	}
	if code := h.run(t, "-p", "--no-session", "--session", "s1", "x"); code != 0 {
		t.Errorf("read-only use must work: %d %q", code, h.errOut.String())
	}
	release()
	if code := h.run(t, "-p", "--session", "s1", "x"); code != 0 {
		t.Fatalf("after the owner is gone: %d %q", code, h.errOut.String())
	}
}
