import { afterEach, describe, expect, it, vi } from "vitest";
import { preview, ring, shouldRing } from "./sound";
import type { Sound } from "./types";

const sound: Sound = { enabled: true, only_blur: false, volume: 75, tone: "chime" };

afterEach(() => vi.unstubAllGlobals());

function audioMock() {
  const play = vi.fn().mockResolvedValue(undefined);
  const pause = vi.fn();
  const sources: string[] = [];
  const instances: { volume: number }[] = [];
  vi.stubGlobal("Audio", class {
    volume = 1;
    play = play;
    pause = pause;
    constructor(src: string) { sources.push(src); instances.push(this); }
  });
  return { play, pause, sources, instances };
}

describe("notification sound", () => {
  it("uses one tone for questions, answers, and errors", () => {
    const audio = audioMock();
    for (const event of ["ask", "turn", "error"] as const) ring(sound, true, event);
    expect(audio.sources).toEqual(Array(3).fill("/api/sounds/chime.wav"));
    expect(audio.instances.map((a) => a.volume)).toEqual([.75, .75, .75]);
    expect(audio.pause).toHaveBeenCalledTimes(2);
  });

  it("respects mute, volume, and focus", () => {
    expect(shouldRing(sound, true)).toBe(true);
    expect(shouldRing({ ...sound, enabled: false }, false)).toBe(false);
    expect(shouldRing({ ...sound, volume: 0 }, false)).toBe(false);
    expect(shouldRing({ ...sound, only_blur: true }, true)).toBe(false);
    expect(shouldRing({ ...sound, only_blur: true }, false)).toBe(true);
  });

  it("previews even when notifications are disabled", () => {
    const audio = audioMock();
    preview({ ...sound, enabled: false, only_blur: true });
    expect(audio.play).toHaveBeenCalledOnce();
  });

  it("handles browser autoplay rejection", async () => {
    const audio = audioMock();
    audio.play.mockRejectedValue(new Error("blocked"));
    ring(sound, false);
    await Promise.resolve();
    expect(audio.play).toHaveBeenCalledOnce();
  });
});
