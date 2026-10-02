package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
)

const askSchema = `{"type":"function","function":{"name":"ask_user","description":"Ask the user one or more questions and wait for the answers. Use it only for real ambiguity that blocks the work. Each question may offer options; the user can always type a free answer instead.","parameters":{"type":"object","properties":{"questions":{"type":"array","items":{"type":"object","properties":{"question":{"type":"string"},"options":{"type":"array","items":{"type":"string"},"description":"Optional suggested answers"}},"required":["question"],"additionalProperties":false}}},"required":["questions"],"additionalProperties":false}}}`

type Question struct {
	Question string   `json:"question"`
	Options  []string `json:"options,omitempty"`
}

// AskFunc shows the questions to the user and returns one answer for each.
type AskFunc func(ctx context.Context, questions []Question) ([]string, error)

type askKey struct{}

// WithAsker tells the ask_user tool how to reach the user.
func WithAsker(ctx context.Context, fn AskFunc) context.Context {
	return context.WithValue(ctx, askKey{}, fn)
}

type Ask struct{}

func NewAsk() Ask { return Ask{} }

func (Ask) Name() string { return "ask_user" }

func (Ask) Schema() json.RawMessage { return json.RawMessage(askSchema) }

func parseQuestions(argumentsJSON string) ([]Question, bool) {
	var args struct {
		Questions []Question `json:"questions"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || len(args.Questions) == 0 {
		return nil, false
	}
	for i := range args.Questions {
		args.Questions[i].Question = strings.TrimSpace(args.Questions[i].Question)
		if args.Questions[i].Question == "" {
			return nil, false
		}
	}
	return args.Questions, true
}

func (Ask) Summary(argumentsJSON string) (string, bool) {
	questions, ok := parseQuestions(argumentsJSON)
	if !ok {
		return "", false
	}
	text := questions[0].Question
	if len(questions) > 1 {
		text += " (+" + strconv.Itoa(len(questions)-1) + " more)"
	}
	return text, true
}

func (Ask) Run(ctx context.Context, argumentsJSON string) (string, error) {
	questions, ok := parseQuestions(argumentsJSON)
	if !ok {
		return "", errors.New("invalid ask_user tool arguments")
	}
	ask, ok := ctx.Value(askKey{}).(AskFunc)
	if !ok {
		return "", errors.New("the user cannot be asked in this run")
	}
	answers, err := ask(ctx, questions)
	if err != nil {
		return "", err
	}
	return FormatAnswers(questions, answers), nil
}

// FormatAnswers renders one "question → answer" line per question.
func FormatAnswers(questions []Question, answers []string) string {
	lines := make([]string, len(questions))
	for i, q := range questions {
		answer := ""
		if i < len(answers) {
			answer = answers[i]
		}
		lines[i] = q.Question + " → " + answer
	}
	return strings.Join(lines, "\n")
}
