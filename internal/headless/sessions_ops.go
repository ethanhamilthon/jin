package headless

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"jin/internal/session"
)

// handoffBrief asks the model for the brief of a fresh session and prints it.
// A one-shot process has no draft to put it in; feed it to `jin -p`.
func (s *managed) handoffBrief(ctx context.Context, io_ ioSet, p sessionArgs) error {
	before := s.entries()
	if err := s.m.Handoff(s.id); err != nil {
		return err
	}
	if err := s.idle(ctx); err != nil {
		return err
	}
	select {
	case next := <-s.handoff:
		snap, err := s.m.Snapshot(next)
		if err != nil {
			return err
		}
		if p.format == "json" {
			writeJSON(io_.out, map[string]string{"session": s.id, "brief": snap.State.Draft})
		} else {
			fmt.Fprintln(io_.out, snap.State.Draft)
		}
		return nil
	default:
		return s.failure(before)
	}
}

func (s *managed) failure(before []session.Entry) error {
	for _, e := range s.entries()[len(before):] {
		if e.Text != "" {
			return errors.New(e.Text)
		}
	}
	return errors.New("the model returned no brief")
}

// rewind lists the messages you typed, or with --to forks the session before
// message n and prints the id of the new session.
func (s *managed) rewind(io_ ioSet, p sessionArgs) error {
	if p.to == 0 {
		points, err := s.m.Points(s.id)
		if err != nil {
			return err
		}
		if p.format == "json" {
			writeJSON(io_.out, points)
			return nil
		}
		for i, point := range points {
			line, _, _ := strings.Cut(strings.TrimSpace(point.Text), "\n")
			fmt.Fprintf(io_.out, "%d\t%s\n", i+1, session.Cut(line, 100))
		}
		return nil
	}
	snap, err := s.m.Fork(s.id, p.to-1)
	if err != nil {
		return err
	}
	if p.format == "json" {
		writeJSON(io_.out, map[string]string{"session": snap.State.ID, "message": snap.State.Draft})
		return nil
	}
	fmt.Fprintln(io_.out, snap.State.ID)
	fmt.Fprintln(io_.err, "message: "+snap.State.Draft)
	return nil
}

func (s *managed) context(io_ ioSet, p sessionArgs) error {
	report, err := s.m.Context(s.id)
	if err != nil {
		return err
	}
	if p.format == "json" {
		writeJSON(io_.out, report)
		return nil
	}
	fmt.Fprintf(io_.out, "context: %d of %d tokens used\n", report.Used, report.Window)
	for _, part := range report.Prompt {
		fmt.Fprintf(io_.out, "prompt %s: ~%d\n", part.Name, part.Tokens)
	}
	fmt.Fprintf(io_.out, "tool schemas: ~%d\n", report.ToolSchemas)
	fmt.Fprintf(io_.out, "conversation: ~%d tokens in %d messages\n", report.Conversation, report.Messages)
	for _, part := range report.Results {
		fmt.Fprintf(io_.out, "large result %s: ~%d\n", part.Name, part.Tokens)
	}
	if report.Cache != nil {
		fmt.Fprintf(io_.out, "cache: %d%%\n", *report.Cache)
	}
	if p.full {
		for _, part := range report.Prompt {
			fmt.Fprintf(io_.out, "\n--- %s ---\n%s\n", part.Name, part.Text)
		}
		for _, tool := range report.Tools {
			fmt.Fprintf(io_.out, "\n--- tool schema: %s ---\n%s\n", tool.Name, tool.Text)
		}
	}
	return nil
}
