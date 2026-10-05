package tasks

import (
	"fmt"
	"regexp"
	"strings"
)

// Event tells the owner that a task ended. Tasks the agent stopped itself,
// and tasks killed because jin exits, send none.
type Event struct {
	Owner string
	Task  Info
	// Text is the message for the agent, made by ResultText.
	Text string
}

const closeTag = "</task-result"

// ResultText wraps what the agent sees when a task ends. A negative exit
// leaves the exit attribute out. A closing tag inside the output is broken
// up, so output can never close the wrapper and pose as the user.
func ResultText(id, status string, exit int, body string) string {
	body = strings.TrimRight(body, " \t\r\n")
	if body == "" {
		body = "(no output)"
	}
	body = strings.ReplaceAll(body, closeTag, `<\/task-result`)
	attrs := fmt.Sprintf(`id=%q status=%q`, id, status)
	if exit >= 0 {
		attrs += fmt.Sprintf(` exit="%d"`, exit)
	}
	return "<task-result " + attrs + ">\n" + body + "\n" + closeTag + ">"
}

const summaryLines = 6

var attrPattern = regexp.MustCompile(`(\w+)="([^"]*)"`)

// Summary turns a result message into the lines the chat shows: a headline
// and the first lines of the output. ok is false for any other text.
func Summary(text string) (summary string, ok bool) {
	rest, found := strings.CutPrefix(text, "<task-result ")
	if !found {
		return "", false
	}
	header, body, _ := strings.Cut(rest, ">\n")
	attrs := map[string]string{}
	for _, m := range attrPattern.FindAllStringSubmatch(header, -1) {
		attrs[m[1]] = m[2]
	}
	line := "background task " + attrs["id"] + " " + attrs["status"]
	if exit, has := attrs["exit"]; has {
		line += " (exit " + exit + ")"
	}
	body = strings.TrimSuffix(strings.TrimRight(body, "\n"), closeTag+">")
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(lines) > summaryLines {
		lines = append(lines[:summaryLines], "...")
	}
	return line + "\n" + strings.Join(lines, "\n"), true
}
