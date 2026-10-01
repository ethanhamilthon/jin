package ui

import "github.com/gdamore/tcell/v3"

// Markdown stays quieter than the input box: body text is a notch below
// the input's foreground, and emphasis is carried by color rather than bold.
var (
	bodyStyle  = base.Foreground(colorText)
	codeStyle  = base.Foreground(colorTeal)
	quoteStyle = base.Foreground(colorPurple)
)

func headingStyle(level int) tcell.Style {
	switch level {
	case 1:
		return base.Foreground(colorBlueFG)
	case 2:
		return base.Foreground(colorTeal)
	case 3:
		return base.Foreground(colorPurple)
	default:
		return base.Foreground(colorPink)
	}
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
