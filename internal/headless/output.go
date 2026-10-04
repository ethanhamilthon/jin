package headless

import (
	"encoding/json"
	"fmt"
	"io"

	"jin/internal/provider"
	"jin/internal/store"
)

// result is the last thing a run reports.
type result struct {
	Text      string
	Err       string
	SaveErr   string
	SessionID string
	Duration  int64
	Usage     store.Usage
}

// writer shows a run to the caller. Progress and errors always go to stderr,
// so stdout holds only the answer (text) or strict JSONL (json).
type writer interface {
	Session(id, cwd, model, effort string)
	Message(msg provider.Message)
	Progress(line string)
	Result(r result)
}

type textWriter struct{ out, err io.Writer }

func (textWriter) Session(string, string, string, string) {}
func (textWriter) Message(provider.Message)               {}
func (w textWriter) Progress(line string)                 { fmt.Fprintln(w.err, line) }
func (w textWriter) Result(r result) {
	if r.Err != "" {
		fmt.Fprintln(w.err, "jin: "+r.Err)
		return
	}
	fmt.Fprintln(w.out, r.Text)
}

type jsonWriter struct {
	out, err io.Writer
}

func (w jsonWriter) line(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintln(w.err, "jin: cannot encode output:", err)
		return
	}
	fmt.Fprintln(w.out, string(data))
}

func (w jsonWriter) Session(id, cwd, model, effort string) {
	w.line(map[string]string{"type": "session", "id": id, "cwd": cwd, "model": model, "effort": effort})
}

func (w jsonWriter) Message(msg provider.Message) {
	w.line(struct {
		Type    string           `json:"type"`
		Message provider.Message `json:"message"`
	}{"message", msg})
}

func (w jsonWriter) Progress(line string) { fmt.Fprintln(w.err, line) }

func (w jsonWriter) Result(r result) {
	type usage struct {
		Input   int     `json:"input"`
		Output  int     `json:"output"`
		Context int     `json:"context"`
		Cost    float64 `json:"cost"`
	}
	rec := struct {
		Type      string `json:"type"`
		IsError   bool   `json:"is_error"`
		Result    string `json:"result"`
		Error     string `json:"error,omitempty"`
		SaveError string `json:"save_error,omitempty"`
		SessionID string `json:"session_id"`
		Duration  int64  `json:"duration_ms"`
		Usage     usage  `json:"usage"`
	}{"result", r.Err != "", r.Text, r.Err, r.SaveErr, r.SessionID, r.Duration,
		usage{r.Usage.Input, r.Usage.Output, r.Usage.Context, r.Usage.Cost}}
	w.line(rec)
	if r.Err != "" {
		fmt.Fprintln(w.err, "jin: "+r.Err)
	}
}

func newWriter(format string, out, err io.Writer) writer {
	if format == "json" {
		return jsonWriter{out, err}
	}
	return textWriter{out, err}
}
