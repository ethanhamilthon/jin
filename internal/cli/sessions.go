package cli

import (
	"encoding/json"
	"fmt"
	"io"

	"jin/internal/store"
)

const sessionsUsage = `usage:
  jin sessions list [--all] [--format json]
  jin sessions search <words...> [--all] [--format json]`

// SessionsMain runs `jin sessions ...`.
func SessionsMain(args []string, db *store.DB, dir string, out, errOut io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(errOut, sessionsUsage)
		return 2
	}
	subcmd := args[0]
	opts, err := parseSessionsArgs(args[1:])
	if err != nil {
		fmt.Fprintf(errOut, "jin sessions: %v\n", err)
		return 2
	}
	switch subcmd {
	case "list":
		return runSessionsList(db, dir, opts, out, errOut)
	case "search":
		return runSessionsSearch(db, dir, opts, out, errOut)
	default:
		fmt.Fprintf(errOut, "jin sessions: unknown command %q\n%s\n", subcmd, sessionsUsage)
		return 2
	}
}

func runSessionsList(db *store.DB, dir string, opts sessionsOptions, out, errOut io.Writer) int {
	results, err := db.ListSessions(dir, opts.all)
	if err != nil {
		fmt.Fprintf(errOut, "jin sessions: %v\n", err)
		return 1
	}
	if opts.format == "json" {
		if results == nil {
			results = []store.SessionResult{}
		}
		_ = json.NewEncoder(out).Encode(results)
		return 0
	}
	for _, r := range results {
		fmt.Fprintf(out, "%s\t%s\t%s\n", r.ID, r.Date, r.Title)
	}
	return 0
}

func runSessionsSearch(db *store.DB, dir string, opts sessionsOptions, out, errOut io.Writer) int {
	if len(opts.words) == 0 {
		fmt.Fprintln(errOut, "jin sessions: search requires at least one search word")
		return 2
	}
	results, err := db.SearchSessions(dir, opts.all, opts.words)
	if err != nil {
		fmt.Fprintf(errOut, "jin sessions: %v\n", err)
		return 1
	}
	if opts.format == "json" {
		if results == nil {
			results = []store.SessionResult{}
		}
		_ = json.NewEncoder(out).Encode(results)
		return 0
	}
	for _, r := range results {
		fmt.Fprintf(out, "%s\t%s\t%s\t%s\n", r.ID, r.Date, r.Title, r.Snippet)
	}
	return 0
}
