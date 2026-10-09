package hooks

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"jin/internal/dyn"
)

// Settings are the switches that decide which hooks a session uses.
type Settings struct {
	// Disabled are the names of the hooks that are switched off.
	Disabled []string
	// Trusted lets the project hooks of the folder run.
	Trusted bool
}

// commandTimeout leaves room inside the 10 seconds of the {{jin hooks render}} command
// that runs this, so a slow hook command is reported in place instead of killing
// the whole output.
const commandTimeout = 8 * time.Second

// RenderFilled prints the hooks of a session in dir, filled in and joined by a
// blank line: what the system prompt takes from `{{jin hooks render}}`.
// The commands of the hooks run at the same time.
func RenderFilled(ctx context.Context, dir string, settings Settings, out, errOut io.Writer) int {
	active, warnings := LoadIn(dir, settings.Disabled, settings.Trusted)
	if !settings.Trusted {
		if names, _ := ListProject(dir); len(names) > 0 {
			warnings = append(warnings, "the project hooks of this folder are not trusted, so they were left out")
		}
	}
	dyn.Timeout = commandTimeout
	texts := make([]string, len(active))
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i, hook := range active {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result := dyn.Expand(ctx, hook.Body, dyn.Options{Dir: dir})
			texts[i] = strings.TrimSpace(result.Text)
			mu.Lock()
			warnings = append(warnings, result.Warnings...)
			mu.Unlock()
		}()
	}
	wg.Wait()
	var parts []string
	for _, text := range texts {
		if text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) > 0 {
		fmt.Fprintln(out, strings.Join(parts, "\n\n"))
	}
	for _, warning := range warnings {
		fmt.Fprintln(errOut, "jin hooks:", warning)
	}
	return 0
}
