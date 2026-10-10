package web

import (
	"context"
	"net"
	"net/http"
	"time"

	"jin/internal/session"
	"jin/internal/store"
)

type Service struct {
	server   *server
	http     *http.Server
	listener net.Listener
	cancel   context.CancelFunc
}

func StartService(ctx context.Context, manager *session.Manager, db *store.DB, dir, version string, opt Options) (*Service, error) {
	listener, err := listen(opt.Port)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	s := &server{ctx: ctx, db: db, m: manager, hub: newHub(), dir: dir, version: version, quit: cancel, shared: true}
	_, port, _ := net.SplitHostPort(listener.Addr().String())
	s.guard = newGuard(port, opt.Hosts, db)
	service := &Service{server: s, listener: listener, cancel: cancel}
	s.remote = &remoteAccess{service: service}
	service.http = &http.Server{Handler: s.guard.wrap(s.routes()), ReadHeaderTimeout: 10 * time.Second}
	events, unsubscribe := manager.Subscribe()
	go func() {
		defer unsubscribe()
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-events:
				s.publish(event)
			}
		}
	}()
	go service.http.Serve(listener)
	go s.checkUpdate()
	return service, nil
}

func (s *Service) URL() string {
	return "http://" + s.listener.Addr().String() + "/?token=" + s.server.guard.token
}

func (s *Service) Close() {
	if s.server.remote.state().Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = tailnetDisable(ctx)
		cancel()
	}
	s.cancel()
	s.server.hub.closeAll()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = s.http.Shutdown(ctx)
}
