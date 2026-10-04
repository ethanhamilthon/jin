package store

import (
	"strings"
)

const snippetRadius = 40

func escapeLike(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == '%' || r == '_' || r == '\\' {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func cleanWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func makeSnippet(text, word string) string {
	cleaned := cleanWhitespace(text)
	runes := []rune(cleaned)
	lowerRunes := []rune(strings.ToLower(cleaned))
	target := []rune(strings.ToLower(word))
	idx := -1
	if len(target) > 0 {
		for i := 0; i <= len(lowerRunes)-len(target); i++ {
			match := true
			for j := 0; j < len(target); j++ {
				if lowerRunes[i+j] != target[j] {
					match = false
					break
				}
			}
			if match {
				idx = i
				break
			}
		}
	}
	if idx < 0 {
		if len(runes) > 80 {
			return string(runes[:80]) + "..."
		}
		return string(runes)
	}
	start := max(0, idx-snippetRadius)
	end := min(len(runes), idx+len(target)+snippetRadius)
	snippet := string(runes[start:end])
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet = snippet + "..."
	}
	return snippet
}

func (db *DB) findSnippet(sessionID, title string, words []string) string {
	rows, err := db.sql.Query(`SELECT json_extract(data, '$.content') FROM messages
		WHERE session_id = ? AND json_extract(data, '$.role') = 'user' ORDER BY id`, sessionID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var content string
			if err := rows.Scan(&content); err != nil {
				continue
			}
			lower := strings.ToLower(content)
			for _, w := range words {
				if strings.Contains(lower, strings.ToLower(w)) {
					return makeSnippet(content, w)
				}
			}
		}
	}
	lowerTitle := strings.ToLower(title)
	for _, w := range words {
		if strings.Contains(lowerTitle, strings.ToLower(w)) {
			return makeSnippet(title, w)
		}
	}
	return cleanWhitespace(title)
}
