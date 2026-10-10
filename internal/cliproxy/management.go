package cliproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

func (p *process) management(ctx context.Context, method, path string, value any) (json.RawMessage, error) {
	var body []byte
	var err error
	if value != nil {
		body, err = json.Marshal(value)
		if err != nil {
			return nil, err
		}
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, method, p.endpoint+"/v0/management"+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.keys.Management)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("CLIProxyAPI management is unavailable")
	}
	defer resp.Body.Close()
	if strings.TrimPrefix(resp.Header.Get("X-CPA-VERSION"), "v") != strings.TrimPrefix(p.version, "v") {
		return nil, errors.New("CLIProxyAPI executable version does not match its release")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return nil, errors.New("invalid CLIProxyAPI management response")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, errors.New("CLIProxyAPI management operation failed")
	}
	if !json.Valid(data) {
		return nil, errors.New("invalid CLIProxyAPI management JSON")
	}
	return data, nil
}
