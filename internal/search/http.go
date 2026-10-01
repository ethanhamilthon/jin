package search

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const maxSearchBody = 2 << 20
const userAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"

var client = &http.Client{Timeout: 20 * time.Second}

func fetch(req *http.Request, key string) ([]byte, error) {
	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %s", redact(err.Error(), key))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSearchBody))
	if err != nil {
		return nil, fmt.Errorf("cannot read search response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := strings.TrimSpace(string(body[:min(len(body), 300)]))
		return nil, fmt.Errorf("search HTTP %d: %s", resp.StatusCode, redact(snippet, key))
	}
	return body, nil
}

func redact(text, key string) string {
	if key == "" {
		return text
	}
	return strings.ReplaceAll(text, key, "[redacted]")
}
