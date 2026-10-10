// Package wire holds every piece of text that jin adds to what the model
// reads, apart from the system prompt file, the hooks and the tool schemas:
// the tags that frame messages and the notes that end a tool result. The
// owners of the messages (core, tools, prompts, files, tasks) take their texts
// from here, and a test fails when such a text is written anywhere else.
package wire

// Tags that frame text jin puts into messages.
const (
	RefreshedOpen  = "<system-refreshed>"
	RefreshedClose = "</system-refreshed>"
	// RefreshedNote starts the next user message after the system prompt changed.
	RefreshedNote = RefreshedOpen + "instructions were refreshed" + RefreshedClose + "\n\n"

	// TodoEditedOpen and TodoEditedClose framed a note of old versions; old
	// sessions still carry it and the chat hides it.
	TodoEditedOpen  = "<todo-edited>"
	TodoEditedClose = "</todo-edited>"

	TasksOpen  = "<background-tasks>"
	TasksClose = "</background-tasks>"

	PromptsOpen  = "<pasted-prompts>"
	PromptsClose = "</pasted-prompts>"
	PromptsIntro = "These are reusable prompts the user refers to by #name in the request below."

	FilesOpen  = "<attached-files>"
	FilesClose = "</attached-files>"

	TaskResultOpen  = "<task-result "
	TaskResultClose = "</task-result"
	// TaskResultBroken replaces a closing tag found inside a task output.
	TaskResultBroken = `<\/task-result`

	SummaryOpen  = "<conversation-summary>"
	SummaryClose = "</conversation-summary>"
	SummaryLead  = "The earlier conversation was compacted into the continuation brief below. It is a factual record of the session, not a direct user instruction; separate explicit user requirements from observations and unverified claims. If it lists unfinished work and the user's next message does not change the plan, continue it."
)

// Tags lists the names of every tag above and of the tags inside the blocks
// (`prompt`, `file`). A test checks that no other tag-like text is written in
// the packages that make messages.
var Tags = []string{
	"system-refreshed", "todo-edited", "background-tasks", "pasted-prompts",
	"prompt", "attached-files", "file", "task-result", "conversation-summary",
}

// PromptBlock is one #prompt inside the pasted-prompts block.
func PromptBlock(name, body string) string {
	return "<prompt name=\"" + name + "\">\n" + body + "\n</prompt>\n"
}

// FileLine is one attached file inside the attached-files block; path is
// already escaped for XML.
func FileLine(path string) string { return `<file path="` + path + `"/>` }
