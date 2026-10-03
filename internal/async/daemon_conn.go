//go:build unix

package async

import (
	"encoding/json"
	"io"
	"net"
	"strconv"
	"time"
)

func (d *daemon) accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go d.serve(conn)
	}
}

func (d *daemon) serve(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(connLimit))
	var req request
	if err := json.NewDecoder(io.LimitReader(conn, requestMax)).Decode(&req); err != nil {
		return
	}
	d.touch()
	resp := d.handle(req)
	resp.Version = d.version
	_ = json.NewEncoder(conn).Encode(resp)
}

func (d *daemon) handle(req request) response {
	switch req.Op {
	case opHello:
		d.mu.Lock()
		idle := len(d.tasks) == 0
		d.mu.Unlock()
		return response{OK: true, Idle: idle}
	case opShutdown:
		d.mu.Lock()
		idle := len(d.tasks) == 0
		d.mu.Unlock()
		if !idle {
			return response{Error: "daemon has running tasks"}
		}
		go func() { time.Sleep(100 * time.Millisecond); close(d.done) }()
		return response{OK: true}
	case opRun:
		id, err := d.run(req)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true, ID: id}
	case opAdopt:
		id, err := d.adopt(req)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true, ID: id}
	case opInput:
		if err := d.input(req); err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true}
	case opStop:
		if err := d.stop(req); err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true}
	}
	return response{Error: "unknown operation " + strconv.Quote(req.Op)}
}

func (d *daemon) idleFor() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.tasks) > 0 {
		return 0
	}
	return time.Since(d.active)
}

func (d *daemon) touch() {
	d.mu.Lock()
	d.active = time.Now()
	d.mu.Unlock()
}
