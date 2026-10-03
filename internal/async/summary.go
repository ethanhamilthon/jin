package async

import (
	"fmt"
	"jin/internal/tasklog"
	"os"
	"regexp"
	"strings"
)

// ResultText wraps what the agent should see when a task ends. A negative
// exit leaves the exit attribute out. The closing tag inside the output is
// broken up, so output can never close the wrapper and pose as the user.
func ResultText(id, status string, exit int, body string) string {
	body = strings.TrimRight(body, " \t\r\n")
	if body == "" {
		body = "(no output)"
	}
	body = strings.ReplaceAll(body, closeTag, `<\/async-task-result`)
	attrs := fmt.Sprintf(`id=%q status=%q`, id, status)
	if exit >= 0 {
		attrs += fmt.Sprintf(` exit="%d"`, exit)
	}
	return "<async-task-result " + attrs + ">\n" + body + "\n</async-task-result>"
}

var attrPattern = regexp.MustCompile(`(\w+)="([^"]*)"`)

// Summary turns a result message into the lines the chat shows: a headline
// and the first lines of the output. ok is false for any other text.
func Summary(text string) (summary string, ok bool) {
	rest, found := strings.CutPrefix(text, "<async-task-result ")
	if !found {
		return "", false
	}
	header, body, _ := strings.Cut(rest, ">\n")
	attrs := map[string]string{}
	for _, m := range attrPattern.FindAllStringSubmatch(header, -1) {
		attrs[m[1]] = m[2]
	}
	line := "async task " + attrs["id"] + " " + attrs["status"]
	if exit, has := attrs["exit"]; has {
		line += " (exit " + exit + ")"
	}
	body = strings.TrimSuffix(strings.TrimRight(body, "\n"), "</async-task-result>")
	lines := strings.Split(strings.TrimRight(body, "\n"), "\n")
	if len(lines) > summaryLines {
		lines = append(lines[:summaryLines], "...")
	}
	return line + "\n" + strings.Join(lines, "\n"), true
}

// LogPath is the file that holds the output of a task.
func LogPath(id string) (string, error) {
	dir, err := dataDir()
	if err != nil {
		return "", err
	}
	return dir + string(os.PathSeparator) + id + ".log", nil
}

// Tail reads the last limit characters of a file; limit <= 0 reads all of
// it. The second result tells that the start was cut off.
func Tail(path string, limit int) (string, bool, error) { return tasklog.Tail(path, limit) }
