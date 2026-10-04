package tools

import (
	"bufio"
	"context"
	"io"
	"strconv"
	"strings"
)

// readLines streams the requested line range of r, stopping at the output
// budget, and says where it stopped.
func readLines(ctx context.Context, r io.Reader, args readArgs) (string, error) {
	scanner := lineScanner{bufio.NewReaderSize(r, readBufSize)}
	start := max(args.Offset, 1) - 1
	var out strings.Builder
	for n := 1; ; n++ {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		text, chars, err := scanner.next()
		if err == io.EOF {
			if out.Len() == 0 {
				return pastEnd(n-1, args), nil
			}
			return out.String(), nil
		}
		if err != nil {
			return "", err
		}
		if n <= start {
			continue
		}
		if args.Limit > 0 && n-start > args.Limit {
			return out.String(), nil
		}
		row := formatLine(n, text, chars)
		if out.Len()+len(row) > maxReadOutput {
			out.WriteString("[truncated at line " + strconv.Itoa(n-1) + "; continue with offset=" + strconv.Itoa(n) + "]")
			return out.String(), nil
		}
		out.WriteString(row)
	}
}

func pastEnd(total int, args readArgs) string {
	if args.Offset > max(total, 1) {
		return "[file has " + strconv.Itoa(total) + " lines; offset " + strconv.Itoa(args.Offset) + " is past the end]"
	}
	return ""
}
