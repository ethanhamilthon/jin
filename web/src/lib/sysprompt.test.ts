import { describe, expect, it } from "vitest";
import { parseSections, renderSections } from "./sysprompt";

describe("system prompt file", () => {
  it("round-trips the three sections", () => {
    const sections = { system: "You are jin.\n\n## Rules\nBe brief.", compact: "Summarize.", handoff: "Write a brief." };
    expect(parseSections(renderSections(sections))).toEqual(sections);
  });

  it("ignores text before the first section and keeps other headings in their section", () => {
    const got = parseSections("notes\n# system\nA\n# other heading\nB\n# compact\nC\r\n");
    expect(got.system).toBe("A\n# other heading\nB");
    expect(got.compact).toBe("C");
    expect(got.handoff).toBe("");
  });

  it("keeps the last text of a name that appears twice", () => {
    expect(parseSections("# system\nfirst\n# system\nsecond").system).toBe("second");
  });
});
