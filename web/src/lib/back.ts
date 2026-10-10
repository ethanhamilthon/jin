import { app } from "./app.svelte";
import { closePane } from "./actions";
import { mobile } from "./mobile.svelte";

let pushed = false;

// A phone has a back gesture. While a panel hides the chat, one history entry
// stands for it, so the gesture closes it instead of leaving jin.
const layerOpen = () => mobile.on && app.pane.kind !== "chat";

export function syncBackLayer() {
  const open = layerOpen();
  if (open && !pushed) {
    history.pushState({ jin: 1 }, "");
    pushed = true;
  } else if (!open && pushed) {
    pushed = false;
    history.back();
  }
}

export function onBack() {
  if (!pushed) return;
  pushed = false;
  closePane();
}
