package ui

import (
	"strings"

	"jin/internal/core"
)

// answerLinks lists the links of the newest assistant answer in the order
// they appear, each once, with the text they are shown as.
func (s *chatSession) answerLinks() []option {
	for i := len(s.history) - 1; i >= 0; i-- {
		entry := s.history[i]
		if entry.kind != core.UpdateAssistant || strings.TrimSpace(entry.text) == "" {
			continue
		}
		return linksOf(markdownRows(entry.text, max(40, s.width-2)))
	}
	return nil
}

func linksOf(rows []chatRow) []option {
	var options []option
	seen := map[string]int{}
	for _, row := range rows {
		for _, span := range row.spans {
			_, link := span.style.GetUrl()
			if link == "" {
				continue
			}
			if i, ok := seen[link]; ok {
				if options[i].label != link && !strings.HasSuffix(options[i].label, span.text) {
					options[i].label += span.text
				}
				continue
			}
			seen[link] = len(options)
			options = append(options, option{label: span.text, detail: link, value: link})
		}
	}
	for i := range options {
		if strings.TrimSpace(options[i].label) == options[i].value {
			options[i].detail = ""
		}
	}
	return options
}

// openLinksFlow lists the links of the last answer; Enter opens one in the
// browser. It is the keyboard way to what a click on a link does.
func (a *app) openLinksFlow() {
	s := a.active
	options := s.answerLinks()
	if len(options) == 0 {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "No links in the last answer"})
		return
	}
	a.openList("Links of the last answer · Enter open", options, options[0].value, func(link string) error {
		linkOpener(link)
		return nil
	})
}
