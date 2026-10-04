package store

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestMakeSnippetMultilingual(t *testing.T) {
	// Thirty multi-byte characters, keyword, thirty multi-byte characters
	prefix := strings.Repeat("界", 50)
	suffix := strings.Repeat("界", 50)
	text := prefix + "match" + suffix

	got := makeSnippet(text, "match")
	if !utf8.ValidString(got) {
		t.Fatalf("makeSnippet produced invalid UTF-8: %q", got)
	}
	if !strings.Contains(got, "match") {
		t.Fatalf("makeSnippet result missing target word: %q", got)
	}
	if !strings.HasPrefix(got, "...") || !strings.HasSuffix(got, "...") {
		t.Fatalf("expected ellipsis on both sides: %q", got)
	}
}

func TestMakeSnippetCyrillic(t *testing.T) {
	text := "Это длинное русское предложение для проверки работы сниппетов в поиске сессий jin"
	got := makeSnippet(text, "работы")
	if !utf8.ValidString(got) {
		t.Fatalf("invalid UTF-8 in cyrillic snippet: %q", got)
	}
	if !strings.Contains(got, "работы") {
		t.Fatalf("snippet missing target word: %q", got)
	}
}

func TestMakeSnippetFallbackTruncate(t *testing.T) {
	longText := strings.Repeat("你好", 60)
	got := makeSnippet(longText, "nonexistent")
	if !utf8.ValidString(got) {
		t.Fatalf("invalid UTF-8 in fallback snippet: %q", got)
	}
	if !strings.HasSuffix(got, "...") {
		t.Fatalf("expected trailing ellipsis in fallback: %q", got)
	}
}
