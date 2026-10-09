package prompts

import (
	_ "embed"
	"errors"
	"strings"
)

//go:embed system/plan.md
var systemPlan string

//go:embed system/review.md
var systemReview string

// ErrReserved reports that a prompt name is reserved for a built-in prompt.
var ErrReserved = errors.New("reserved prompt name")

// Info describes a prompt.
type Info struct {
	Name   string
	System bool
}

var systemPromptList = []Info{
	{Name: "plan", System: true},
	{Name: "review", System: true},
}

// IsSystem reports whether the name is a built-in prompt.
func IsSystem(name string) bool {
	clean := strings.TrimSuffix(strings.TrimSpace(name), ext)
	return clean == "plan" || clean == "review"
}

func systemBody(name string) (string, bool) {
	clean := strings.TrimSuffix(strings.TrimSpace(name), ext)
	switch clean {
	case "plan":
		return strings.TrimSpace(systemPlan), true
	case "review":
		return strings.TrimSpace(systemReview), true
	default:
		return "", false
	}
}
