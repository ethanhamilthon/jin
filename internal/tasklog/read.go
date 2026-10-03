package tasklog

import (
	"errors"
	"io"
	"os"
	"strings"
	"unicode/utf8"
)

// Head reads the first n bytes. The second result tells that more follows.
func Head(path string, n int) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()
	buf := make([]byte, n+1)
	got, err := io.ReadFull(file, buf)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", false, err
	}
	truncated := got > n
	if truncated {
		got = n
	}
	return strings.ToValidUTF8(string(buf[:got]), ""), truncated, nil
}

// Tail reads the last limit characters; limit <= 0 reads everything. The
// second result tells that the start was cut off.
func Tail(path string, limit int) (string, bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", false, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", false, err
	}
	start := int64(0)
	if limit > 0 {
		// A character takes at most 4 bytes.
		if cut := info.Size() - int64(limit)*utf8.UTFMax; cut > 0 {
			start = cut
		}
	}
	buf := make([]byte, info.Size()-start)
	if _, err := file.ReadAt(buf, start); err != nil && info.Size() > start && !errors.Is(err, io.EOF) {
		return "", false, err
	}
	text := strings.ToValidUTF8(string(buf), "")
	truncated := start > 0
	if limit > 0 && utf8.RuneCountInString(text) > limit {
		runes := []rune(text)
		text, truncated = string(runes[len(runes)-limit:]), true
	}
	return text, truncated, nil
}

// Trim keeps a log from growing without end: above max bytes it keeps only
// the last keep bytes. A process that writes with O_APPEND goes on at the new
// end; a few bytes written during the cut may land before the kept part.
func Trim(path string, max, keep int64) error {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= max {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	buf := make([]byte, keep)
	got, err := file.ReadAt(buf, info.Size()-keep)
	file.Close()
	if err != nil && !errors.Is(err, io.EOF) {
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
	_, err = out.Write(buf[:got])
	return err
}
