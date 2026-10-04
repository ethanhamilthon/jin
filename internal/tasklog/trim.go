package tasklog

import (
	"errors"
	"io"
	"os"
)

// MaxLog is the most a task log keeps on disk while its task runs.
const MaxLog = 64 << 20

const cutNote = "\n[... output cut here to keep the log small ...]\n"

// Trim keeps a log from growing without end: above MaxLog it keeps the first
// and the last quarter with a note between them. A process that writes with
// O_APPEND goes on at the new end; a few bytes written during the cut may be
// lost.
func Trim(path string) error { return trim(path, MaxLog) }

func trim(path string, max int64) error {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= max {
		return err
	}
	part := max / 4
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	head, tail := make([]byte, part), make([]byte, part)
	_, headErr := file.ReadAt(head, 0)
	got, tailErr := file.ReadAt(tail, info.Size()-part)
	file.Close()
	if err := errors.Join(headErr, tailErr); err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	if err := os.Truncate(path, 0); err != nil {
		return err
	}
	out, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = out.Write(append(append(head, cutNote...), tail[:got]...))
	return err
}
