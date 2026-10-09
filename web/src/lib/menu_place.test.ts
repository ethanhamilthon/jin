import { describe, expect, it } from "vitest";
import { place } from "./menu_place";

const pane = { left: 320, right: 800 };

describe("place", () => {
  it("starts at the anchor when there is room", () => {
    expect(place({ left: 400, right: 430 }, pane, 220)).toEqual({ left: 400, width: 220 });
  });
  it("shifts left at the right edge of the pane", () => {
    expect(place({ left: 780, right: 800 }, pane, 220)).toEqual({ left: 572, width: 220 });
  });
  it("stays inside the left edge of the pane", () => {
    expect(place({ left: 300, right: 330 }, pane, 220)).toEqual({ left: 328, width: 220 });
  });
  it("shrinks in a narrow pane", () => {
    expect(place({ left: 100, right: 130 }, { left: 0, right: 200 }, 220)).toEqual({ left: 8, width: 184 });
  });
});
