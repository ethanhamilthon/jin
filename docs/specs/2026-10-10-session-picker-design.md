# Session picker instead of the web sidebar

Status: approved in chat. Implemented on branch `web/session-picker` in the
worktree `../jin-ui`. The sidebar of `jin web` is removed and a picker dialog
replaces it. This closes TODO item 2 (resizable sidebar) and moves item 3
(grouping sessions by recency) into the picker.

## Decisions

- The project and session sidebar is deleted. The workspace gets its width back.
- Every chat pane header shows a `switch` icon before the session name. Clicking
  it opens the picker for that pane.
- The picker has two steps: project, then the sessions of that project, with a
  back control. Projects keep the order of `GET /api/projects`.
- Search matches session titles only.
- "New session" is the first row of the session step. "Add project" is the last
  row of the project step, and the top bar button that toggled the sidebar
  becomes an "Add project" button.
- Choosing a session always opens it in the pane the picker was opened from,
  even when another pane already shows that session.
- Project marks (working, unread, tasks) move to the project rows of the picker.
- Sessions older than a week are not listed (owner decision 2026-10-10).
- Mobile keeps one pane. The picker is the only navigation there: no drawer, no
  scrim, and no history layer for the sidebar.
- Sizes are unchanged: no resizable panes, no resizable sidebar.

## Removed

- `web/src/components/Sidebar.svelte`, `ProjectList.svelte`, `ProjectNode.svelte`
- `web/src/lib/sidebar.svelte.ts` (the `jin.tree` key)
- `app.sidebar`, `saveSidebar`, the `jin.sidebar` key, the sidebar branch of
  `back.ts`, `.shell.collapsed`, the scrim in `App.svelte`
- copy that says "hidden from the sidebar" in `settings/Archived.svelte` and
  `ProjectPane.svelte`

## Kept

- `app.project` semantics, the `/focus` and `/seen` calls, and `askTrust` when a
  session is shown
- the project path button of the pane header (project settings), split, close
- `GET /api/projects` and `GET /api/sessions?path=`; no server change

## Picker contract

`web/src/components/dialogs/SwitchDialog.svelte` renders inside `<Dialog>`.
It is opened with `app.open("switch")` and closes with `app.dialog = null`.
It calls `switchSession(id)` from `lib/actions` on choice; the plumbing adds
that function next to `openSession`.

## Checks

- `npm run check` and `npm run build` in `web/` stay green.
- Panes, the files panel, the project pane, tasks and pairing keep working.
