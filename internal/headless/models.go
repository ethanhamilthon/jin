package headless

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
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
	providerID := fs.String("provider", "", "")
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
	if err == nil && *providerID != "" && command != "models" {
		err = errors.New("--provider belongs to models")
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
	live := *providerID != "" && *providerID != cfg.ActiveProvider
	if *providerID != "" {
		cfg.ActiveProvider, err = useSaved(&cfg, *providerID)
	}
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
	entries, err := listModels(ctx, db, client, scope, *all, live)
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
