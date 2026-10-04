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
	var result RevertResult
	steps, err := planRevert(changes)
	if err != nil {
		return result, err
	}
	for _, step := range steps {
		if step.skip {
			result.Skipped = append(result.Skipped, step.path)
			continue
		}
		if err := restore(step.path, step.first); err != nil {
			return result, err
		}
		result.Restored = append(result.Restored, step.path)
	}
	return result, nil
}

func restore(path string, original Change) error {
	if !original.Existed {
		return os.Remove(path)
	}
	if err := writeFileAtomic(path, []byte(original.Before), 0o644); err != nil {
		return errors.New("cannot restore " + path + ": " + err.Error())
	}
	return nil
}
