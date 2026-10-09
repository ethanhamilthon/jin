import { describe, expect, it } from "vitest";
import { accept, imageLabel, rank, shownSuggestion, trigger } from "./composer";

describe("trigger", () => {
  it("ignores a leading slash", () => {
    expect(trigger("/comp", 5)).toBeNull();
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

describe("shownSuggestion", () => {
  const idle = { suggestion: "run it", busy: false, ready: true };
  it("shows in an empty open input", () => {
    expect(shownSuggestion(idle, "", false, "")).toBe("run it");
  });
  it("hides while typing, attached, busy, starting or read-only", () => {
    expect(shownSuggestion(idle, "x", false, "")).toBe("");
    expect(shownSuggestion(idle, "", true, "")).toBe("");
    expect(shownSuggestion({ ...idle, busy: true }, "", false, "")).toBe("");
    expect(shownSuggestion({ ...idle, ready: false }, "", false, "")).toBe("");
    expect(shownSuggestion({ ...idle, read_only: 12 }, "", false, "")).toBe("");
  });
  it("stays hidden once taken, but a new one shows", () => {
    expect(shownSuggestion(idle, "", false, "run it")).toBe("");
    expect(shownSuggestion({ ...idle, suggestion: "next" }, "", false, "run it")).toBe("next");
  });
});
