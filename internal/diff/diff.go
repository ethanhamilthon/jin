// Package diff finds the changed lines between two texts.
package diff

import "strings"

// Op marks a line as kept, removed or added.
type Op byte

const (
	Keep   Op = ' '
	Remove Op = '-'
	Add    Op = '+'
)

type Line struct {
	Op   Op
	Text string
}

// maxCells bounds the LCS table; a bigger change is shown as everything
// removed and then everything added.
const maxCells = 4_000_000

// Lines returns the removed and added lines between before and after, in
// order, without the lines both share.
func Lines(before, after string) []Line {
	a, b := split(before), split(after)
	for len(a) > 0 && len(b) > 0 && a[0] == b[0] {
		a, b = a[1:], b[1:]
	}
	for len(a) > 0 && len(b) > 0 && a[len(a)-1] == b[len(b)-1] {
		a, b = a[:len(a)-1], b[:len(b)-1]
	}
	if len(a)*len(b) > maxCells {
		return append(mark(Remove, a), mark(Add, b)...)
	}
	return lcs(a, b)
}

func split(text string) []string {
	if text == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(text, "\n"), "\n")
}

func mark(op Op, lines []string) []Line {
	out := make([]Line, len(lines))
	for i, text := range lines {
		out[i] = Line{Op: op, Text: text}
	}
	return out
}

func lcs(a, b []string) []Line {
	width := len(b) + 1
	table := make([]int, (len(a)+1)*width)
	for i := len(a) - 1; i >= 0; i-- {
		for j := len(b) - 1; j >= 0; j-- {
			if a[i] == b[j] {
				table[i*width+j] = table[(i+1)*width+j+1] + 1
			} else {
				table[i*width+j] = max(table[(i+1)*width+j], table[i*width+j+1])
			}
		}
	}
	var out []Line
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch {
		case a[i] == b[j]:
			i, j = i+1, j+1
		case table[(i+1)*width+j] >= table[i*width+j+1]:
			out = append(out, Line{Op: Remove, Text: a[i]})
			i++
		default:
			out = append(out, Line{Op: Add, Text: b[j]})
			j++
		}
	}
	return append(append(out, mark(Remove, a[i:])...), mark(Add, b[j:])...)
}
