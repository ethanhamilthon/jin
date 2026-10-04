package ui

import (
	"errors"
	"strings"

	"jin/internal/provider"
	"jin/internal/store"
)

// openVoiceProvider asks for the speech-to-text endpoint of /voice: base URL,
// key, model and language. An empty key keeps the saved one. done runs after
// the settings are saved.
func (a *app) openVoiceProvider(done func()) {
	old := a.cfg.Voice
	a.openField("Voice base URL", firstOf(old.BaseURL, "https://api.groq.com/openai/v1"), false, func(baseURL string) error {
		baseURL = provider.NormalizeBaseURL(baseURL)
		if err := (provider.Config{BaseURL: baseURL, APIKey: "pending"}).Validate(); err != nil {
			return err
		}
		a.openField("Voice API key · empty keeps the saved one", "", true, func(key string) error {
			key = firstOf(strings.TrimSpace(key), old.APIKey)
			if key == "" {
				return errors.New("API key is required")
			}
			a.openField("Voice model", firstOf(old.Model, "whisper-large-v3-turbo"), false, func(model string) error {
				if strings.TrimSpace(model) == "" {
					return errors.New("model is required")
				}
				a.openField("Voice language · ISO code like en, ru, kk · empty detects it", old.Language, false, func(lang string) error {
					v := store.Voice{BaseURL: baseURL, APIKey: key, Model: strings.TrimSpace(model), Language: strings.ToLower(strings.TrimSpace(lang))}
					if err := a.store.SaveVoice(v); err != nil {
						return err
					}
					a.cfg.Voice = v
					if done != nil {
						done()
					}
					return nil
				})
				return nil
			})
			return nil
		})
		return nil
	})
}

func firstOf(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
