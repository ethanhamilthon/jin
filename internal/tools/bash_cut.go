package tools

import (
	"bufio"
	"bytes"
	"io"
	"jin/internal/wire"
	"os"

	"jin/internal/tasklog"
)

// cutOutput is the result text of a log bigger than maxBashOutput: its
// start and its end, with a note in between that says how much was cut and
// where the whole output stays for the read tool. The log is kept.
func cutOutput(path string) (string, bool) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= maxBashOutput {
		return "", false
	}
	head, _, err := tasklog.Head(path, maxBashOutput/2)
	if err != nil {
		return "", false
	}
	tail, _, err := tasklog.Tail(path, maxBashOutput/2)
	if err != nil {
		return "", false
	}
	lines := countLines(path)
	shownLines := bytes.Count([]byte(head), []byte("\n")) + bytes.Count([]byte(tail), []byte("\n"))
	cutBytes := info.Size() - int64(len(head)) - int64(len(tail))
	return head + wire.OutputCut(cutBytes, max(lines-shownLines, 0), info.Size(), lines, path) + tail, true
}

func countLines(path string) int {
	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()
	reader := bufio.NewReaderSize(file, 64<<10)
	buf := make([]byte, 64<<10)
	lines := 0
	for {
		n, err := reader.Read(buf)
		lines += bytes.Count(buf[:n], []byte("\n"))
		if err == io.EOF || err != nil {
			return lines
		}
	}
}
