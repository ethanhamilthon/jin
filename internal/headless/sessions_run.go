package headless

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"jin/internal/core"
	"jin/internal/session"
	"jin/internal/store"
)

// runSessionAction executes `jin sessions <action> <id>`: one operation on a
// saved session, then exit.
func runSessionAction(ctx context.Context, args []string, db *store.DB, io_ ioSet) int {
	parsed, err := parseSessionArgs(args)
	if err != nil {
		fmt.Fprintln(io_.err, "jin sessions:", err)
		return exitError
	}
	s, closeAll, err := openManaged(ctx, db, parsed.id)
	if err != nil {
		fmt.Fprintln(io_.err, "jin sessions:", err)
		return exitError
	}
	defer closeAll()
	if err := s.run(ctx, parsed, io_); err != nil {
		fmt.Fprintln(io_.err, "jin sessions:", err)
		return exitError
	}
	return exitOK
}

func (s *managed) run(ctx context.Context, p sessionArgs, io_ ioSet) error {
	switch p.action {
	case "compact":
		return s.report(ctx, io_, p, s.m.Compact)
	case "undo":
		return s.report(ctx, io_, p, s.m.Undo)
	case "reload":
		return s.report(ctx, io_, p, s.m.Reload)
	case "handoff":
		return s.handoffBrief(ctx, io_, p)
	case "rewind":
		return s.rewind(io_, p)
	}
	return s.context(io_, p)
}

// report runs an operation that writes entries to the session, waits for the
// session to settle and prints what was added.
func (s *managed) report(ctx context.Context, io_ ioSet, p sessionArgs, do func(string) error) error {
	before := s.entries()
	if err := do(s.id); err != nil {
		return err
	}
	if err := s.idle(ctx); err != nil {
		return err
	}
	added := s.entries()[len(before):]
	return s.print(io_, p, added)
}

func (s *managed) entries() []session.Entry {
	snap, _ := s.m.Snapshot(s.id)
	return snap.Entries
}

type entryJSON struct {
	Kind string `json:"kind"`
	Text string `json:"text"`
}

func (s *managed) print(io_ ioSet, p sessionArgs, added []session.Entry) error {
	var failure error
	var list []entryJSON
	for _, e := range added {
		if e.Text == "" {
			continue
		}
		if e.Kind == core.UpdateError && failure == nil {
			failure = fmt.Errorf("%s", e.Text)
		}
		list = append(list, entryJSON{Kind: fmt.Sprint(e.Kind), Text: e.Text})
		if p.format == "text" && e.Kind != core.UpdateError {
			fmt.Fprintln(io_.out, e.Text)
		}
	}
	if p.format == "json" {
		writeJSON(io_.out, map[string]any{"session": s.id, "entries": list})
	}
	return failure
}

func writeJSON(w io.Writer, v any) {
	data, _ := json.Marshal(v)
	fmt.Fprintln(w, string(data))
}
