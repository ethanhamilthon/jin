package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

// shownCluster is what the input draws for one element: a token's label, or
// nothing for a marker that is not known.
func shownCluster(cluster string) string {
	if !isTokenMark(cluster) {
		return cluster
	}
	t, _ := tokenOf(cluster)
	return t.label
}

func clusterWidth(cluster string) int {
	return max(1, displaywidth.String(shownCluster(cluster)))
}

func tokenStyle(focused bool) tcell.Style {
	if !focused {
		return muted
	}
	return accent
}
