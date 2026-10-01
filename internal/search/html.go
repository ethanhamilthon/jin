package search

import (
	"strings"

	"golang.org/x/net/html"
)

func hasClass(n *html.Node, class string) bool {
	if n.Type != html.ElementNode {
		return false
	}
	for _, c := range strings.Fields(attr(n, "class")) {
		if c == class {
			return true
		}
	}
	return false
}

func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func textOf(n *html.Node) string {
	var b strings.Builder
	for d := range n.Descendants() {
		if d.Type == html.TextNode {
			b.WriteString(d.Data)
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func stripTags(fragment string) string {
	nodes, err := html.ParseFragment(strings.NewReader(fragment), &html.Node{Type: html.ElementNode, Data: "div"})
	if err != nil {
		return fragment
	}
	var b strings.Builder
	for _, n := range nodes {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		} else {
			b.WriteString(textOf(n))
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}
