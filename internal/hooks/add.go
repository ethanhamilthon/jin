package hooks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxHookSize = 256 << 10

// Add copies a markdown file from a URL or a local path into the hooks
// folder, or into the project's .jin/hooks when projectDir is set. The name
// comes from the file name unless name is given. It refuses to overwrite.
func Add(ctx context.Context, source, name, projectDir string) (string, error) {
	data, base, err := fetch(ctx, source)
	if err != nil {
		return "", err
	}
	if name == "" {
		name = strings.TrimSuffix(base, ext)
	}
	target, err := Path(name)
	if projectDir != "" {
		target, err = ProjectPath(projectDir, name)
	}
	if err != nil {
		return "", fmt.Errorf("hook name %q: %w; pass --name", name, err)
	}
	if _, err := os.Stat(target); err == nil {
		return "", errors.New(target + " exists; delete it first or pass --name")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", err
	}
	return target, os.WriteFile(target, data, 0o644)
}

func fetch(ctx context.Context, source string) ([]byte, string, error) {
	u, err := url.Parse(source)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		data, err := os.ReadFile(source)
		if err != nil {
			return nil, "", err
		}
		return checked(data, filepath.Base(source))
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("%s: HTTP %d", source, resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxHookSize+1))
	if err != nil {
		return nil, "", err
	}
	return checked(data, path.Base(u.Path))
}

func checked(data []byte, base string) ([]byte, string, error) {
	if len(data) > maxHookSize {
		return nil, "", errors.New("a hook must be under 256 KB")
	}
	if strings.ContainsRune(string(data), 0) {
		return nil, "", errors.New("not a text file")
	}
	return data, base, nil
}
