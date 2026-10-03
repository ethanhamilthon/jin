package headless

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"jin/internal/provider"
	"jin/internal/store"
)

// runModels handles `jin models` and `jin refresh-models`.
func runModels(ctx context.Context, command string, args []string, db *store.DB, io_ ioSet) int {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	format := fs.String("format", "text", "")
	efforts := fs.Bool("efforts", false, "")
	all := fs.Bool("all", false, "")
	words, err := parseInterleaved(fs, args)
	if err == nil && len(words) > 0 {
		err = fmt.Errorf("unexpected argument %q", words[0])
	}
	if err == nil && *format != "text" && *format != "json" {
		err = fmt.Errorf("unknown format %q (use text or json)", *format)
	}
	if err == nil && *efforts && command != "refresh-models" {
		err = errors.New("--efforts belongs to refresh-models")
	}
	if err == nil && *all && command != "models" {
		err = errors.New("--all belongs to models")
	}
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return 1
	}
	cfg, err := db.LoadConfig()
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return 1
	}
	applyEnv(&cfg, io_.getenv)
	if !cfg.Provider.Ready() {
		fmt.Fprintln(io_.err, "jin: provider is not configured: set JIN_BASE_URL and JIN_API_KEY, or configure it in the TUI")
		return 1
	}
	client := provider.NewClient(cfg.Provider)
	if command == "refresh-models" {
		n, err := refreshModels(ctx, db, client, *efforts)
		if err != nil {
			fmt.Fprintln(io_.err, "jin:", err)
			return 1
		}
		fmt.Fprintf(io_.err, "%d models\n", n)
		return 0
	}
	scope, err := db.LoadScopeFor(cfg.ActiveProvider)
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return 1
	}
	entries, err := listModels(ctx, db, client, scope, *all)
	if err == nil {
		slices.SortFunc(entries, func(a, b modelEntry) int { return strings.Compare(a.ID, b.ID) })
		err = printModels(io_.out, entries, *format)
	}
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return 1
	}
	return 0
}

func formatPrice(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	return s
}

func printModels(w io.Writer, entries []modelEntry, format string) error {
	if format == "json" {
		if entries == nil {
			entries = []modelEntry{}
		}
		return json.NewEncoder(w).Encode(entries)
	}
	for _, e := range entries {
		in, out, ctxWin := "", "", ""
		if e.InputPerMTok != nil {
			in = formatPrice(*e.InputPerMTok)
		}
		if e.OutputPerMTok != nil {
			out = formatPrice(*e.OutputPerMTok)
		}
		if e.ContextWindow != nil {
			ctxWin = strconv.Itoa(*e.ContextWindow)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", e.ID, in, out, ctxWin, strings.Join(e.Efforts, ","))
	}
	return nil
}
