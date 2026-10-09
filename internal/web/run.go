package web

import (
	"context"
	"fmt"
	"io"
	"net"
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
	jinupdate "jin/internal/update"
)

// Main runs `jin web` until the user stops it. A returned action must be
// carried out after the database is closed.
func Main(args []string, db *store.DB, dir, version string, out, errOut io.Writer) (int, *datadir.Action) {
	opt, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintf(errOut, "jin web: %v\n%s", err, Usage)
		return 2, nil
	}
	if opt.Help {
		fmt.Fprint(out, Usage)
		return 0, nil
	}
	if opt.Cwd != "" {
		dir = opt.Cwd
	}
	if dir, err = session.ProjectPath(dir, dir); err != nil {
		fmt.Fprintln(errOut, "jin web:", err)
		return 1, nil
	}
	if _, err := db.EnsureProject(dir); err != nil {
		fmt.Fprintln(errOut, "jin web:", err)
		return 1, nil
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if opt.Remote {
		name, err := tailnetName(ctx)
		if err != nil {
			fmt.Fprintln(errOut, "jin web:", err)
			return 1, nil
		}
		opt.Hosts = append(opt.Hosts, name)
	}
	listener, err := listen(opt.Port)
	if err != nil {
		fmt.Fprintln(errOut, "jin web:", err)
		return 1, nil
	}
	if _, ok := assets(); !ok {
		fmt.Fprintln(errOut, "jin web:", noUI)
	}
	if opt.Remote {
		_, port, _ := net.SplitHostPort(listener.Addr().String())
		if err := tailnetServe(ctx, port); err != nil {
			listener.Close()
			fmt.Fprintln(errOut, "jin web:", err)
			return 1, nil
		}
		defer tailnetStop()
		keepAwake(out)
	}
	return serve(ctx, listener, db, dir, version, opt, out)
}

func serve(ctx context.Context, listener net.Listener, db *store.DB, dir, version string, opt Options, out io.Writer) (int, *datadir.Action) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	s := &server{ctx: ctx, db: db, hub: newHub(), dir: dir, version: version, quit: cancel}
	s.m = session.NewManager(ctx, db, version, func(ev session.Event) { s.publish(ev) })
	prices := make(chan pricing.Table, 1)
	go func() { prices <- pricing.Load(ctx) }()
	s.m.Start(prices, tasks.Shared().Events())
	go s.checkUpdate()
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	g := newGuard(port, opt.Hosts, db)
	s.guard = g
	if opt.Remote {
		s.remoteHost = opt.Hosts[len(opt.Hosts)-1]
	}
	httpServer := &http.Server{Handler: g.wrap(s.routes()), ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = httpServer.Serve(listener) }()
	link := "http://127.0.0.1:" + port + "/?token=" + g.token
	fmt.Fprintf(out, "jin web is running at %s\nPress Ctrl+C to stop.\n", link)
	if opt.Remote {
		fmt.Fprintf(out, "Remote access is on at https://%s/\nClick the badge at the top of jin web to pair a phone (Tailscale must be on there).\n", s.remoteHost)
	}
	if !opt.NoOpen && !openBrowser(link) {
		fmt.Fprintln(out, "Could not open a browser; open the address above.")
	}
	<-ctx.Done()
	s.publish(map[string]string{"type": "quit"})
	s.hub.closeAll()
	shutdown, done := context.WithTimeout(context.Background(), 3*time.Second)
	defer done()
	_ = httpServer.Shutdown(shutdown)
	s.m.Shutdown()
	return 0, s.dataAction()
}

func (s *server) checkUpdate() {
	tag := jinupdate.Available(s.ctx, s.db, s.version, time.Now())
	if tag == "" {
		return
	}
	s.mu.Lock()
	s.latest = tag
	s.mu.Unlock()
	s.m.SetLatest(tag)
	s.publish(map[string]string{"type": "config"})
}
