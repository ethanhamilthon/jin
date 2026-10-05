package ui

// show opens sel on top of the open panel, so Esc in sel comes back to it.
// A panel that opens again (a flow returning to its list) takes the place of
// its earlier copy instead of stacking on top of the steps after it.
func (a *app) show(sel *selector) {
	sel.back = a.sel
	for open := a.sel; open != nil; open = open.back {
		if open.title == sel.title {
			sel.back = open.back
			break
		}
	}
	a.sel = sel
}
