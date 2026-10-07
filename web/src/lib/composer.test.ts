import { describe, expect, it } from "vitest";
import { accept, imageLabel, rank, trigger } from "./composer";

describe("trigger", () => {
  it("finds commands only at the start", () => {
    expect(trigger("/comp", 5)).toEqual({ kind: "/", start: 0, query: "comp" });
    expect(trigger("a /comp", 7)).toBeNull();
  });
  it("finds prompts and paths after whitespace", () => {
    expect(trigger("use #rev", 8)).toEqual({ kind: "#", start: 4, query: "rev" });
    expect(trigger("see @src/ma", 11)).toEqual({ kind: "@", start: 4, query: "src/ma" });
    expect(trigger("mail@x", 6)).toBeNull();
  });
});

describe("accept", () => {
  it("replaces the token and adds a space", () => {
    const t = trigger("use #rev now", 8)!;
    expect(accept("use #rev now", 8, t, { label: "review", insert: "#review" })).toEqual({ text: "use #review now", cursor: 12 });
  });
  it("keeps going into folders", () => {
    const t = trigger("@sr", 3)!;
    expect(accept("@sr", 3, t, { label: "src/", insert: "@src/", more: true })).toEqual({ text: "@src/", cursor: 5 });
  });
});

describe("helpers", () => {
  it("names images", () => expect(imageLabel(3)).toBe("[image 03]"));
  it("ranks prefix matches first", () => {
    expect(rank([{ label: "areview" }, { label: "review" }], "rev").map((i) => i.label)).toEqual(["review", "areview"]);
  });
});
