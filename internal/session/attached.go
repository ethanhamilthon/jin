package session

import "strings"

// File is a file attached to a message: Name is shown, Path is where it lies.
type File struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// addAttached lists uploaded files among the attached paths for the model and
// by name in the text the chat shows.
func addAttached(paths []string, shown string, attached []File) ([]string, string) {
	if len(attached) == 0 {
		return paths, shown
	}
	var names []string
	for _, file := range attached {
		paths = append(paths, file.Path)
		names = append(names, "[file: "+file.Name+"]")
	}
	return paths, strings.TrimSpace(shown + "\n" + strings.Join(names, "\n"))
}
