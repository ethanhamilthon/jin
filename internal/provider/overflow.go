package provider

import (
	"errors"
	"net/http"
	"strings"
)

var overflowPhrases = []string{
	"context_length_exceeded",
	"prompt is too long",
	"maximum context length",
	"context window",
}

// IsContextOverflow reports whether the provider refused a request because
// the conversation does not fit into the model's context window.
func IsContextOverflow(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	for _, phrase := range overflowPhrases {
		if strings.Contains(text, phrase) {
			return true
		}
	}
	var status *statusError
	return errors.As(err, &status) && status.status == http.StatusRequestEntityTooLarge && strings.Contains(text, "context")
}
