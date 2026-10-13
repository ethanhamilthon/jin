import { describe, expect, it } from "vitest";

const modules = import.meta.glob("../**/*.{svelte,ts}", { query: "?raw", import: "default", eager: true }) as Record<string, string>;
const files = Object.keys(modules).filter((f) => !f.endsWith(".test.ts"));

describe("retired title and prompt UI", () => {
  it("has no removed component files", () => {
    for (const name of ["Title", "SystemPrompt", "CLIProxy", "Subscription", "signin"]) {
      expect(files.filter((f) => f.includes("/" + name))).toEqual([]);
    }
  });

  it("calls no removed title or prompt endpoints", () => {
    const text = files.map((f) => modules[f]).join("\n");
    for (const ghost of ["generate-title", "api/settings/title", "api/sysprompt", "api/cliproxy", "system_prompt"]) {
      expect(text).not.toContain(ghost);
    }
  });
});
