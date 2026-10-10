package main

import (
	"context"
	"fmt"
	"os"

	"jin/internal/cli"
	"jin/internal/cliproxy"
	"jin/internal/datadir"
	"jin/internal/docs"
	"jin/internal/export"
	"jin/internal/headless"
	"jin/internal/hooks"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tasks"
	"jin/internal/tools"
	"jin/internal/ui"
	"jin/internal/update"
	"jin/internal/upgrade"
	"jin/internal/web"
)

// version is overridden at release build time with -X main.version=<tag>.
var version = "v0.14.0"

func main() {
	code, err := run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "jin:", err)
		code = 1
	}
	os.Exit(code)
}

func run(args []string) (int, error) {
	if len(args) == 2 && args[0] == "--internal-cliproxy" {
		if err := datadir.Hold(); err != nil {
			return 1, err
		}
		defer datadir.Release()
		return 0, cliproxy.Serve(args[1])
	}
	defer cliproxy.CloseAll()
	kind := cli.Classify(args)
	switch kind {
	case cli.Help:
		cli.PrintHelp(os.Stdout)
		return 0, nil
	case cli.Version:
		cli.PrintVersion(os.Stdout, version)
		return 0, nil
	case cli.Docs:
		return docs.Main(args[1:], docsFS(), os.Stdout, os.Stderr), nil
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
	if err := datadir.Hold(); err != nil {
		return 1, err
	}
	defer datadir.Release()
	db, err := store.Open()
	if err != nil {
		return 1, err
	}
	defer db.Close()
	defer tools.KillBackground()
	defer tasks.Shared().StopAll()
	switch kind {
	case cli.Sessions:
		return cli.SessionsMain(args[1:], db, dir, os.Stdout, os.Stderr), nil
	case cli.Hooks:
		cfg, _ := db.LoadConfig()
		trust, _ := db.HooksTrust(dir)
		settings := hooks.Settings{Disabled: cfg.HooksDisabled, Trusted: trust == store.Trusted}
		return hooks.Main(context.Background(), args[1:], dir, settings, os.Stdout, os.Stderr), nil
	case cli.Export:
		return export.Main(args[1:], db, os.Stdout, os.Stderr), nil
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
	if kind == cli.Web {
		code, action := web.Main(args[1:], db, dir, version, os.Stdout, os.Stderr)
		if action == nil {
			return code, nil
		}
		return moveData(db, action)
	}
	cfg, err := db.LoadConfig()
	if err != nil {
		return 1, err
	}
	// The full registry only describes old tool calls; every session builds
	// its own registry from the enabled tools.
	registry := tools.Build(tools.Catalog())
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
