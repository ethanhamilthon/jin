package ui

func (a *app) requestQuit() {
	if !a.anyWorking() {
		a.quit = true
		return
	}
	options := []option{{label: "Quit", detail: "stop running requests", value: "quit"}, {label: "Cancel", value: "cancel"}}
	a.openList("A request is still running. Quit jin?", options, "cancel", func(value string) error {
		a.quit = value == "quit"
		return nil
	})
}
