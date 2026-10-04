package tools

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"os"
)

func readPicture(file *os.File, size int64, path string, knownImage bool) (string, []Image, error) {
	refusal := errors.New(path + " is a binary file; read only handles text and pictures")
	if size > maxImageFile {
		if knownImage {
			return "", nil, fmt.Errorf("%s is an image file of %d MB; the limit is %d MB", path, size>>20, maxImageFile>>20)
		}
		return "", nil, refusal
	}
	data, err := io.ReadAll(io.NewSectionReader(file, 0, size))
	if err != nil {
		return "", nil, err
	}
	picture, err := loadImage(data)
	switch {
	case err == nil:
		return "Read image file [" + picture.MimeType + "]", []Image{picture}, nil
	case errors.Is(err, errNotImage) && knownImage:
		return "", nil, errors.New(path + " is a damaged image: its header is cut off or corrupt")
	case errors.Is(err, errNotImage):
		return "", nil, refusal
	}
	return "", nil, errors.New(path + ": " + err.Error())
}

var imageMagics = []string{"\x89PNG\r\n\x1a\n", "\xff\xd8\xff", "GIF87a", "GIF89a"}

// hasImageHeader recognises a picture by its decodable config or, for a cut
// file, by the magic bytes of the formats that are sent as they are.
func hasImageHeader(head []byte) bool {
	if _, _, err := image.DecodeConfig(bytes.NewReader(head)); err == nil {
		return true
	}
	for _, magic := range imageMagics {
		if bytes.HasPrefix(head, []byte(magic)) {
			return true
		}
	}
	return false
}
