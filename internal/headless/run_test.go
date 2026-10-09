package headless

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"jin/internal/pricing"
	"jin/internal/store"
)

func sse(w http.ResponseWriter, chunks ...string) {
	w.Header().Set("Content-Type", "text/event-stream")
	for _, c := range chunks {
		_, _ = w.Write([]byte("data: " + c + "\n\n"))
	}
	_, _ = w.Write([]byte("data: [DONE]\n\n"))
}

const answerChunk = `{"choices":[{"delta":{"role":"assistant","content":"hello there"}}],"usage":{"prompt_tokens":10,"completion_tokens":3}}`

type harness struct {
	db      *store.DB
	out     bytes.Buffer
	errOut  bytes.Buffer
	signals chan os.Signal
	env     map[string]string
	// dir is the working directory of a run; "/work" does not exist, which
	// is fine until a test runs a command in it.
	dir string
}

func newHarness(t *testing.T, handler http.HandlerFunc) *harness {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("JIN_DEPTH", "")
	loadPricing = func(context.Context) pricing.Table {
		return pricing.Table{"m": {InputCostPerToken: 1, OutputCostPerToken: 2}}
	}
	t.Cleanup(func() { loadPricing = pricing.Load })
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &harness{db: db, signals: make(chan os.Signal, 1), dir: "/work",
		env: map[string]string{"JIN_BASE_URL": server.URL, "JIN_API_KEY": "k", "JIN_MODEL": "m"}}
}

func (h *harness) run(t *testing.T, args ...string) int {
	t.Helper()
	h.out.Reset()
	h.errOut.Reset()
	return Run(t.Context(), args, h.db, h.dir, h.signals, ioSet{
		in: strings.NewReader(""), out: &h.out, err: &h.errOut,
		getenv: func(k string) string { return h.env[k] },
	})
}

func TestPlainAnswerTextAndSession(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	if code := h.run(t, "-p", "say", "hi"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	if h.out.String() != "hello there\n" {
		t.Fatalf("stdout = %q", h.out.String())
	}
	list, _ := h.db.ListByPath("/work")
	if len(list) != 1 || list[0].Title != "say hi" || list[0].Usage.Input != 10 || list[0].Usage.Cost != 16 {
		t.Fatalf("sessions = %+v", list)
	}
	msgs, _ := h.db.LoadMessages(list[0].ID)
	if len(msgs) != 2 {
		t.Fatalf("messages = %+v", msgs)
	}
	if code := h.run(t, "-p", "-c", "again"); code != 0 {
		t.Fatalf("continue: code %d, stderr %q", code, h.errOut.String())
	}
	list, _ = h.db.ListByPath("/work")
	msgs, _ = h.db.LoadMessages(list[0].ID)
	if len(list) != 1 || len(msgs) != 4 {
		t.Fatalf("after continue: %d sessions, %d messages", len(list), len(msgs))
	}
}

func TestNoSessionLeavesNoTrace(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	if code := h.run(t, "-p", "--no-session", "hi"); code != 0 {
		t.Fatalf("code %d: %s", code, h.errOut.String())
	}
	if list, _ := h.db.ListByPath("/work"); len(list) != 0 {
		t.Fatalf("sessions = %+v", list)
	}
}

func TestJSONOutputIsStrictJSONL(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	if code := h.run(t, "-p", "--format", "json", "hi"); code != 0 {
		t.Fatalf("code %d: %s", code, h.errOut.String())
	}
	var types []string
	var last map[string]any
	for _, line := range strings.Split(strings.TrimSpace(h.out.String()), "\n") {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("non-JSON line %q", line)
		}
		types = append(types, rec["type"].(string))
		last = rec
	}
	if strings.Join(types, ",") != "session,message,message,result" {
		t.Fatalf("types = %v", types)
	}
	if last["result"] != "hello there" || last["is_error"] != false || last["usage"].(map[string]any)["cost"] != float64(16) {
		t.Fatalf("result = %v", last)
	}
}

func TestToolCallGoesToStderr(t *testing.T) {
	calls := 0
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			sse(w, `{"choices":[{"delta":{"role":"assistant","tool_calls":[{"index":0,"id":"c1","type":"function","function":{"name":"bash","arguments":"{\"command\":\"echo tool-ran\"}"}}]}}]}`)
			return
		}
		sse(w, answerChunk)
	})
	if code := h.run(t, "-p", "go"); code != 0 {
		t.Fatalf("code %d: %s", code, h.errOut.String())
	}
	if !strings.Contains(h.errOut.String(), "bash: echo tool-ran") || h.out.String() != "hello there\n" {
		t.Fatalf("stdout %q stderr %q", h.out.String(), h.errOut.String())
	}
}

func TestAskUserIsNotOffered(t *testing.T) {
	var body string
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body = buf.String()
		sse(w, answerChunk)
	})
	h.run(t, "-p", "hi")
	if strings.Contains(body, `"ask_user"`) || !strings.Contains(body, `"todo"`) {
		t.Fatalf("tools in request: %s", body)
	}
	h.run(t, "-p", "--no-tools", "hi")
	if strings.Contains(body, `"tools"`) {
		t.Fatalf("tools field sent: %s", body)
	}
}

func TestProviderErrorExitsOne(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { http.Error(w, `{"error":{"message":"boom"}}`, 500) })
	if code := h.run(t, "-p", "--format", "json", "hi"); code != 1 {
		t.Fatalf("code %d", code)
	}
	lines := strings.Split(strings.TrimSpace(h.out.String()), "\n")
	var last map[string]any
	_ = json.Unmarshal([]byte(lines[len(lines)-1]), &last)
	if last["type"] != "result" || last["is_error"] != true || last["error"] == "" {
		t.Fatalf("last line %q", lines[len(lines)-1])
	}
}

func TestTimeoutAndInterrupt(t *testing.T) {
	slow := func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(io.Discard, r.Body)
		<-r.Context().Done()
	}
	h := newHarness(t, slow)
	if code := h.run(t, "-p", "--timeout", "200ms", "hi"); code != 1 || !strings.Contains(h.errOut.String(), "timed out after 200ms") {
		t.Fatalf("timeout: code %d stderr %q", code, h.errOut.String())
	}
	go func() { time.Sleep(200 * time.Millisecond); h.signals <- os.Interrupt }()
	if code := h.run(t, "-p", "hi"); code != 130 {
		t.Fatalf("interrupt: code %d stderr %q", code, h.errOut.String())
	}
}

func TestSetupErrors(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	delete(h.env, "JIN_MODEL")
	if code := h.run(t, "-p", "hi"); code != 1 || !strings.Contains(h.errOut.String(), "no model") {
		t.Fatalf("no model: %d %q", code, h.errOut.String())
	}
	h.env["JIN_MODEL"] = "m"
	if code := h.run(t, "-p", "-c", "hi"); code != 1 {
		t.Fatalf("continue without session: %d", code)
	}
	if code := h.run(t, "-p", "--session", "nope", "hi"); code != 1 {
		t.Fatalf("unknown session: %d", code)
	}
	h.env["JIN_DEPTH"] = "3"
	if code := h.run(t, "-p", "hi"); code != 1 || !strings.Contains(h.errOut.String(), "depth limit") {
		t.Fatalf("depth: %d %q", code, h.errOut.String())
	}
}

func TestDBStaysFreeOfEnvValues(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	h.run(t, "-p", "hi")
	cfg, _ := h.db.LoadConfig()
	if cfg.Provider.APIKey != "" || cfg.Provider.BaseURL != "" || cfg.Model != "" {
		t.Fatalf("env leaked into settings: %+v", cfg)
	}
}

// system returns the system prompt that the last request carried.
func systemPromptOf(t *testing.T, body string) string {
	t.Helper()
	var req struct {
		Messages []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal([]byte(body), &req); err != nil {
		t.Fatalf("bad request body: %v\n%s", err, body)
	}
	if len(req.Messages) == 0 || req.Messages[0].Role != "system" {
		t.Fatalf("no system message in %s", body)
	}
	return req.Messages[0].Content
}

func captureBody(h **harness, body *string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		*body = buf.String()
		sse(w, answerChunk)
	}
}

func TestHeadlessSystemPromptHasLiveDataAndAsyncBlock(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	h.dir = t.TempDir()
	if code := h.run(t, "-p", "hi"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	system := systemPromptOf(t, body)
	if strings.Contains(system, "{{") {
		t.Errorf("a placeholder was left:\n%s", system)
	}
	if !strings.Contains(system, "Date: 20") || !strings.Contains(system, "Jin documentation:") {
		t.Errorf("system prompt:\n%s", system)
	}
	list, _ := h.db.ListByPath(h.dir)
	if !strings.Contains(system, "Your session id: "+list[0].ID) {
		t.Errorf("a saved session must tell its id to the agent:\n%s", system)
	}
}

func TestHeadlessWithoutASessionHasNoAsyncBlock(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	if code := h.run(t, "-p", "--no-session", "hi"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	if system := systemPromptOf(t, body); strings.Contains(system, "jin async run") || strings.Contains(system, "Your session id") {
		t.Errorf("no session id, so no async block:\n%s", system)
	}
}

func TestHeadlessWithoutBashHasNoAsyncBlock(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	if code := h.run(t, "-p", "--exclude-tools", "bash", "hi"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	if system := systemPromptOf(t, body); strings.Contains(system, "jin async run") {
		t.Errorf("no bash tool, so no async block:\n%s", system)
	}
}

func TestHeadlessRunsCommandsOfTheSystemPromptFileAndReportsFailures(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	h.dir = t.TempDir()
	root := os.Getenv("HOME") + "/.jin-dev"
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	custom := "# system\n\nI am {{echo custom}}. Broken: {{exit 2}}\n"
	if err := os.WriteFile(root+"/system-prompt.md", []byte(custom), 0o600); err != nil {
		t.Fatal(err)
	}
	if code := h.run(t, "-p", "--no-session", "hi"); code != 0 {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	system := systemPromptOf(t, body)
	if !strings.Contains(system, "I am custom. Broken: [command failed: exit status 2]") {
		t.Errorf("system prompt:\n%s", system)
	}
	if !strings.Contains(h.errOut.String(), "exit status 2") {
		t.Errorf("the failure must be reported on stderr, got %q", h.errOut.String())
	}
}
