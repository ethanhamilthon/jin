package ui

import (
	"os"

	"jin/internal/daemon"
	"jin/internal/files"
	"jin/internal/session"
)

func (a *app) sendBackendDraft(s *chatSession, text string) error {
	home, _ := os.UserHomeDir()
	clean, paths := files.Extract(text, home, s.path)
	var images []session.Image
	clean = renderTokens(clean, func(token inputToken) string {
		if token.kind == tokenImage {
			images = append(images, session.Image{Label: token.label, Path: token.payload})
			return token.label
		}
		return tokenModelText(token)
	})
	if block := files.Block(paths); block != "" {
		clean += "\n\n" + block
	}
	return a.backend.Command(a.ctx, daemon.Command{
		Action: "send-prepared", Session: s.id, Text: clean,
		Shown: renderTokens(text, tokenLabel), Prompts: promptNames(text), Images: images,
	}, nil)
}
