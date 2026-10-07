package session

import (
	"encoding/base64"
	"errors"

	"jin/internal/provider"
)

// EntryImage returns the bytes and MIME type of the picture an entry shows.
func (m *Manager) EntryImage(id string, index, n int) ([]byte, string, error) {
	var data []byte
	var mime string
	err := m.Do(id, func(s *Session) error {
		image := s.pictureAt(index, n)
		if image == nil {
			return errors.New("no picture at this entry")
		}
		decoded, err := base64.StdEncoding.DecodeString(image.Data)
		data, mime = decoded, image.MimeType
		return err
	})
	return data, mime, err
}

func (s *Session) pictureAt(index, n int) *provider.Image {
	if index < 0 || index >= len(s.entries) {
		return nil
	}
	entry := s.entries[index]
	if entry.Image != nil && n == 0 {
		return entry.Image
	}
	if n >= 0 && n < len(entry.Images) {
		return &entry.Images[n]
	}
	return nil
}
