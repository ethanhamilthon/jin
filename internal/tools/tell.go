package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

const tellDescription = "Talk to the user without ending your turn. mode message shows text in the chat at once, " +
	"for progress worth knowing (\"tests take about two minutes\"); keep it to a line or two and do not repeat it in your answer. " +
	"mode suggest offers the user's likely next request, shown greyed in the empty input when you finish; " +
	"the user sends it with Enter. Write it as the user would, short (\"Run the migrations on staging\"); the latest suggestion wins."

// Teller shows a message or a suggestion to the user.
type Teller func(mode, text string)

type tellerKey struct{}

// WithTeller tells the tell_user tool how to reach the user.
func WithTeller(ctx context.Context, fn Teller) context.Context {
	return context.WithValue(ctx, tellerKey{}, fn)
}

// Tell is the tell_user tool.
type Tell struct{}

func (Tell) Name() string { return "tell_user" }

func (Tell) Schema() json.RawMessage {
	schema, _ := json.Marshal(map[string]any{"type": "function", "function": map[string]any{
		"name":        "tell_user",
		"description": tellDescription,
		"parameters": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"mode": map[string]any{"type": "string", "enum": []string{"message", "suggest"}},
				"text": map[string]any{"type": "string"},
			},
			"required":             []string{"mode", "text"},
			"additionalProperties": false,
		},
	}})
	return schema
}

func parseTell(argumentsJSON string) (string, string, error) {
	var args struct{ Mode, Text string }
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil {
		return "", "", errors.New("invalid tell_user arguments")
	}
	text := strings.TrimSpace(args.Text)
	if text == "" {
		return "", "", errors.New("tell_user needs text")
	}
	if args.Mode != "message" && args.Mode != "suggest" {
		return "", "", errors.New(`mode must be "message" or "suggest"`)
	}
	return args.Mode, text, nil
}

func (Tell) Summary(argumentsJSON string) (string, bool) {
	mode, text, err := parseTell(argumentsJSON)
	return mode + ": " + text, err == nil
}

func (Tell) Run(ctx context.Context, argumentsJSON string) (string, error) {
	mode, text, err := parseTell(argumentsJSON)
	if err != nil {
		return "", err
	}
	tell, ok := ctx.Value(tellerKey{}).(Teller)
	if !ok {
		return "", errors.New("no user to tell here")
	}
	tell(mode, text)
	if mode == "suggest" {
		return "suggestion set", nil
	}
	return "shown to the user", nil
}
