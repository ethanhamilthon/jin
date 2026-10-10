package cliproxy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"
)

func download(ctx context.Context, address string, limit int64) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", address, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("cannot download CLIProxyAPI release")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("CLIProxyAPI release download failed")
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil || int64(len(body)) > limit {
		return nil, errors.New("CLIProxyAPI download is invalid or too large")
	}
	return body, nil
}

func verify(data, sums []byte, name string) error {
	digest := sha256.Sum256(data)
	for _, line := range strings.Split(string(sums), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == name {
			if fields[0] != hex.EncodeToString(digest[:]) {
				return errors.New("CLIProxyAPI checksum mismatch")
			}
			return nil
		}
	}
	return errors.New("CLIProxyAPI archive checksum was not listed")
}
