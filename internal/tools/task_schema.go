package tools

import "encoding/json"

const taskDescription = "Run commands in the background and manage them. start runs a command as a task and returns its id at once; " +
	"when the task ends, its result arrives as a message by itself, so never wait, sleep or poll for it. " +
	"Use it for servers, watchers, long builds and test runs, sub-agents (jin -p), and anything that may outlive the bash timeout; " +
	"do not start background jobs with & or nohup in bash. check shows the status and the end of the output (pass limit, e.g. 2000). " +
	"input writes a line to a task started with stdin. stop ends a task. list shows the tasks of this session. " +
	"Tasks stop when jin exits."

func taskSchema() json.RawMessage {
	schema, _ := json.Marshal(map[string]any{"type": "function", "function": map[string]any{
		"name":        "task",
		"description": taskDescription,
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action":  map[string]any{"type": "string", "enum": []string{"start", "check", "input", "stop", "list"}},
				"command": map[string]any{"type": "string", "description": "start: the shell command to run"},
				"dir":     map[string]any{"type": "string", "description": "start: directory to run in, relative to the working directory or absolute"},
				"stdin":   map[string]any{"type": "boolean", "description": "start: give the task a stdin pipe for input; otherwise stdin is closed"},
				"id":      map[string]any{"type": "string", "description": "check, input, stop: the task id"},
				"limit":   map[string]any{"type": "integer", "minimum": 1, "description": "check: show only the last limit characters of the output"},
				"text":    map[string]any{"type": "string", "description": "input: the text to write; a newline is added"},
			},
			"required":             []string{"action"},
			"additionalProperties": false,
		},
	}})
	return schema
}
