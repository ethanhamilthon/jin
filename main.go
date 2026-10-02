package main

import (
	"context"
	"fmt"
	"os"

	"jin/internal/headless"
	"jin/internal/pricing"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
	"jin/internal/ui"
)

// version is overridden at release build time with -X main.version=<tag>.
var version = "v0.3"

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "jin:", err)
		code = 1
	}
	os.Exit(code)
}

func run(args []string) (int, error) {
	dir, err := os.Getwd()
	if err != nil {
		return 1, err
	}
	db, err := store.Open()
	if err != nil {
		return 1, err
	}
	defer db.Close()
	defer tools.KillBackground()
	if err := db.RecoverInterrupted(); err != nil {
		return 1, err
	}
	if headless.Handles(args) {
		return headless.Main(args, db, dir), nil
	}
	if err := prompts.EnsureDefaults(); err != nil {
		fmt.Fprintln(os.Stderr, "jin: warning: default prompts were not created:", err)
	}
	cfg, err := db.LoadConfig()
	if err != nil {
		return 1, err
	}
	// The full registry only describes old tool calls; every session builds
	// its own registry from the enabled tools.
	registry := tools.Build(tools.Catalog(), &tools.MemoryTodos{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	prices := make(chan pricing.Table, 1)
	go func() { prices <- pricing.Load(ctx) }()
	err = ui.Run(ctx, ui.Deps{
		Store: db, Config: cfg, Client: provider.NewClient(cfg.Provider),
		Registry: registry, Pricing: prices, Dir: dir, Version: version,
	})
	return 0, err
}
