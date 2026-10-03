package provider

import (
	"fmt"
	"strings"
)

func httpError(path string, status int, body []byte, key string) error {
	return fmt.Errorf("%s HTTP %d: %s", path, status, redact(strings.TrimSpace(string(body)), key))
}
