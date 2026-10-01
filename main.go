package main

import (
	"context"
	"fmt"
	"os"

	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
	"jin/internal/ui"
)

// version is overridden at release build time with -X main.version=<tag>.
var version = "v0.1"

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "jin:", err)
		os.Exit(1)
	}
}

func run() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	db, err := store.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.RecoverInterrupted(); err != nil {
		return err
	}
	cfg, err := db.LoadConfig()
	if err != nil {
		return err
	}
	registry := tools.NewRegistry(
		tools.NewRead(), tools.NewWrite(), tools.NewEdit(), tools.NewBash(),
	)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	prices := make(chan pricing.Table, 1)
	go func() { prices <- pricing.Load(ctx) }()
	return ui.Run(ctx, ui.Deps{
		Store: db, Config: cfg, Client: provider.NewClient(cfg.Provider),
		Registry: registry, Pricing: prices, Dir: dir, Version: version,
	})
}
