package store

import "strings"

const (
	keyVoiceURL      = "voice.base_url"
	keyVoiceKey      = "voice.api_key"
	keyVoiceModel    = "voice.model"
	keyVoiceLanguage = "voice.language"
)

// Voice is the speech-to-text provider of /voice: any OpenAI-compatible
// /audio/transcriptions endpoint. An empty Language lets the model detect it.
type Voice struct {
	BaseURL  string
	APIKey   string
	Model    string
	Language string
}

func (v Voice) Ready() bool {
	return v.BaseURL != "" && v.APIKey != "" && v.Model != ""
}

func parseVoice(values map[string]string) Voice {
	return Voice{
		BaseURL:  strings.TrimSpace(values[keyVoiceURL]),
		APIKey:   strings.TrimSpace(values[keyVoiceKey]),
		Model:    strings.TrimSpace(values[keyVoiceModel]),
		Language: strings.TrimSpace(values[keyVoiceLanguage]),
	}
}

func (db *DB) SaveVoice(v Voice) error {
	return db.setSettings(map[string]string{
		keyVoiceURL: v.BaseURL, keyVoiceKey: v.APIKey,
		keyVoiceModel: v.Model, keyVoiceLanguage: v.Language,
	})
}
