package ui

import "jin/internal/core"

func (a *app) splitPane(vertical bool) bool {
	a.initPanes()
	leaf := a.focusedLeaf()
	if leaf == nil {
		return false
	}
	if len(paneLeaves(a.panes)) >= maxPanes {
		a.active.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Cannot split: four panes are already open"})
		return false
	}
	w, h := a.screen.Size()
	r := paneRects(a.panes, w, paneArea(a, w, h).h)[leaf]
	if !canSplit(r, vertical) || leaf.session == nil {
		a.active.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Cannot split: the terminal is too small"})
		return false
	}
	old := leaf.session
	if err := old.projectError(); err != nil {
		a.report(err)
		return false
	}
	dir := old.path
	if dir == "" {
		dir = a.dir
	}
	s := a.startSessionAt(dir, newSessionID(), a.cfg.ActiveProvider, a.cfg.Model, a.cfg.Effort, nil, a.introEntriesAt(dir))
	first, second := &paneNode{session: old}, &paneNode{session: s}
	*leaf = paneNode{vertical: vertical, first: first, second: second}
	a.focused = second
	a.focus(s)
	return true
}

func (a *app) closePane() {
	if a.panes == nil || len(paneLeaves(a.panes)) <= 1 {
		a.requestQuit()
		return
	}
	target := a.focusedLeaf()
	if target == nil {
		return
	}
	root, removed := removePane(a.panes, target)
	if !removed {
		return
	}
	a.panes = root
	remaining := firstPaneLeaf(root)
	a.focused = remaining
	if remaining != nil {
		a.focus(remaining.session)
	}
}

func removePane(n, target *paneNode) (*paneNode, bool) {
	if n == nil || n.leaf() {
		return n, false
	}
	if n.first == target {
		return n.second, true
	}
	if n.second == target {
		return n.first, true
	}
	if child, ok := removePane(n.first, target); ok {
		n.first = child
		return n, true
	}
	if child, ok := removePane(n.second, target); ok {
		n.second = child
		return n, true
	}
	return n, false
}

func firstPaneLeaf(n *paneNode) *paneNode {
	if n == nil || n.leaf() {
		return n
	}
	if leaf := firstPaneLeaf(n.first); leaf != nil {
		return leaf
	}
	return firstPaneLeaf(n.second)
}
