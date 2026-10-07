package tools

import "os"

// LoadImageFile reads a picture from disk the way the read tool does.
func LoadImageFile(path string) (Image, error) {
	info, err := os.Stat(path)
	if err != nil {
		return Image{}, err
	}
	if info.Size() > maxImageFile {
		return Image{}, errNotImage
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Image{}, err
	}
	return loadImage(data)
}
