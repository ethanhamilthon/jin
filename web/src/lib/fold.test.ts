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
});
