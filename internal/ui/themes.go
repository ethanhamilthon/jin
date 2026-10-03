package ui

import "github.com/gdamore/tcell/v3/color"

// theme is a palette. Every theme sets the same slots; a theme changes
// colors only, never the layout.
type theme struct {
	name string
	// bg, fg: the page; text: markdown body; muted, argument, detail, dim:
	// quieter greys; border: rules; raised: user messages and inline code.
	bg, fg, text, muted, argument, detail, dim, border, raised uint32
	// status is the status bar; onStatus and statusTitle are its text.
	status, onStatus, statusTitle uint32
	// accent is links and the cursor; the rest are the named hues.
	accent, green, amber, red, purple, pink, teal uint32
}

func rgb(v uint32) color.Color { return color.NewHexColor(int32(v)) }

func applyTheme(t theme) {
	colorBG, colorFG, colorText, colorMuted = rgb(t.bg), rgb(t.fg), rgb(t.text), rgb(t.muted)
	colorArgument, colorDetail, colorDim = rgb(t.argument), rgb(t.detail), rgb(t.dim)
	colorBorder, colorRaised = rgb(t.border), rgb(t.raised)
	colorBlue, colorOnBlue, colorWhite = rgb(t.status), rgb(t.onStatus), rgb(t.statusTitle)
	colorBlueFG, colorGreen, colorAmber, colorRed = rgb(t.accent), rgb(t.green), rgb(t.amber), rgb(t.red)
	colorPurple, colorPink, colorTeal = rgb(t.purple), rgb(t.pink), rgb(t.teal)
	rebuildStyles()
}

// themes are the built-in palettes; the first is the default.
var themes = []theme{
	{name: "Jin Original", bg: 0x000000, fg: 0xEDEDED, text: 0xB8B8B8, muted: 0xA1A1A1, argument: 0x6E6E6E, detail: 0x4F4F4F, dim: 0x666666, border: 0x2E2E2E, raised: 0x1A1A1A,
		status: 0x0070F3, onStatus: 0xCCE3FF, statusTitle: 0xFFFFFF,
		accent: 0x52A8FF, green: 0x62C073, amber: 0xFFB224, red: 0xFF6166, purple: 0xBF7AF0, pink: 0xF75F8F, teal: 0x0AC7B4},
	{name: "Tokyo Night", bg: 0x1A1B26, fg: 0xC0CAF5, text: 0xA9B1D6, muted: 0x9AA5CE, argument: 0x737AA2, detail: 0x565F89, dim: 0x565F89, border: 0x292E42, raised: 0x24283B,
		status: 0x3D59A1, onStatus: 0xC0CAF5, statusTitle: 0xFFFFFF,
		accent: 0x7AA2F7, green: 0x9ECE6A, amber: 0xE0AF68, red: 0xF7768E, purple: 0xBB9AF7, pink: 0xFF9E64, teal: 0x73DACA},
	{name: "Catppuccin Mocha", bg: 0x1E1E2E, fg: 0xCDD6F4, text: 0xBAC2DE, muted: 0xA6ADC8, argument: 0x7F849C, detail: 0x585B70, dim: 0x6C7086, border: 0x45475A, raised: 0x313244,
		status: 0x1E66F5, onStatus: 0xDCE0E8, statusTitle: 0xFFFFFF,
		accent: 0x89B4FA, green: 0xA6E3A1, amber: 0xF9E2AF, red: 0xF38BA8, purple: 0xCBA6F7, pink: 0xF5C2E7, teal: 0x94E2D5},
	{name: "Gruvbox Dark", bg: 0x282828, fg: 0xEBDBB2, text: 0xD5C4A1, muted: 0xBDAE93, argument: 0x928374, detail: 0x665C54, dim: 0x7C6F64, border: 0x3C3836, raised: 0x32302F,
		status: 0x076678, onStatus: 0xEBDBB2, statusTitle: 0xFBF1C7,
		accent: 0x83A598, green: 0xB8BB26, amber: 0xFABD2F, red: 0xFB4934, purple: 0xD3869B, pink: 0xFE8019, teal: 0x8EC07C},
	{name: "Nord", bg: 0x2E3440, fg: 0xECEFF4, text: 0xD8DEE9, muted: 0xC5CCD8, argument: 0x7B88A1, detail: 0x616E88, dim: 0x677691, border: 0x434C5E, raised: 0x3B4252,
		status: 0x5E81AC, onStatus: 0xECEFF4, statusTitle: 0xFFFFFF,
		accent: 0x88C0D0, green: 0xA3BE8C, amber: 0xEBCB8B, red: 0xBF616A, purple: 0xB48EAD, pink: 0xD08770, teal: 0x8FBCBB},
	{name: "Dracula", bg: 0x282A36, fg: 0xF8F8F2, text: 0xE2E2DC, muted: 0xBFBFBF, argument: 0x7D84A8, detail: 0x6272A4, dim: 0x6272A4, border: 0x44475A, raised: 0x343746,
		status: 0x6272A4, onStatus: 0xF8F8F2, statusTitle: 0xFFFFFF,
		accent: 0x8BE9FD, green: 0x50FA7B, amber: 0xF1FA8C, red: 0xFF5555, purple: 0xBD93F9, pink: 0xFF79C6, teal: 0x80FFEA},
	{name: "One Dark", bg: 0x282C34, fg: 0xDCDFE4, text: 0xABB2BF, muted: 0x9DA5B4, argument: 0x7F848E, detail: 0x5C6370, dim: 0x5C6370, border: 0x3E4451, raised: 0x2C313A,
		status: 0x3E6BD6, onStatus: 0xDBE6FF, statusTitle: 0xFFFFFF,
		accent: 0x61AFEF, green: 0x98C379, amber: 0xE5C07B, red: 0xE06C75, purple: 0xC678DD, pink: 0xD19A66, teal: 0x56B6C2},
	{name: "Rosé Pine", bg: 0x191724, fg: 0xE0DEF4, text: 0xCFCBE6, muted: 0x908CAA, argument: 0x6E6A86, detail: 0x524F67, dim: 0x6E6A86, border: 0x26233A, raised: 0x1F1D2E,
		status: 0x31748F, onStatus: 0xE0DEF4, statusTitle: 0xFFFFFF,
		accent: 0x9CCFD8, green: 0xA6D189, amber: 0xF6C177, red: 0xEB6F92, purple: 0xC4A7E7, pink: 0xEBBCBA, teal: 0x3E8FB0},
	{name: "Solarized Light", bg: 0xFDF6E3, fg: 0x073642, text: 0x586E75, muted: 0x657B83, argument: 0x93A1A1, detail: 0xB4BCB4, dim: 0x93A1A1, border: 0xDDD6C1, raised: 0xEEE8D5,
		status: 0x268BD2, onStatus: 0xFDF6E3, statusTitle: 0xFFFFFF,
		accent: 0x268BD2, green: 0x859900, amber: 0xB58900, red: 0xDC322F, purple: 0x6C71C4, pink: 0xD33682, teal: 0x2AA198},
	{name: "GitHub Light", bg: 0xFFFFFF, fg: 0x1F2328, text: 0x31363C, muted: 0x59636E, argument: 0x818B98, detail: 0xAFB8C1, dim: 0x8C959F, border: 0xD1D9E0, raised: 0xF6F8FA,
		status: 0x0969DA, onStatus: 0xDDF4FF, statusTitle: 0xFFFFFF,
		accent: 0x0969DA, green: 0x1A7F37, amber: 0x9A6700, red: 0xCF222E, purple: 0x8250DF, pink: 0xBF3989, teal: 0x1B7C83},
}

// themeByName returns the theme called name, or the default.
func themeByName(name string) theme {
	for _, t := range themes {
		if t.name == name {
			return t
		}
	}
	return themes[0]
}
