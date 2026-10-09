package tools

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	grepMaxOutput = 32 << 10
	grepMaxLine   = 300
)

// grepResult collects the answer. Past the output limit it keeps counting,
// so the model learns how much it did not see.
type grepResult struct {
	mode    string
	context int
	out     strings.Builder
	matches int
	files   int
	capped  bool
	skipped map[string]int
}

func (r *grepResult) put(text string) {
	if r.capped || r.out.Len()+len(text) > grepMaxOutput {
		r.capped = true
		return
	}
	r.out.WriteString(text)
}

func (r *grepResult) add(shown string, found fileHits) {
	if found.skipped != "" {
		r.skipped[found.skipped]++
		return
	}
	if len(found.hits) == 0 {
		return
	}
	r.files++
	r.matches += len(found.hits)
	switch r.mode {
	case "files":
		r.put(shown + "\n")
	case "count":
		r.put(shown + ":" + strconv.Itoa(len(found.hits)) + "\n")
	default:
		r.lines(shown, found)
	}
}

// lines prints each match with its context; groups that are not next to each
// other are told apart by a line with two dashes, as grep does.
func (r *grepResult) lines(shown string, found fileHits) {
	hit := map[int]bool{}
	for _, i := range found.hits {
		hit[i] = true
	}
	next := 0
	for _, i := range found.hits {
		from, to := max(i-r.context, next), min(i+r.context, len(found.lines)-1)
		if from > to {
			continue
		}
		if r.context > 0 && r.out.Len() > 0 && (from > next || next == 0) {
			r.put("--\n")
		}
		for k := from; k <= to; k++ {
			sep := "-"
			if hit[k] {
				sep = ":"
			}
			r.put(shown + sep + strconv.Itoa(k+1) + sep + cutLine(found.lines[k]) + "\n")
		}
		next = to + 1
	}
}

func cutLine(line string) string {
	line = strings.TrimSuffix(line, "\r")
	if len(line) <= grepMaxLine {
		return line
	}
	end := grepMaxLine
	for end > 0 && !utf8.RuneStart(line[end]) {
		end--
	}
	return line[:end] + "…"
}

func (r *grepResult) String() string {
	var notes []string
	if r.capped {
		notes = append(notes, fmt.Sprintf("output cut at 32 KB; %d matches in %d files in total; narrow the pattern, path or glob", r.matches, r.files))
	}
	for _, reason := range []string{"binary", "over 2 MB", "unreadable"} {
		if n := r.skipped[reason]; n > 0 {
			notes = append(notes, fmt.Sprintf("skipped %d %s files", n, reason))
		}
	}
	text := strings.TrimSuffix(r.out.String(), "\n")
	if r.matches == 0 {
		text = "No matches."
	}
	if len(notes) > 0 {
		text += "\n[" + strings.Join(notes, "; ") + "]"
	}
	return text
}
