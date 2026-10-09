package ui

func (a *app) paneTimelineHeight(w, h int) int {
	if w < 10 || h < 8 || a.active == nil {
		return 0
	}
	statusY := h - statusLines
	box := a.inputBox()
	inputHeight := inputHeight(box.visible(), box.cursor, w, h)
	asking := a.sel == nil && a.active.ask != nil
	if asking {
		inputHeight = min(a.active.ask.height(w), max(2, h/2))
	}
	ruleY := statusY - 1 - inputHeight - 1
	if panel := a.panel(); a.selectorHeight(panel, h) > 0 {
		return max(0, ruleY-a.selectorHeight(panel, h)-1)
	}
	if pinned := a.active.pinnedTodos(); pinned != nil && a.sel == nil {
		rows := a.active.todoBlockRows(todoRows(pinned, w), h)
		return max(0, ruleY-rows-1)
	}
	return max(0, ruleY)
}

func paneRects(root *paneNode, w, h int) map[*paneNode]paneRect {
	out := make(map[*paneNode]paneRect)
	layoutPanes(root, paneRect{w: w, h: h}, out)
	return out
}

func canSplit(r paneRect, vertical bool) bool {
	if r.w < minPaneWidth || r.h < minPaneHeight {
		return false
	}
	if vertical {
		return r.w/2 >= minPaneWidth && r.w-r.w/2 >= minPaneWidth
	}
	return r.h/2 >= minPaneHeight && r.h-r.h/2 >= minPaneHeight
}
