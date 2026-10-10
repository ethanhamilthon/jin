package headless

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"jin/internal/store"
)

// runProjectsList prints the active projects, and the archived ones with --all.
func runProjectsList(args []string, db *store.DB, out io.Writer) error {
	fs := flag.NewFlagSet("projects list", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	all := fs.Bool("all", false, "")
	format := fs.String("format", "text", "")
	words, err := parseInterleaved(fs, args)
	if err == nil && len(words) > 0 {
		err = fmt.Errorf("unexpected argument %q", words[0])
	}
	if err == nil && *format != "text" && *format != "json" {
		err = fmt.Errorf("unknown format %q (use text or json)", *format)
	}
	if err != nil {
		return err
	}
	projects, err := db.Projects()
	if err != nil {
		return err
	}
	shown := make([]store.Project, 0, len(projects))
	for _, p := range projects {
		if *all || !p.Archived {
			shown = append(shown, p)
		}
	}
	if *format == "json" {
		return json.NewEncoder(out).Encode(shown)
	}
	for _, p := range shown {
		printProjectRow(out, p)
	}
	return nil
}

// printProjectRow writes one project as tab separated path, name and state.
func printProjectRow(out io.Writer, p store.Project) {
	state := "active"
	if p.Archived {
		state = "archived"
	}
	fmt.Fprintf(out, "%s\t%s\t%s\n", p.Path, p.Name, state)
}
