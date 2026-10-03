package ui

// wordBreaks returns the index where each row starts when items of the
// given widths are laid out in rows of width cells. Rows break after the last
// space that fits, so words stay whole; a word longer than a row is split
// between graphemes. A space that does not fit starts the next row.
func wordBreaks(widths []int, spaces []bool, width int) []int {
	width = max(1, width)
	starts := []int{0}
	used, lastSpace := 0, -1
	for i, w := range widths {
		if used > 0 && used+w > width {
			switch {
			case spaces[i] || lastSpace < 0:
				starts = append(starts, i)
				used = 0
			default:
				start := lastSpace + 1
				starts = append(starts, start)
				used = 0
				for _, prev := range widths[start:i] {
					used += prev
				}
				if used > 0 && used+w > width {
					starts = append(starts, i)
					used = 0
				}
			}
			lastSpace = -1
		}
		if spaces[i] {
			lastSpace = i
		}
		used += w
	}
	return starts
}

// rowBounds turns row starts into [start, end) pairs. With trim set, the
// spaces at a row break are dropped: they were the break.
func rowBounds(starts []int, total int, spaces []bool, trim bool) [][2]int {
	bounds := make([][2]int, len(starts))
	for i, start := range starts {
		end := total
		if i+1 < len(starts) {
			end = starts[i+1]
		}
		for trim && i > 0 && start < end && spaces[start] {
			start++
		}
		for trim && i+1 < len(starts) && end > start && spaces[end-1] {
			end--
		}
		bounds[i] = [2]int{start, end}
	}
	return bounds
}

func isSpace(cluster string) bool { return cluster == " " || cluster == "\u00a0" }
