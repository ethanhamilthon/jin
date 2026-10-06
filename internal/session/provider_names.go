package session

import (
	"strconv"
	"strings"

	"jin/internal/provider"
	"jin/internal/store"
)

// DefaultProviderName is a free name for a new provider of a kind.
func DefaultProviderName(kind string, existing []store.ProviderEntry) string {
	base := "openai"
	switch kind {
	case provider.KindAnthropic:
		base = "anthropic"
	case provider.KindResponses:
		base = "responses"
	}
	return uniqueName(base, existing)
}

func uniqueName(base string, existing []store.ProviderEntry) string {
	name := base
	for n := 2; providerNameTaken(name, existing); n++ {
		name = base + "-" + strconv.Itoa(n)
	}
	return name
}

func providerNameTaken(name string, existing []store.ProviderEntry) bool {
	for _, p := range existing {
		if p.Name == name || p.ID == name {
			return true
		}
	}
	return false
}

// NewProviderID turns a name into a short unique id.
func NewProviderID(name string, existing []store.ProviderEntry) string {
	var b strings.Builder
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case b.Len() > 0 && !strings.HasSuffix(b.String(), "-"):
			b.WriteByte('-')
		}
	}
	id := strings.Trim(b.String(), "-")
	if id == "" {
		id = "provider"
	}
	return uniqueName(id, existing)
}
