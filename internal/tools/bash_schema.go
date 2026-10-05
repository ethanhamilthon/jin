package tools

import "encoding/json"

const bashDescription = "Run a shell command in the working directory. Each call starts a new shell: cd and exported variables do not carry over; to run in a subdirectory, pass dir instead of cd. No TTY and no stdin: use non-interactive flags (-y, --no-edit, -m). Output over 16 KB keeps its first and last 8 KB and names the file with the rest."

const (
	bashForeground = " A command still running at its timeout moves to the background (you get a task id, see jin async check/stop)."
	bashHeadless   = " A command still running at its timeout is killed."

	timeoutForeground = "Seconds to wait before the command moves to the background; at least 1. Defaults to 120 if omitted."
	timeoutHeadless   = "Seconds to wait before the command is killed; at least 1. Defaults to 120 if omitted."
)

func bashSchema(headless bool) json.RawMessage {
	description, timeout := bashDescription+bashForeground, timeoutForeground
	if headless {
		description, timeout = bashDescription+bashHeadless, timeoutHeadless
	}
	schema, _ := json.Marshal(map[string]any{"type": "function", "function": map[string]any{
		"name":        "bash",
		"description": description,
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{"type": "string", "description": "Shell command to execute"},
				"timeout": map[string]any{"type": "integer", "minimum": 1, "description": timeout},
				"dir":     map[string]any{"type": "string", "description": "Directory to run the command in, relative to the working directory or absolute. Defaults to the working directory."},
			},
			"required":             []string{"command"},
			"additionalProperties": false,
		},
	}})
	return schema
}
