package wire

import "strconv"

// Notes that jin adds at the end of a tool result, or in place of one.
const (
	// NoVision follows a tool result whose picture the model cannot see.
	NoVision = "\n[The current model does not support images. The image is omitted from this request.]"
	// TimedOut ends the output of a command that was killed.
	TimedOut = "\n[command timed out]"
	// NoOutput is the result of a command that printed nothing.
	NoOutput = "[command completed with no output]"
	// OmittedPrefix starts the note that replaces an old tool result.
	OmittedPrefix = "[output of "
)

// Omitted replaces an old tool result of size bytes.
func Omitted(tool string, size int) string {
	if tool == "" {
		tool = "tool"
	}
	return OmittedPrefix + tool + " omitted, " + strconv.Itoa(size) + " bytes; run it again if needed]"
}

// ExitCode ends the output of a command that exited non-zero: often an answer
// (grep found nothing), not a failure.
func ExitCode(code int) string { return "[exit code: " + strconv.Itoa(code) + "]" }

// CommandFailed ends the output of a command that could not run.
func CommandFailed(reason string) string { return "[command failed: " + reason + "]" }

// ReadTruncated ends a read that hit the output limit at line last.
func ReadTruncated(last int) string {
	return "[truncated at line " + strconv.Itoa(last) + "; continue with offset=" + strconv.Itoa(last+1) + "]"
}

// LogTruncated ends the start of a log that keeps growing in path.
func LogTruncated(limit int, path string) string {
	return "\n[output truncated after " + strconv.Itoa(limit) + " bytes; the full output keeps growing in " + path +
		": read it with the read tool or check the task]"
}

// OutputCut marks the middle that was cut from a long output kept in path.
func OutputCut(cutBytes int64, cutLines int, total int64, lines int, path string) string {
	note := "\n[output truncated: " + strconv.FormatInt(cutBytes, 10) + " bytes"
	if cutLines > 0 {
		note += " (about " + strconv.Itoa(cutLines) + " lines)"
	}
	return note + " cut from the middle of " + strconv.FormatInt(total, 10) + " bytes, " + strconv.Itoa(lines) +
		" lines. The full output is in " + path + "; read it with the read tool (offset and limit) instead of running the command again]\n"
}

// Lines of the note before an answer ends while background tasks still run.
const (
	TasksStillRunning = "Still running:"
	TasksEndWithRun   = "They are stopped when you finish. If you need a result, check the task before you finish; otherwise reply in one short line."
	TasksKeepRunning  = "They keep running after you finish, and their results arrive later as messages. Stop the ones that are no longer needed with the task tool, then reply in one short line. If they all should keep running, reply in one short line saying so."
)
