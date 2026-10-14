import { beforeEach, describe, expect, it } from "vitest";
import { app } from "./app.svelte";
import { closePane, split } from "./actions";
import { keys } from "./keys";
import { openPanel } from "./panels";
import { blockedByZoom, toggleZoom, zoom, zoomed } from "./zoom.svelte";

function reset(chats: number) {
  app.panes = Array.from({ length: chats }, (_, i) => ({ key: i + 1, kind: "chat" as const, session: "s" + (i + 1) }));
  app.nextKey = chats + 1;
  app.focused = 0;
  zoom.on = false;
}

const alt = (key: string) => ({ key, code: "Digit" + key, altKey: true, metaKey: false, ctrlKey: false, preventDefault() {} }) as unknown as KeyboardEvent;

describe("zoom", () => {
  beforeEach(() => reset(2));

  it("does nothing with one pane", () => {
    reset(1);
    toggleZoom();
    expect(zoomed()).toBe(false);
  });

  it("turns on and off", () => {
    toggleZoom();
    expect(zoomed()).toBe(true);
    toggleZoom();
    expect(zoomed()).toBe(false);
  });

  it("is off when only one pane is left", () => {
    toggleZoom();
    app.panes.pop();
    expect(zoomed()).toBe(false);
  });

  it("blocks pane focus keys", () => {
    toggleZoom();
    keys(alt("2"));
    expect(app.focused).toBe(0);
  });

  it("blocks split and close", () => {
    toggleZoom();
    split();
    closePane(1);
    expect(app.panes).toHaveLength(2);
    expect(blockedByZoom()).toBe(true);
  });

  it("is left when a panel opens", () => {
    toggleZoom();
    openPanel("settings");
    expect(zoom.on).toBe(false);
    expect(app.focused).toBe(2);
  });
});
