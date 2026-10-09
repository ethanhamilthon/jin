import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { every } from "./poll";

describe("every", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.stubGlobal("document", { hidden: false });
  });
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("runs once a second until stopped", async () => {
    const fn = vi.fn();
    const stop = every(fn);
    await vi.advanceTimersByTimeAsync(3000);
    expect(fn).toHaveBeenCalledTimes(3);
    stop();
    await vi.advanceTimersByTimeAsync(3000);
    expect(fn).toHaveBeenCalledTimes(3);
  });

  it("skips a tick while the last run goes on and while the page is hidden", async () => {
    let release = () => {};
    const fn = vi.fn(() => new Promise<void>((resolve) => (release = resolve)));
    const stop = every(fn);
    await vi.advanceTimersByTimeAsync(3000);
    expect(fn).toHaveBeenCalledTimes(1);
    release();
    vi.stubGlobal("document", { hidden: true });
    await vi.advanceTimersByTimeAsync(3000);
    expect(fn).toHaveBeenCalledTimes(1);
    stop();
  });
});
