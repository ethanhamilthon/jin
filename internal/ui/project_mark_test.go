package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestProjectMarkPriorities(t *testing.T) {
	green, blue, purple := dotGreen, dotBlue, dotPurple
	t.Cleanup(func() { dotGreen, dotBlue, dotPurple = green, blue, purple })
	dotGreen, dotBlue, dotPurple = tcell.StyleDefault.Bold(true), tcell.StyleDefault.Italic(true), tcell.StyleDefault.Underline(true)
	a := &app{dir: "/open", sessions: map[string]*chatSession{}, tasksRunning: map[string]int{}}
	act := projectActivity{sessions: map[string][]string{"/done": {"old"}, "/tasks": {"bg"}}, unread: map[string]bool{"old": true}}
	a.sessions["live"] = &chatSession{id: "live", path: "/busy", working: true}
	a.sessions["seen"] = &chatSession{id: "seen", path: "/read", unread: true}
	a.tasksRunning["bg"] = 1
	cases := []struct {
		path    string
		on, off string
		style   string
	}{
		{"/open", "●", "●", "green"},
		{"/busy", "●", " ", "blue"},
		{"/done", "●", "●", "blue"},
		{"/read", "●", "●", "blue"},
		{"/tasks", "●", " ", "purple"},
		{"/idle", " ", " ", "purple"},
	}
	styles := map[string]any{"green": dotGreen, "blue": dotBlue, "purple": dotPurple}
	for _, c := range cases {
		for frame, want := range map[int]string{0: c.on, 6: c.off} {
			a.frame = frame
			mark, style := a.projectMark(c.path, act)
			if mark != want || style != styles[c.style] {
				t.Errorf("%s frame %d: %q %v, want %q %s", c.path, frame, mark, style, want, c.style)
			}
		}
	}
}
