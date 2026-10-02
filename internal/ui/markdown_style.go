package ui

import "github.com/gdamore/tcell/v3"

// Markdown stays quieter than the input box: body text is a notch below
// the input's foreground, and emphasis is carried by color rather than bold.
var (
	bodyStyle  = base.Foreground(colorText)
	codeStyle  = base.Foreground(colorTeal)
	quoteStyle = base.Foreground(colorPurple)
)

// headingStyle paints a heading as a full-width band of its level's color
// with dark text on it, so headings stand out while scrolling.
func headingStyle(level int) tcell.Style {
	band := colorPink
	switch level {
	case 1:
		band = colorBlueFG
	case 2:
		band = colorTeal
	case 3:
		band = colorPurple
	}
	return base.Background(band).Foreground(colorBG).Bold(true)
}

func strongStyle(style tcell.Style) tcell.Style {
	return style.Foreground(colorAmber)
}

func emphasisStyle(style tcell.Style) tcell.Style {
	return style.Italic(true).Foreground(colorPurple)
}

func inlineCodeStyle(style tcell.Style) tcell.Style {
	return style.Foreground(colorTeal).Background(colorRaised)
}

func tableHeaderStyle(style tcell.Style) tcell.Style {
	return style.Foreground(colorBlueFG)
}

// linkStyle marks text as a link: underlined, and an OSC 8 hyperlink for
// terminals that open those themselves.
func linkStyle(url string) tcell.Style {
	return accent.Underline(true).Url(url)
}
