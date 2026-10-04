package ui

import "os"

func (a *app) pollAsyncPaths(paths []string) asyncBatch {
	batch := asyncBatch{running: map[string]int{}}
	seen, taskIDs := map[string]bool{}, map[string]bool{}
	for _, path := range paths {
		if seen[path] {
			continue
		}
		seen[path] = true
		events, _ := a.store.ClaimAsyncEvents(path, os.Getpid())
		batch.events = append(batch.events, events...)
		tasks, _ := a.store.RunningAsyncTasks(path)
		for _, task := range tasks {
			if !taskIDs[task.ID] {
				taskIDs[task.ID] = true
				batch.running[task.SessionID]++
			}
		}
	}
	return batch
}
