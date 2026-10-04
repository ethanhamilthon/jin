package headless

import (
	"context"
	"errors"
	"io"
	"strings"
)

// maxStdin caps what is read from a pipe.
const maxStdin = 10 << 20

// BuildPrompt joins the prompt words and piped input. With words, the pipe is
// context appended in a <stdin> block; without, the pipe is the prompt.
func BuildPrompt(ctx context.Context, words string, stdin io.Reader, piped bool) (string, error) {
	words = strings.TrimSpace(words)
	var data string
	if piped && stdin != nil {
		raw, err := readLimited(ctx, stdin)
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

// readLimited reads at most maxStdin+1 bytes and gives up when ctx ends, even
// if the pipe never closes.
func readLimited(ctx context.Context, stdin io.Reader) ([]byte, error) {
	type readResult struct {
		raw []byte
		err error
	}
	done := make(chan readResult, 1)
	go func() {
		raw, err := io.ReadAll(io.LimitReader(stdin, maxStdin+1))
		done <- readResult{raw, err}
	}()
	select {
	case res := <-done:
		return res.raw, res.err
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
}
