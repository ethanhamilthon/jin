import { app } from "./app.svelte";
import { closePane, newSession, split } from "./actions";

// keys handles the shortcuts that work anywhere on the page.
export function keys(event: KeyboardEvent) {
  const mod = event.metaKey || event.ctrlKey;
  if (mod && event.key.toLowerCase() === "k") {
    event.preventDefault();
    app.dialog = app.dialog?.name === "palette" ? null : { name: "palette", session: app.pane.session };
    return;
  }
  if (!event.altKey || mod) return;
  const digit = Number(event.key);
  if (digit >= 1 && digit <= app.panes.length) {
    event.preventDefault();
    app.focused = digit - 1;
    return;
  }
  const actions: Record<string, () => void> = {
    n: () => newSession(app.current?.state.path),
    "\\": split,
    w: () => closePane(),
  };
  const action = actions[event.code.replace("Key", "").toLowerCase()] ?? actions[event.key];
  if (action) {
    event.preventDefault();
    action();
  }
}
