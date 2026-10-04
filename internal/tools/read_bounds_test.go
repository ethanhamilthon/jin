package tools

import (
	"context"
	"io"
	"strings"
	"testing"
)

type countingReader struct {
	data []byte
	read int
}

func (c *countingReader) Read(p []byte) (int, error) {
	if len(c.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, c.data)
	c.data = c.data[n:]
	c.read += n
	return n, nil
}

func TestReadLimitDoesNotConsumeNextLine(t *testing.T) {
	reader := &countingReader{data: []byte("short\n" + strings.Repeat("z", 5<<20) + "\n")}
	got, err := readLines(context.Background(), reader, readArgs{Limit: 1})
	if err != nil || got != "1\tshort\n" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if reader.read > 2*readBufSize {
		t.Fatalf("consumed %d bytes", reader.read)
	}
}

type cancelReader struct {
	cancel context.CancelFunc
	reads  int
}

func (c *cancelReader) Read(p []byte) (int, error) {
	c.reads++
	if c.reads == 3 {
		c.cancel()
	}
	if c.reads > 1000 {
		return 0, io.EOF
	}
	return copy(p, strings.Repeat("a", len(p))), nil
}

func TestReadCancelsInsideLongLine(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	reader := &cancelReader{cancel: cancel}
	_, err := readLines(ctx, reader, readArgs{})
	if err != context.Canceled || reader.reads > 10 {
		t.Fatalf("err=%v reads=%d", err, reader.reads)
	}
}

func TestReadCountsRuneSplitAcrossChunks(t *testing.T) {
	path := writeTempFile(t, strings.Repeat("€", 30000)+"\n")
	got, err := runRead(t, map[string]any{"path": path})
	if err != nil || !strings.Contains(got, "[line truncated, 30000 chars]") {
		t.Fatalf("got tail %q err=%v", got[len(got)-40:], err)
	}
}
