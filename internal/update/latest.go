// Package update finds the latest jin release and installs it in place of
// the running binary.
package update

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Repo is the GitHub repository releases come from.
const Repo = "ethanhamilthon/jin"

// releaseBase is the GitHub site; tests point it at a local server.
var releaseBase = "https://github.com"

// Latest returns the tag of the newest release. It reads the redirect of
// /releases/latest, which needs no API token and has no API rate limit.
func Latest(ctx context.Context) (string, error) {
	client := &http.Client{
		Timeout:       15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, releaseBase+"/"+Repo+"/releases/latest", nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	location := resp.Header.Get("Location")
	_, tag, ok := strings.Cut(location, "/releases/tag/")
	if !ok || tag == "" {
		return "", errors.New("cannot find the latest release")
	}
	return tag, nil
}

// Newer reports whether tag is a later version than current. Versions look
// like v0.6 or v0.6.1; a dev build ("dev" or anything unparsable) is never
// offered an update.
func Newer(tag, current string) bool {
	a, okA := parse(tag)
	b, okB := parse(current)
	if !okA || !okB {
		return false
	}
	for i := range 3 {
		if a[i] != b[i] {
			return a[i] > b[i]
		}
	}
	return false
}

func parse(version string) ([3]int, bool) {
	var out [3]int
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(version), "v"), ".")
	if len(parts) < 2 || len(parts) > 3 {
		return out, false
	}
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return out, false
		}
		out[i] = n
	}
	return out, true
}
