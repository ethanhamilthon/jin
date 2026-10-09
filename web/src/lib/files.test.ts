import { describe, expect, it } from "vitest";
import { languageOf, mention, size } from "./files";

describe("files helpers", () => {
  it("quotes a path with spaces like the composer", () => {
    expect(mention("src/main.go")).toBe("@src/main.go");
    expect(mention("my docs/a b.md")).toBe('@"my docs/a b.md"');
    expect(mention('say "hi".txt')).toBe('@"say \\"hi\\".txt"');
  });

  it("formats sizes", () => {
    expect([size(900), size(2048), size(5 * 1024 * 1024)]).toEqual(["900 B", "2.0 KB", "5.0 MB"]);
  });

  it("guesses the language from the name", () => {
    expect([languageOf("a.go"), languageOf("Makefile"), languageOf("x.unknown")]).toEqual(["go", "makefile", ""]);
  });
});
