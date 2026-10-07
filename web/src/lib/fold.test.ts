import { describe, expect, it } from "vitest";
import { shows } from "./fold";

describe("shows", () => {
  it("follows the TUI modes", () => {
    expect([3, 0, 1, 2].map((m) => shows(m, "tool_result"))).toEqual([true, false, false, false]);
    expect([3, 0, 1, 2].map((m) => shows(m, "tool_call"))).toEqual([true, true, false, false]);
    expect([3, 0, 1, 2].map((m) => shows(m, "reasoning"))).toEqual([true, true, true, false]);
    expect([3, 0, 1, 2].every((m) => shows(m, "assistant"))).toBe(true);
  });
});
