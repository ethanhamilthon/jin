package tools

import (
	"bytes"
	"os"
	"regexp"
	"strings"
)

const grepMaxFile = 2 << 20

type fileHits struct {
	lines   []string
	hits    []int
	skipped string
}

// searchFile finds the lines of one file that match; binary and very big
// files are skipped and say why.
func searchFile(re *regexp.Regexp, abs string) fileHits {
	info, err := os.Stat(abs)
	if err != nil {
		return fileHits{skipped: "unreadable"}
	}
	if info.Size() > grepMaxFile {
		return fileHits{skipped: "over 2 MB"}
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return fileHits{skipped: "unreadable"}
	}
	if bytes.IndexByte(data[:min(len(data), binarySniff)], 0) >= 0 {
		return fileHits{skipped: "binary"}
	}
	found := fileHits{lines: strings.Split(string(data), "\n")}
	for i, line := range found.lines {
		if re.MatchString(line) {
			found.hits = append(found.hits, i)
		}
	}
	return found
}
