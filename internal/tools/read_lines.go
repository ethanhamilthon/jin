package tools

import (
	"bufio"
	"bytes"
	"io"
	"strconv"
	"unicode/utf8"
)

const (
	maxLineChars = 2000
	readBufSize  = 64 << 10
)

type lineScanner struct{ r *bufio.Reader }

// next returns the next line without its newline, cut to maxLineChars, and
// the full length of the line in characters. It never holds more than the
// cut line in memory.
func (s lineScanner) next() (string, int, error) {
	var kept []byte
	keptChars, chars, got := 0, 0, false
	for {
		chunk, err := s.r.ReadSlice('\n')
		got = got || len(chunk) > 0
		data := bytes.TrimSuffix(chunk, []byte("\n"))
		chars += utf8.RuneCount(data)
		for len(data) > 0 && keptChars < maxLineChars {
			_, size := utf8.DecodeRune(data)
			kept = append(kept, data[:size]...)
			data = data[size:]
			keptChars++
		}
		switch err {
		case bufio.ErrBufferFull:
			continue
		case io.EOF:
			if !got {
				return "", 0, io.EOF
			}
			return string(kept), chars, nil
		case nil:
			return string(kept), chars, nil
		}
		return "", 0, err
	}
}

func formatLine(number int, text string, chars int) string {
	row := strconv.Itoa(number) + "\t" + text
	if chars > maxLineChars {
		row += "... [line truncated, " + strconv.Itoa(chars) + " chars]"
	}
	return row + "\n"
}
