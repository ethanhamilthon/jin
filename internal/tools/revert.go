package tools

import (
	"errors"
	"os"
)

// RevertResult tells which files Revert restored and which it left alone
// because they changed again after the agent wrote them.
type RevertResult struct {
	Restored, Skipped []string
}

// Revert undoes changes, given in the order they were made. A file is
// restored only when its content is still what the agent wrote last;
// otherwise it is skipped so the user's later edits are not lost.
func Revert(changes []Change) (RevertResult, error) {
	first := map[string]Change{}
	last := map[string]Change{}
	var order []string
	for _, c := range changes {
		if _, ok := first[c.Path]; !ok {
			first[c.Path] = c
			order = append(order, c.Path)
		}
		last[c.Path] = c
	}
	var result RevertResult
	for _, path := range order {
		current, exists, err := currentContent(path)
		if err != nil {
			return result, err
		}
		if !exists || current != last[path].After {
			result.Skipped = append(result.Skipped, path)
			continue
		}
		if err := restore(path, first[path]); err != nil {
			return result, err
		}
		result.Restored = append(result.Restored, path)
	}
	return result, nil
}

func restore(path string, original Change) error {
	if !original.Existed {
		return os.Remove(path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, []byte(original.Before), info.Mode()); err != nil {
		return errors.New("cannot restore " + path + ": " + err.Error())
	}
	return nil
}
