package tools

type revertStep struct {
	path        string
	first, last Change
	skip        bool
}

// planRevert decides per file, in the order first changed, whether Revert
// restores it or skips it because it changed after the agent wrote it.
func planRevert(changes []Change) ([]revertStep, error) {
	index := map[string]int{}
	var steps []revertStep
	for _, c := range changes {
		i, ok := index[c.Path]
		if !ok {
			i = len(steps)
			index[c.Path] = i
			steps = append(steps, revertStep{path: c.Path, first: c})
		}
		steps[i].last = c
	}
	for i := range steps {
		current, exists, err := currentContent(steps[i].path)
		if err != nil {
			return nil, err
		}
		steps[i].skip = !exists || current != steps[i].last.After
	}
	return steps, nil
}

// PlanRevert reports what Revert would restore and skip, without writing.
func PlanRevert(changes []Change) (RevertResult, error) {
	steps, err := planRevert(changes)
	var result RevertResult
	for _, step := range steps {
		if step.skip {
			result.Skipped = append(result.Skipped, step.path)
		} else {
			result.Restored = append(result.Restored, step.path)
		}
	}
	return result, err
}
