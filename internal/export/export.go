// Package export prints a saved session as Markdown or JSON.
package export

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"jin/internal/provider"
	"jin/internal/store"
)

const usageText = "usage: jin export <session-id> [--md|--json]"

// Main runs `jin export`. It returns the exit code.
func Main(args []string, db *store.DB, out, errOut io.Writer) int {
	id, asJSON, err := parseArgs(args)
	if err != nil {
		fmt.Fprintln(errOut, "jin:", err)
		return 2
	}
	session, found, err := findSession(db, id)
	if err == nil && !found {
		err = errors.New("no session " + id)
	}
	if err != nil {
		fmt.Fprintln(errOut, "jin:", err)
		return 1
	}
	messages, err := db.LoadMessages(session.ID)
	if err != nil {
		fmt.Fprintln(errOut, "jin:", err)
		return 1
	}
	if asJSON {
		err = writeJSON(out, session, messages)
	} else {
		err = writeMarkdown(out, session, messages)
	}
	if err != nil {
		fmt.Fprintln(errOut, "jin:", err)
		return 1
	}
	return 0
}

func parseArgs(args []string) (string, bool, error) {
	var id string
	asJSON := false
	for _, arg := range args {
		switch arg {
		case "--json":
			asJSON = true
		case "--md":
			asJSON = false
		default:
			if strings.HasPrefix(arg, "-") || id != "" {
				return "", false, errors.New(usageText)
			}
			id = arg
		}
	}
	if id == "" {
		return "", false, errors.New(usageText)
	}
	return id, asJSON, nil
}

// findSession accepts a full id or a unique prefix of one.
func findSession(db *store.DB, id string) (store.Session, bool, error) {
	if s, ok, err := db.GetSession(id); ok || err != nil {
		return s, ok, err
	}
	return db.SessionByPrefix(id)
}

type jsonExport struct {
	ID       string             `json:"id"`
	Path     string             `json:"path"`
	Title    string             `json:"title"`
	Model    string             `json:"model"`
	Effort   string             `json:"effort,omitempty"`
	Created  time.Time          `json:"created_at"`
	Updated  time.Time          `json:"updated_at"`
	Cost     float64            `json:"cost"`
	Messages []provider.Message `json:"messages"`
}

func writeJSON(out io.Writer, s store.Session, messages []provider.Message) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(jsonExport{
		ID: s.ID, Path: s.Path, Title: s.Title, Model: s.Model, Effort: s.Effort,
		Created: s.CreatedAt.UTC(), Updated: s.UpdatedAt.UTC(), Cost: s.Usage.Cost, Messages: messages,
	})
}
