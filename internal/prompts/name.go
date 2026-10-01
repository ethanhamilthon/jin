package prompts

import (
	"errors"
	"path/filepath"
	"strings"
	"unicode"
)

var errBadName = errors.New("name may use letters, digits, - _ . and / for folders")

// IsNameRune reports whether r may appear in a prompt name.
func IsNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || strings.ContainsRune("-_./", r)
}

func validName(name string) bool {
	if name == "" {
		return false
	}
	for _, segment := range strings.Split(name, "/") {
		if segment == "" || strings.HasPrefix(segment, ".") {
			return false
		}
		for _, r := range segment {
			if !IsNameRune(r) {
				return false
			}
		}
	}
	return true
}

// Path resolves a prompt name to its file, rejecting names that could escape
// the prompts directory.
func Path(name string) (string, error) {
	name = strings.TrimSuffix(strings.TrimSpace(name), ext)
	if !validName(name) {
		return "", errBadName
	}
	dir, err := root()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filepath.FromSlash(name)+ext), nil
}
