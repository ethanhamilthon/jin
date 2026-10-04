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
	lower := strings.ToLower(cleaned)
	target := strings.ToLower(word)
	idx := strings.Index(lower, target)
	if idx < 0 {
		if len(cleaned) > 80 {
			return cleaned[:80] + "..."
		}
		return cleaned
	}
	start := max(0, idx-snippetRadius)
	end := min(len(cleaned), idx+len(word)+snippetRadius)
	snippet := cleaned[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(cleaned) {
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
