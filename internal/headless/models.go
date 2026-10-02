package headless

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"
	"strings"
	"sync"
	"time"

	"jin/internal/provider"
	"jin/internal/store"
)

const (
	probeWorkers = 8
	probeTimeout = 10 * time.Second
	listTimeout  = 30 * time.Second
)

// RefreshModels fetches the model list and caches it; with probe it also
// finds the reasoning levels of every model. A failed probe means "unknown".
func refreshModels(ctx context.Context, db *store.DB, client *provider.Client, probe bool) (int, error) {
	listCtx, cancel := context.WithTimeout(ctx, listTimeout)
	ids, err := client.Models(listCtx)
	cancel()
	if err != nil {
		return 0, err
	}
	if err := db.SaveModelsCache(ids); err != nil {
		return 0, err
	}
	if !probe {
		return len(ids), nil
	}
	levels := map[string][]string{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	jobs := make(chan string)
	for range probeWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
				found, err := client.Efforts(probeCtx, id)
				cancel()
				if err == nil && len(found) > 0 {
					mu.Lock()
					levels[id] = found
					mu.Unlock()
				}
			}
		}()
	}
	for _, id := range ids {
		jobs <- id
	}
	close(jobs)
	wg.Wait()
	return len(ids), db.SaveModelLevels(levels)
}

type modelEntry struct {
	ID      string   `json:"id"`
	Efforts []string `json:"efforts,omitempty"`
}

// listModels reads the cache, fetching and caching once when it is empty.
func listModels(ctx context.Context, db *store.DB, client *provider.Client) ([]modelEntry, error) {
	ids, err := db.LoadModelsCache()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		if _, err := refreshModels(ctx, db, client, false); err != nil {
			return nil, err
		}
		if ids, err = db.LoadModelsCache(); err != nil {
			return nil, err
		}
	}
	levels, err := db.LoadModelLevels()
	if err != nil {
		return nil, err
	}
	out := make([]modelEntry, len(ids))
	for i, id := range ids {
		out[i] = modelEntry{ID: id, Efforts: levels[id]}
	}
	return out, nil
}

func printModels(w io.Writer, entries []modelEntry, format string) error {
	if format == "json" {
		if entries == nil {
			entries = []modelEntry{}
		}
		return json.NewEncoder(w).Encode(entries)
	}
	for _, e := range entries {
		line := e.ID
		if len(e.Efforts) > 0 {
			line += "\t" + strings.Join(e.Efforts, ",")
		}
		fmt.Fprintln(w, line)
	}
	return nil
}

// runModels handles `jin models` and `jin refresh-models`.
func runModels(ctx context.Context, command string, args []string, db *store.DB, io_ ioSet) int {
	fs := flag.NewFlagSet(command, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	format := fs.String("format", "text", "")
	efforts := fs.Bool("efforts", false, "")
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
	entries, err := listModels(ctx, db, client)
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
