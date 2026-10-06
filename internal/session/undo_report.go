package session

import (
	"strings"

	"jin/internal/tools"
)

// UndoScope says what undo covers.
const UndoScope = "/undo restores edit and write changes of the last turn; changes made through bash are not covered."

// UndoReport says what an undo restored, skipped and failed on.
func UndoReport(result tools.RevertResult, left []string, failure error) string {
	var lines []string
	if len(result.Restored) > 0 {
		lines = append(lines, "Undone: "+strings.Join(result.Restored, ", "))
	}
	if len(result.Skipped) > 0 {
		lines = append(lines, "Left as is, changed after the agent wrote them: "+strings.Join(result.Skipped, ", "))
	}
	if failure != nil {
		lines = append(lines, "Undo failed: "+failure.Error(), "Not restored, run /undo again to retry: "+strings.Join(left, ", "))
	}
	return strings.Join(append(lines, UndoScope), "\n")
}
