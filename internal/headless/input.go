package headless

import (
	"errors"
	"io"
	"strings"
)

// maxStdin caps what is read from a pipe.
const maxStdin = 10 << 20

// BuildPrompt joins the prompt words and piped input. With words, the pipe is
// context appended in a <stdin> block; without, the pipe is the prompt.
func BuildPrompt(words string, stdin io.Reader, piped bool) (string, error) {
	words = strings.TrimSpace(words)
	var data string
	if piped && stdin != nil {
		raw, err := io.ReadAll(io.LimitReader(stdin, maxStdin+1))
		if err != nil {
			return "", err
		}
		if len(raw) > maxStdin {
			return "", errors.New("stdin is larger than 10 MB")
		}
		data = string(raw)
	}
	switch {
	case words == "" && strings.TrimSpace(data) == "":
		if !piped {
			return "", errors.New("no prompt: pass it as an argument or on stdin")
		}
		return "", errors.New("empty prompt")
	case words == "":
		return strings.TrimSpace(data), nil
	case strings.TrimSpace(data) == "":
		return words, nil
	}
	return words + "\n\n<stdin>\n" + strings.TrimRight(data, "\n") + "\n</stdin>", nil
}
