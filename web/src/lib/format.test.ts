import { describe, expect, it } from "vitest";
import { shortPath } from "./format";

describe("shortPath", () => {
  it("writes the home folder as ~", () => {
    expect(shortPath("/Users/me", "/Users/me")).toBe("~");
    expect(shortPath("/Users/me/code/jin", "/Users/me")).toBe("~/code/jin");
  });
  it("leaves other paths and lookalikes alone", () => {
    expect(shortPath("/srv/app", "/Users/me")).toBe("/srv/app");
    expect(shortPath("/Users/meow/x", "/Users/me")).toBe("/Users/meow/x");
    expect(shortPath("/srv/app", "")).toBe("/srv/app");
  });
});
