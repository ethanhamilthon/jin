package ui

type paneNode struct {
	vertical bool
	first    *paneNode
	second   *paneNode
	session  *chatSession
}

type paneRect struct{ x, y, w, h int }

const (
	maxPanes      = 4
	minPaneWidth  = 16
	minPaneHeight = 5
)

func (n *paneNode) leaf() bool { return n != nil && n.first == nil && n.second == nil }

func (a *app) initPanes() {
	if a.panes == nil && a.active != nil {
		a.panes = &paneNode{session: a.active}
		a.focused = a.panes
	}
}

func paneLeaves(n *paneNode) []*paneNode {
	if n == nil {
		return nil
	}
	if n.leaf() {
		return []*paneNode{n}
	}
	return append(paneLeaves(n.first), paneLeaves(n.second)...)
}

func (a *app) focusedLeaf() *paneNode {
	if a.focused != nil && a.focused.leaf() {
		return a.focused
	}
	return nil
}

func (a *app) paneForSession(s *chatSession) *paneNode {
	for _, leaf := range paneLeaves(a.panes) {
		if leaf.session == s {
			return leaf
		}
	}
	return nil
}

func layoutPanes(n *paneNode, r paneRect, out map[*paneNode]paneRect) {
	if n == nil {
		return
	}
	out[n] = r
	if n.leaf() {
		return
	}
	if n.vertical {
		left := r.w / 2
		layoutPanes(n.first, paneRect{r.x, r.y, left, r.h}, out)
		layoutPanes(n.second, paneRect{r.x + left, r.y, r.w - left, r.h}, out)
		return
	}
	top := r.h / 2
	layoutPanes(n.first, paneRect{r.x, r.y, r.w, top}, out)
	layoutPanes(n.second, paneRect{r.x, r.y + top, r.w, r.h - top}, out)
}

func paneArea(a *app, w, h int) paneRect {
	return paneRect{w: w, h: a.paneTimelineHeight(w, h)}
}

func (r paneRect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}
