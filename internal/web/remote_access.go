package web

import (
	"context"
	"net"
	"os/exec"
	"runtime"
	"sync"
)

const remoteKey = "web.remote"

type RemoteState struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

type remoteAccess struct {
	mu      sync.Mutex
	service *Service
	host    string
	failure string
	awake   *exec.Cmd
}

func (r *remoteAccess) state() RemoteState {
	r.mu.Lock()
	defer r.mu.Unlock()
	st := RemoteState{Enabled: r.host != "", Error: r.failure}
	if r.host != "" {
		st.URL = "https://" + r.host + "/"
	}
	return st
}

func (r *remoteAccess) set(ctx context.Context, enabled bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if enabled == (r.host != "") {
		return r.service.server.db.SetSetting(remoteKey, boolText(enabled))
	}
	if enabled {
		lock, err := lockServeEntry(ctx)
		if err != nil {
			r.failure = err.Error()
			return err
		}
		defer lock.release()
		host, err := tailnetName(ctx)
		_, port, _ := net.SplitHostPort(r.service.listener.Addr().String())
		if err == nil {
			err = tailnetAvailable(ctx, port)
		}
		if err != nil {
			r.failure = err.Error()
			return err
		}
		r.service.server.guard.setRemote(host, true)
		if err := tailnetServe(ctx, port); err != nil {
			r.service.server.guard.setRemote(host, false)
			r.failure = err.Error()
			return err
		}
		r.host, r.failure = host, ""
		if runtime.GOOS == "darwin" {
			r.awake = exec.CommandContext(r.service.server.ctx, "caffeinate", "-i")
			if err := r.awake.Start(); err != nil {
				r.awake = nil
			} else {
				go func(cmd *exec.Cmd) { _ = cmd.Wait() }(r.awake)
			}
		}
	} else {
		if err := r.disable(ctx); err != nil {
			r.failure = err.Error()
			return err
		}
		r.service.server.guard.setRemote(r.host, false)
		r.host, r.failure = "", ""
		if r.awake != nil {
			_ = r.awake.Process.Kill()
			r.awake = nil
		}
	}
	r.service.server.publish(map[string]string{"type": "devices"})
	return r.service.server.db.SetSetting(remoteKey, boolText(enabled))
}

// disable turns the Serve entry off only while it still points at this
// daemon. A configuration somebody else changed after jin enabled remote
// access is theirs, and jin leaves it alone.
func (r *remoteAccess) disable(ctx context.Context) error {
	lock, err := lockServeEntry(ctx)
	if err != nil {
		return err
	}
	defer lock.release()
	_, port, _ := net.SplitHostPort(r.service.listener.Addr().String())
	proxy, occupied, err := localServeEntry(ctx)
	if err == nil && occupied && proxy != proxyURL(port) {
		return nil
	}
	return tailnetDisable(ctx)
}

func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
