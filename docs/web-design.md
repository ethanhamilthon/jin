# jin web: design system

Adapted from a style reference the user supplied (dark editorial terminal: near-black
surfaces, hairline borders, one chromatic accent, a light serif against a tight sans and a
mono). Changes for jin: the accent is blue by default and set by the user, and the system
is fitted to a working app instead of a marketing page.

## Principles

- Dark, near-black, monochrome. One accent color does all the signaling.
- Depth comes from surface steps and 1px borders, never from drop shadows. The only glow
  belongs to the accent.
- Near-sharp shapes: cards 0px, buttons and code 4px, tags 2px, pills 9999px.
- Text carries the interface; no illustrations except the onboarding dotted mark.

## Accent

`--accent` defaults to `#52A8FF`, the accent of the TUI theme Jin Original. The user changes
it in settings (presets plus a custom color). Everything else derives from it:

```css
--accent: #52a8ff;
--accent-glow: color-mix(in srgb, var(--accent) 45%, transparent);
--accent-deep: color-mix(in oklab, var(--accent) 45%, black);
--accent-shade: color-mix(in oklab, var(--accent) 25%, black);
--on-accent: #000000;
```

The accent is used only for: the primary button (filled, `--on-accent` text, glow
`0 0 8px 0 var(--accent-glow)`), focus and active borders with the glow, status pills, the
working indicator of a session, the active item in a list, and at most one italic serif
word in a title. Never for body text.

## Colors

| Token | Value | Use |
| --- | --- | --- |
| `--void` | `#000000` | top bar, code block background |
| `--canvas` | `#060606` | page background |
| `--card` | `#1f1f1f` | messages, panels, sidebar items |
| `--raised` | `#252525` | code blocks, active panels, borders of cards |
| `--hover` | `#313131` | hover, popovers, menus |
| `--line` | `#3d3d3d` | table and row dividers |
| `--line-soft` | `#525252` | disabled outlines |
| `--text-muted` | `#7a7a7a` | helper text, labels |
| `--text-soft` | `#8a8a8a` | captions, metadata |
| `--text-dim` | `#c5c5c5` | tertiary text |
| `--text` | `#e5e5e5` | body |
| `--text-strong` | `#ffffff` | titles, icons |

An agent app needs state colors the reference does not have. They stay muted and are used
only for their meaning: `--ok #62c073` (diff added, success), `--warn #ffb224`
(retries, read-only), `--error #ff6166` (errors, diff removed). Diff lines use them at 12%
background tint, not as text color.

## Type

| Family | Use |
| --- | --- |
| Source Serif 4, weight 300 | session and panel titles, onboarding headline, empty states |
| Inter Tight 400/500/600 | all UI: body, buttons, inputs, lists, labels |
| JetBrains Mono 400/500 | code, tool calls and output, diffs, paths, ids, token counts |

The reference names PT Serif 300, which has no 300 weight; its listed substitute Source
Serif 4 has one. Fonts are bundled with the frontend, no requests to font CDNs.

| Role | Size / line height | Notes |
| --- | --- | --- |
| label | 11px / 1.2 | Inter Tight 500, uppercase, 0.22em tracking, `--text-muted` |
| caption | 12px / 1.4 | metadata, timestamps |
| body | 14px / 1.55 | messages, inputs |
| code | 13px / 1.55 | JetBrains Mono |
| title-sm | 20px / 1.3 | serif, -0.01em |
| title | 32px / 1.15 | serif, -0.4px, empty states and onboarding |
| display | 56px / 0.95 | serif, -0.025em, onboarding only |

Serif never goes below 20px and never into buttons, inputs or labels.

## Labels

Metadata labels are uppercase Inter Tight 11px with 0.22em tracking in `--text-muted`, in
literal brackets: `[ BASH ]`, `[ EDIT · src/main.go ]`, `[ READ-ONLY ]`. Sections may carry
a right-aligned index in the same style.

## Spacing and layout

Spacing scale: 4, 6, 8, 10, 12, 14, 16, 20, 24, 28, 32, 44 px.

The app fills the window; it is not a centered page. The message column is at most
760px wide and centered in its pane. Card padding inside the app is 16-24px, not the
reference's 32-48px; gaps between elements 12-16px.

## Components

- Primary button: accent fill, `--on-accent` text, 14px Inter Tight 500, 0.04em tracking,
  padding 10px 18px, radius 4px, accent glow.
- Secondary button: transparent, 1px accent border, accent text, same glow.
- Ghost button: transparent, `--text-strong`, uppercase 12px 0.18em; active turns accent.
- Card: `--card`, 1px `--raised` border, radius 0.
- Code block: `--void`, 1px `--raised` border, radius 4px, a header bar (`--card`) with the language and Copy, mono 13px / 1.6, no wrapping (it scrolls sideways).
- Status pill: transparent, 1px accent border, accent text, 11px uppercase 0.06em,
  padding 4px 10px, dot or check prefix.
- Input: `--card`, 1px `--raised` border; focused it takes the accent border and glow.
- Top bar: `--void`, 1px `--raised` bottom border, 48px high.

## Markdown

Answers and the Markdown preview of Files use one stylesheet (`web/src/styles/markdown*.css`).

- Body text is `--text-dim` at 14.5px / 1.7, 14px between blocks. Bold, italic and headings
  are `--text-strong`; links are the accent with a thin underline.
- Headings h1 to h3 are serif 300 at 30, 23 and 18px; h1 and h2 carry a hairline rule. h4 is
  600 Inter Tight 14.5px. h5 and h6 are 11px labels with 0.22em tracking.
- Inline code has a 20% accent fill, light accent text and a 35% accent hairline. Lists have
  muted markers. A quote is a `--card` block with a 2px accent bar.
- Tables have no vertical lines; the header is a small uppercase label, rows highlight on
  hover.

## Do not

- Add chromatic colors beyond the accent and the three state colors.
- Put shadows on cards, panels or dialogs.
- Round cards or panels beyond 4px.
- Use the accent for body text or full-size icons.
- Use photos, gradients or multicolor icons. Icons are line icons in `--text-strong` or
  the accent.
