import { describe, expect, it } from "vitest";
import { shows, toolsShown } from "./fold";

describe("shows", () => {
  it("hides tool entries and reasoning only when tools are hidden", () => {
    expect([3, 0, 1, 2].map(toolsShown)).toEqual([true, true, false, false]);
    expect([3, 1].map((m) => shows(m, "tool_result"))).toEqual([true, false]);
    expect([3, 1].map((m) => shows(m, "tool_call"))).toEqual([true, false]);
    expect([3, 1].map((m) => shows(m, "reasoning"))).toEqual([true, false]);
    expect([3, 1, 2].every((m) => shows(m, "assistant") && shows(m, "user"))).toBe(true);
  });

  it("hides background task results when chat details are hidden", () => {
    for (const kind of ["info", "error"] as const) {
      expect([3, 0, 1, 2].map((mode) => shows(mode, kind, "task"))).toEqual([true, true, false, false]);
      expect(shows(1, kind)).toBe(true);
      expect(shows(1, kind, "shell")).toBe(true);
    }
  });
});
