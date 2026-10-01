package ui

// confirmDelete asks before removing name with remove, then calls back.
func (a *app) confirmDelete(name string, remove func(string) error, back func()) {
	if name == "" {
		return
	}
	options := []option{{label: "No", value: "no"}, {label: "Yes, delete", value: "yes"}}
	a.openList("Delete "+name+"?", options, "no", func(answer string) error {
		if answer == "yes" {
			if err := remove(name); err != nil {
				return err
			}
		}
		back()
		return nil
	})
}
