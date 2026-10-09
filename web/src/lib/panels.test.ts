import { beforeEach, describe, expect, it } from "vitest";
import { app } from "./app.svelte";
import { closeSection, openPanel } from "./panels";

function reset(chats: number) {
  app.panes = Array.from({ length: chats }, (_, i) => ({ key: i + 1, kind: "chat" as const, session: "s" + (i + 1) }));
  app.nextKey = chats + 1;
  app.focused = 0;
  app.lastChat = 1;
}

describe("openPanel", () => {
  beforeEach(() => reset(1));

  it("adds a pane and focuses it", () => {
    openPanel("settings");
    expect(app.panes.map((p) => p.kind)).toEqual(["chat", "settings"]);
    expect(app.focused).toBe(1);
  });

  it("focuses the pane that is already open instead of adding another", () => {
    openPanel("settings");
    app.focused = 0;
    openPanel("settings", "tools");
    expect(app.panes).toHaveLength(2);
    expect(app.focused).toBe(1);
    expect(app.panes[1].section).toBe("tools");
  });

  it("replaces the focused pane when four are open", () => {
    reset(4);
    app.focused = 2;
    openPanel("settings");
    expect(app.panes).toHaveLength(4);
    expect(app.panes[2].kind).toBe("settings");
    expect(app.focused).toBe(2);
  });

  it("sends chat actions to the last focused chat while a panel is focused", () => {
    reset(2);
    app.focused = 1;
    openPanel("settings");
    expect(app.chat?.session).toBe("s2");
    expect(app.chatIndex).toBe(1);
  });

  it("goes back to the list of sections", () => {
    openPanel("settings", "hooks");
    closeSection();
    expect(app.panes[1].section).toBeUndefined();
  });
});
