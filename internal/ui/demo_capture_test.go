package ui

import (
	"encoding/json"
	"os"
	"testing"
)

func TestCaptureWorkspaceDemo(t *testing.T) {
	output := os.Getenv("JIN_DEMO_OUTPUT")
	if output == "" {
		t.Skip("set JIN_DEMO_OUTPUT to capture the scripted TUI demo")
	}
	a := demoApp(t)
	var frames []demoFrame
	capture := func(label string, count, delay int) {
		for range count {
			a.draw()
			frames = append(frames, captureDemoFrame(a, label, delay))
			a.tick()
		}
	}
	capture("One work view, one shared input", 1, 1200)
	a.openProjectSessions(a.dir)
	capture("Sessions of this project", 1, 1800)
	a.openProjects()
	capture("Projects: Enter opens a project", 1, 1400)
	a.sel = nil
	installDemoPanes(a)
	capture("Four sessions across different projects", 1, 1800)
	a.active.input = clusters("$ git status")
	a.active.cursor = len(a.active.input)
	capture("A leading dollar runs a shell command", 1, 1400)
	a.active.input, a.active.cursor = nil, 0
	a.active.working = true
	capture("Theme primary: foreground request", 8, 100)
	a.active.working = false
	a.tasksRunning[a.active.id] = 1
	capture("Purple: background task", 8, 100)
	data, err := json.Marshal(frames)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(output, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
