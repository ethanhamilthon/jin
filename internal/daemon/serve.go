package daemon

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"jin/internal/datadir"
	"jin/internal/pricing"
	"jin/internal/session"
	"jin/internal/store"
	"jin/internal/tasks"
	"jin/internal/tools"
)

func Serve(version string) error {
	if err := datadir.Hold(); err != nil {
		return err
	}
	defer datadir.Release()
	root, err := datadir.Current()
	if err != nil {
		return err
	}
	listener, cleanup, err := listen(root)
	if err != nil {
		return err
	}
	defer cleanup()
	db, err := store.Open()
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.RecoverInterrupted(); err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	manager := session.NewManager(ctx, db, version, nil)
	manager.UseSharedQueue()
	defer tasks.Shared().StopAll()
	defer tools.KillBackground()
	defer manager.Shutdown()
	prices := make(chan pricing.Table, 1)
	go func() { prices <- pricing.Load(ctx) }()
	manager.Start(prices, tasks.Shared().Events())
	web := &webService{ctx: ctx, manager: manager, db: db, version: version}
	defer web.close()
	mux := routes(version, manager, cancel)
	mux.HandleFunc("POST /web", web.route)
	mux.HandleFunc("POST /remote", web.remoteRoute)
	go web.restore()
	server := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	ended := make(chan error, 1)
	go func() { ended <- server.Serve(listener) }()
	select {
	case <-ctx.Done():
	case err = <-ended:
		if err != http.ErrServerClosed {
			return err
		}
	}
	shutdown, done := context.WithTimeout(context.Background(), 3*time.Second)
	defer done()
	return server.Shutdown(shutdown)
}
