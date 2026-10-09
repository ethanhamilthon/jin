package sysprompt

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
)

// LatestBase is where the newest default prompts live: the files that are
// built into jin, on the main branch. Tests point it at a local server.
var LatestBase = "https://raw.githubusercontent.com/ethanhamilthon/jin/main/internal/sysprompt/defaults/"

const (
	fetchTimeout = 15 * time.Second
	fetchLimit   = 256 << 10
)

// Names lists the names that can be fetched and reset.
func Names() []string { return []string{SectionSystem, SectionCompact, SectionHandoff} }

// Latest downloads the newest default text of each named section.
func Latest(ctx context.Context, names ...string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(ctx, fetchTimeout)
	defer cancel()
	out := map[string]string{}
	for _, name := range names {
		if !slices.Contains(Names(), name) {
			return nil, fmt.Errorf("unknown prompt %q", name)
		}
		text, err := fetch(ctx, name)
		if err != nil {
			return nil, err
		}
		out[name] = text
	}
	return out, nil
}

func fetch(ctx context.Context, name string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, LatestBase+name+".md", nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("cannot reach git for the %s prompt: %w", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("git answered %s for the %s prompt", resp.Status, name)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, fetchLimit))
	if err != nil {
		return "", err
	}
	text := strings.TrimSpace(string(data))
	if text == "" {
		return "", fmt.Errorf("the %s prompt on git is empty", name)
	}
	return text, nil
}
