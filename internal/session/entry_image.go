package session

import (
	"encoding/base64"
	"errors"
)

// EntryImage returns the bytes and MIME type of the picture an entry shows.
func (m *Manager) EntryImage(id string, index int) ([]byte, string, error) {
	var data []byte
	var mime string
	err := m.Do(id, func(s *Session) error {
		if index < 0 || index >= len(s.entries) || s.entries[index].Image == nil {
			return errors.New("no picture at this entry")
		}
		image := s.entries[index].Image
		decoded, err := base64.StdEncoding.DecodeString(image.Data)
		data, mime = decoded, image.MimeType
		return err
	})
	return data, mime, err
}
