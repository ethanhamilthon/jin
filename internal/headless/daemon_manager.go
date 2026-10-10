package headless

import (
	"context"

	"jin/internal/daemon"
	"jin/internal/session"
)

type sessionBackend interface {
	Handoff(string, string) error
	Compact(string) error
	Reload(string) error
	Snapshot(string) (session.Snapshot, error)
	Points(string) ([]session.RewindPoint, error)
	Fork(string, int) (session.Snapshot, error)
	Context(string) (session.ContextReport, error)
	Live() []session.State
}

type daemonManager struct {
	client *daemon.Client
	ctx    context.Context
}

func (m daemonManager) action(action, id string, value any) error {
	return m.client.Command(m.ctx, daemon.Command{Action: action, Session: id}, value)
}
func (m daemonManager) Handoff(id, origin string) error {
	return m.client.Command(m.ctx, daemon.Command{Action: "handoff", Session: id, Origin: origin}, nil)
}
func (m daemonManager) Compact(id string) error { return m.action("compact", id, nil) }
func (m daemonManager) Reload(id string) error  { return m.action("reload", id, nil) }
func (m daemonManager) Snapshot(id string) (session.Snapshot, error) {
	return m.client.Snapshot(m.ctx, id)
}
func (m daemonManager) Points(id string) (value []session.RewindPoint, err error) {
	err = m.action("points", id, &value)
	return
}
func (m daemonManager) Context(id string) (value session.ContextReport, err error) {
	err = m.action("context", id, &value)
	return
}
func (m daemonManager) Fork(id string, point int) (value session.Snapshot, err error) {
	err = m.client.Command(m.ctx, daemon.Command{Action: "fork", Session: id, Point: point}, &value)
	return
}
func (m daemonManager) Live() (value []session.State) { _ = m.action("live", "", &value); return }

var daemonClient *daemon.Client

func UseDaemon(client *daemon.Client) { daemonClient = client }

// daemonOrigin names the shared client that asks for a handoff: the daemon
// sends the brief to every client, and the origin tells the initiator.
func daemonOrigin() string {
	if daemonClient == nil {
		return ""
	}
	return daemonClient.ID
}
