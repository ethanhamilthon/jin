package headless

import (
	"context"
	"jin/internal/provider"
	"jin/internal/store"
	"sync"
	"time"
)

const (
	probeWorkers = 8
	probeTimeout = 10 * time.Second
	listTimeout  = 30 * time.Second
)

// RefreshModels fetches the model list and caches it; with probe it also
// finds the reasoning levels of every model. A failed probe means "unknown".
func refreshModels(ctx context.Context, db *store.DB, client *provider.Client, probe bool) (int, error) {
	listCtx, cancel := context.WithTimeout(ctx, listTimeout)
	ids, err := client.Models(listCtx)
	cancel()
	if err != nil {
		return 0, err
	}
	if err := db.SaveModelsCache(ids); err != nil {
		return 0, err
	}
	if !probe {
		return len(ids), nil
	}
	levels := map[string][]string{}
	var mu sync.Mutex
	var wg sync.WaitGroup
	jobs := make(chan string)
	for range probeWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				probeCtx, cancel := context.WithTimeout(ctx, probeTimeout)
				found, err := client.Efforts(probeCtx, id)
				cancel()
				if err == nil && len(found) > 0 {
					mu.Lock()
					levels[id] = found
					mu.Unlock()
				}
			}
		}()
	}
	for _, id := range ids {
		jobs <- id
	}
	close(jobs)
	wg.Wait()
	return len(ids), db.SaveModelLevels(levels)
}
