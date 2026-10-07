package web

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"jin/internal/tasklog"
	"jin/internal/tasks"
)

const taskTail = 20000

type taskView struct {
	ID      string    `json:"id"`
	Owner   string    `json:"owner"`
	Command string    `json:"command"`
	Dir     string    `json:"dir"`
	Status  string    `json:"status"`
	Exit    int       `json:"exit"`
	Started time.Time `json:"started"`
}

func (s *server) taskRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/tasks", api(func(r *http.Request) (any, error) {
		views := []taskView{}
		for _, t := range tasks.Shared().List("") {
			views = append(views, taskView{t.ID, t.Owner, t.Command, t.Dir, t.Status, t.Exit, t.Started})
		}
		return views, nil
	}))
	mux.HandleFunc("GET /api/tasks/{id}/output", api(func(r *http.Request) (any, error) {
		for _, t := range tasks.Shared().List("") {
			if t.ID != r.PathValue("id") {
				continue
			}
			text, truncated, err := tasklog.Tail(t.Log, taskTail)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
			return map[string]any{"text": strings.TrimRight(text, "\n"), "truncated": truncated, "status": t.Status}, nil
		}
		return nil, errors.New("task " + r.PathValue("id") + " not found")
	}))
	mux.HandleFunc("POST /api/tasks/{id}/stop", api(func(r *http.Request) (any, error) {
		err := tasks.Shared().StopAny(r.PathValue("id"))
		s.publish(map[string]string{"type": "tasks"})
		return done(err)
	}))
}
