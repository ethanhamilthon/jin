package web

import (
	"errors"
	"net/http"

	"jin/internal/datadir"
	"jin/internal/tasks"
)

func (s *server) dataRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/data", api(func(r *http.Request) (any, error) {
		current, err := datadir.Current()
		return map[string]string{"current": current}, err
	}))
	mux.HandleFunc("POST /api/data", api(func(r *http.Request) (any, error) {
		var body struct{ Kind, Path string }
		if err := decode(r, &body); err != nil {
			return nil, err
		}
		current, err := datadir.Current()
		if err != nil {
			return nil, err
		}
		if err := s.dataMoveAllowed(); err != nil {
			return nil, err
		}
		dest, err := datadir.Expand(body.Path)
		if err != nil {
			return nil, err
		}
		switch body.Kind {
		case "reset":
			err = datadir.CheckReset(current, dest)
		case "swap":
			err = datadir.CheckSwap(current, dest)
		default:
			err = errors.New("kind must be reset or swap")
		}
		if err != nil {
			return nil, err
		}
		s.mu.Lock()
		s.action = &datadir.Action{Kind: body.Kind, Path: dest}
		s.mu.Unlock()
		s.quit()
		return nil, nil
	}))
}

// dataMoveAllowed refuses while work runs: a moved database under a running
// request or background task would lose its results.
func (s *server) dataMoveAllowed() error {
	if s.m.Working() {
		return errors.New("a request is still running; stop it first")
	}
	if len(tasks.Shared().Running("")) > 0 {
		return errors.New("background tasks are running; stop them in Tasks first")
	}
	return datadir.CheckAlone()
}

func (s *server) dataAction() *datadir.Action {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.action
}
