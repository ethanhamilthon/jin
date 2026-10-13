package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"jin/internal/wire"
)

// Event tells the owner that a task ended. Tasks the agent stopped itself,
// and tasks killed because jin exits, send none.
type Event struct {
	Owner string
	Task  Info
	// Text is the message for the agent, made by ResultText.
	Text string
}

const closeTag = wire.TaskResultClose

// ResultText wraps what the agent sees when a task ends. A negative exit
// leaves the exit attribute out. A closing tag inside the output is broken
// up, so output can never close the wrapper and pose as the user.
func ResultText(id, status string, exit int, body string) string {
	body = strings.TrimRight(body, " \t\r\n")
	if body == "" {
		body = "(no output)"
	}
	body = strings.ReplaceAll(body, closeTag, wire.TaskResultBroken)
	attrs := fmt.Sprintf(`id=%q status=%q`, id, status)
	if exit >= 0 {
		attrs += fmt.Sprintf(` exit="%d"`, exit)
	}
	return wire.TaskResultOpen + attrs + ">\n" + body + "\n" + closeTag + ">"
}

var attrPattern = regexp.MustCompile(`(\w+)="([^"]*)"`)

// Parse splits a result message into a headline and the lines of the output.
// ok is false for any other text.
func Parse(text string) (headline string, lines []string, ok bool) {
	rest, found := strings.CutPrefix(text, wire.TaskResultOpen)
	if !found {
		return "", nil, false
	}
	header, body, _ := strings.Cut(rest, ">\n")
	attrs := map[string]string{}
	for _, m := range attrPattern.FindAllStringSubmatch(header, -1) {
		attrs[m[1]] = m[2]
	}
	headline = attrs["id"] + " " + attrs["status"]
	if exit, has := attrs["exit"]; has {
		headline += " (exit " + exit + ")"
	}
	body = strings.TrimSuffix(strings.TrimRight(body, "\n"), closeTag+">")
	return headline, strings.Split(strings.TrimRight(body, "\n"), "\n"), true
}
