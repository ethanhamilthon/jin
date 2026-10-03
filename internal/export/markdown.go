package export

import (
	"fmt"
	"io"
	"strings"

	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/store"
)

// writeMarkdown prints the conversation as a readable transcript: user and
// assistant messages in full, tool calls with their arguments and results
// in folded code blocks. The system prompt is left out.
func writeMarkdown(out io.Writer, s store.Session, messages []provider.Message) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", orDefault(s.Title, "Session "+s.ID))
	fmt.Fprintf(&b, "- Session: `%s`\n- Directory: `%s`\n- Model: `%s`\n- Started: %s\n\n",
		s.ID, s.Path, s.Model, s.CreatedAt.Format("2006-01-02 15:04"))
	for _, msg := range messages {
		writeMessage(&b, msg)
	}
	_, err := io.WriteString(out, b.String())
	return err
}

func writeMessage(b *strings.Builder, msg provider.Message) {
	switch {
	case msg.Role == "tool":
		fmt.Fprintf(b, "<details><summary>Result</summary>\n\n%s\n</details>\n\n", fence(msg.Content))
	case core.IsSummary(msg):
		fmt.Fprintf(b, "## Summary (compacted)\n\n%s\n\n", msg.Content)
	case msg.Role == "user":
		fmt.Fprintf(b, "## User\n\n%s\n\n", prompts.Strip(core.StripNotes(msg.Content)))
	case msg.Role == "assistant":
		writeAssistant(b, msg)
	}
}

func writeAssistant(b *strings.Builder, msg provider.Message) {
	if msg.Content == "" && len(msg.ToolCalls) == 0 {
		return
	}
	b.WriteString("## Assistant\n\n")
	if msg.Content != "" {
		b.WriteString(msg.Content + "\n\n")
	}
	for _, call := range msg.ToolCalls {
		fmt.Fprintf(b, "**%s**\n\n%s\n", call.Function.Name, fence(call.Function.Arguments))
	}
}

// fence wraps text in a code block whose fence is longer than any run of
// backticks inside it.
func fence(text string) string {
	ticks := "```"
	for strings.Contains(text, ticks) {
		ticks += "`"
	}
	return ticks + "\n" + strings.TrimRight(text, "\n") + "\n" + ticks + "\n"
}

func orDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
