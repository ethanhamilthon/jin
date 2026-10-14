import { afterEach, describe, expect, it, vi } from "vitest";
import { closeDesktopNotice, desktopNotice, fetchDesktopNotice, loadDesktopNotice } from "./desktop-notice.svelte";

function answer(status: number, body?: unknown) {
  vi.stubGlobal("fetch", vi.fn(async () => new Response(body === undefined ? null : JSON.stringify(body), { status })));
}

afterEach(() => {
  vi.unstubAllGlobals();
  desktopNotice.value = null;
});

describe("desktop notice", () => {
  it("shows nothing in a browser (404)", async () => {
    answer(404);
    expect(await fetchDesktopNotice()).toBeNull();
  });

  it("shows nothing when there is no problem (204)", async () => {
    answer(204);
    expect(await fetchDesktopNotice()).toBeNull();
  });

  it("shows nothing when the request fails", async () => {
    vi.stubGlobal("fetch", vi.fn(async () => { throw new Error("offline"); }));
    expect(await fetchDesktopNotice()).toBeNull();
  });

  it("returns a path problem", async () => {
    answer(200, { kind: "path", dir: "~/.local/bin", line: "export X" });
    await loadDesktopNotice();
    expect(desktopNotice.value?.kind).toBe("path");
  });

  it("ignores an unknown kind", async () => {
    answer(200, { kind: "other" });
    expect(await fetchDesktopNotice()).toBeNull();
  });

  it("closes", async () => {
    answer(200, { kind: "link", message: "m" });
    await loadDesktopNotice();
    closeDesktopNotice();
    expect(desktopNotice.value).toBeNull();
  });
});
