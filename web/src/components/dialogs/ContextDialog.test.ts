import { expect, it } from "vitest";
import source from "./ContextDialog.svelte?raw";

it("renders the system prompt as one text, not per-part blocks", () => {
  expect(source).not.toContain("{#each report.prompt");
  expect(source).toContain("{@html promptHtml}");
});

it("highlights command-collected parts", () => {
  expect(source).toContain('startsWith("command: ")');
  expect(source).toContain("escapeHtml");
});
