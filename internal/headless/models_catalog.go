package headless

import (
	"context"
	"fmt"
	"jin/internal/sources"
	"jin/internal/store"
)

func printCatalog(ctx context.Context, db *store.DB, all bool, format string, io_ ioSet) int {
	catalog, err := sources.List(ctx, db, all)
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return 1
	}
	entries := make([]modelEntry, 0, len(catalog.Models))
	for _, model := range catalog.Models {
		entries = append(entries, modelEntry{ID: model.ID, Provider: model.Provider})
	}
	for _, failure := range catalog.Errors {
		fmt.Fprintf(io_.err, "jin: provider %s: %s\n", failure.Provider, failure.Message)
	}
	if err = printModels(io_.out, entries, format); err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return 1
	}
	if len(entries) == 0 && len(catalog.Errors) > 0 {
		return 1
	}
	return 0
}
