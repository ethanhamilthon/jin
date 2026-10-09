// Package docs serves the documentation of jin from the pages built into the
// binary, so it works offline and always matches the installed version.
package docs

import (
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strings"
)

const usage = `usage:
  jin docs                 print the pointer to the documentation (for the system prompt)
  jin docs --list          list the pages
  jin docs <page>          print one page, for example: jin docs headless`

// Pointer is what the system prompt tells the model about the documentation.
const Pointer = "Jin documentation:\n" +
	"When the user asks about jin itself (TUI, keys, settings, #prompts, hooks, AGENTS.md, sessions, compact, handoff, database, jin -p, extending): run `jin docs README` for the index, then `jin docs <page>` for only the matching pages, and answer with exact names. The pages match the installed version. If they do not cover it, say so.\n" +
	"For user data (sessions, usage, settings), read the local database with `sqlite3`; never print the API key. Do not read the documentation for questions not about jin."

// Main runs `jin docs`. pages holds the .md files at its root.
func Main(args []string, pages fs.FS, out, errOut io.Writer) int {
	names := Names(pages)
	switch {
	case len(args) == 0:
		fmt.Fprintln(out, Pointer)
	case len(args) > 1 || strings.HasPrefix(args[0], "-") && args[0] != "--list":
		fmt.Fprintln(errOut, usage)
		return 2
	case args[0] == "--list":
		fmt.Fprintln(out, strings.Join(names, "\n"))
	default:
		data, err := fs.ReadFile(pages, strings.TrimSuffix(args[0], ".md")+".md")
		if err != nil {
			fmt.Fprintf(errOut, "jin docs: no page %q; the pages are: %s\n", args[0], strings.Join(names, ", "))
			return 1
		}
		fmt.Fprint(out, string(data))
	}
	return 0
}

// Names lists the pages without the .md, sorted.
func Names(pages fs.FS) []string {
	entries, _ := fs.ReadDir(pages, ".")
	var names []string
	for _, entry := range entries {
		if name, ok := strings.CutSuffix(entry.Name(), ".md"); ok && !entry.IsDir() {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
