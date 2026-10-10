package sources

import (
	"context"
	"fmt"
	"jin/internal/store"
	"sort"
	"sync"
	"time"
)

type Model struct {
	Provider     string `json:"provider"`
	ProviderName string `json:"provider_name"`
	ID           string `json:"id"`
}
type Failure struct {
	Provider string `json:"provider"`
	Message  string `json:"message"`
}
type Catalog struct {
	Models []Model   `json:"models"`
	Errors []Failure `json:"errors"`
}

func Key(provider, model string) string {
	return fmt.Sprintf("%d:%s%s", len(provider), provider, model)
}

func List(ctx context.Context, db *store.DB, all bool) (Catalog, error) {
	providers, _, err := db.LoadProviders()
	if err != nil {
		return Catalog{}, err
	}
	result := Catalog{Models: []Model{}, Errors: []Failure{}}
	var mu sync.Mutex
	var group sync.WaitGroup
	slots := make(chan struct{}, 4)
	for _, entry := range providers {
		if entry.Disabled {
			continue
		}
		group.Add(1)
		go func(p store.ProviderEntry) {
			defer group.Done()
			select {
			case slots <- struct{}{}:
				defer func() { <-slots }()
			case <-ctx.Done():
				return
			}
			probe, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			models, err := Client(db, p).Models(probe)
			scope, scopeErr := db.LoadScopeFor(p.ID)
			if err == nil {
				err = scopeErr
			}
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				result.Errors = append(result.Errors, Failure{Provider: p.ID, Message: err.Error()})
				return
			}
			for _, id := range models {
				if !all && len(scope) > 0 && !contains(scope, id) {
					continue
				}
				result.Models = append(result.Models, Model{Provider: p.ID, ProviderName: p.Name, ID: id})
			}
		}(entry)
	}
	group.Wait()
	sort.Slice(result.Models, func(i, j int) bool {
		a, b := result.Models[i], result.Models[j]
		if a.ProviderName == b.ProviderName {
			if a.Provider == b.Provider {
				return a.ID < b.ID
			}
			return a.Provider < b.Provider
		}
		return a.ProviderName < b.ProviderName
	})
	return result, nil
}
func contains(list []string, item string) bool {
	for _, s := range list {
		if s == item {
			return true
		}
	}
	return false
}
