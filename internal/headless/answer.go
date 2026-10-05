package headless

func joinAnswer(o *outcome, content string) string {
	if !o.afterNote || o.answer == "" {
		return content
	}
	o.afterNote = false
	return o.answer + "\n\n" + content
}
