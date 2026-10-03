package pricing

import (
	"context"
	"io"
	"net/http"
	"os"
)

func readCache(file string, parse func([]byte) (Table, error)) (Table, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	return parse(data)
}

func fetch(ctx context.Context, url string, parse func([]byte) (Table, error)) (Table, []byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, nil, err
	}
	table, err := parse(data)
	if err != nil {
		return nil, nil, err
	}
	return table, data, nil
}
