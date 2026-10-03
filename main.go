package main

import (
	"context"
	"fmt"
	"os"

	"jin/internal/async"
	"jin/internal/cli"
	"jin/internal/export"
	"jin/internal/headless"
	"jin/internal/hooks"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
	"jin/internal/ui"
	"jin/internal/update"
	"jin/internal/upgrade"
)

// version is overridden at release build time with -X main.version=<tag>.
var version = "v0.6.2"

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "jin:", err)
		code = 1
	}
	os.Exit(code)
}

func run(args []string) (int, error) {
	kind := cli.Classify(args)
	switch kind {
	case cli.Help:
		cli.PrintHelp(os.Stdout)
		return 0, nil
	case cli.Version:
		cli.PrintVersion(os.Stdout, version)
		return 0, nil
	case cli.Update:
		return update.Main(context.Background(), args[1:], version, os.Stdout, os.Stderr), nil
	case cli.Unknown:
		cli.PrintUnknown(os.Stderr, args)
		return cli.ExitUsage, nil
	}
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
	switch kind {
	case cli.Hooks:
		return hooks.Main(context.Background(), args[1:], dir, os.Stdout, os.Stderr), nil
	case cli.Export:
		return export.Main(args[1:], db, os.Stdout, os.Stderr), nil
	case cli.Async:
		return async.Main(args[1:], db, version, os.Stdout, os.Stderr), nil
	case cli.Daemon:
		return 0, async.Serve(context.Background(), db, version)
	}
	if err := upgrade.Run(db); err != nil {
		fmt.Fprintln(os.Stderr, "jin: warning: upgrade to 0.4 was not finished:", err)
	}
	if err := db.RecoverInterrupted(); err != nil {
		return 1, err
	}
	if kind == cli.Headless {
		return headless.Main(args, db, dir), nil
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
	client := provider.NewClient(cfg.Provider)
	client.SetStallTimeout(cfg.StallTimeout)
	action, err := ui.Run(ctx, ui.Deps{
		Store: db, Config: cfg, Client: client,
		Registry: registry, Pricing: prices, Dir: dir, Version: version,
	})
	if err != nil || action == nil {
		return 0, err
	}
	return moveData(db, action)
}
