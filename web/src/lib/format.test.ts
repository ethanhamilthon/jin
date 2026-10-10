import { describe, expect, it } from "vitest";
import { prettyModel, shortPath } from "./format";

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

describe("prettyModel", () => {
  it("names mapped families and versions", () => {
    expect(prettyModel("claude-sonnet-5.5")).toBe("Claude Sonnet 5.5");
    expect(prettyModel("claude-haiku-4-5")).toBe("Claude Haiku 4.5");
    expect(prettyModel("gpt-4o")).toBe("GPT-4o");
    expect(prettyModel("gemini-2.5-pro")).toBe("Gemini 2.5 Pro");
    expect(prettyModel("gpt-5.3-codex-spark")).toBe("GPT 5.3 Codex Spark");
  });
  it("keeps o-series names uppercase", () => {
    expect(prettyModel("o1")).toBe("O1");
    expect(prettyModel("o4-mini")).toBe("O4 Mini");
  });
  it("keeps the GPT-4o form when more words follow", () => {
    expect(prettyModel("gpt-4o-mini")).toBe("GPT-4o Mini");
  });
  it("capitalizes unknown words and keeps their other characters", () => {
    expect(prettyModel("my_model-x1")).toBe("My Model X1");
    expect(prettyModel("llama3-70b-Instruct")).toBe("Llama3 70b Instruct");
  });
  it("keeps the provider prefix before a slash and names the rest", () => {
    expect(prettyModel("cline-free/kimi-k3")).toBe("cline-free · Kimi K3");
    expect(prettyModel("openrouter/anthropic/claude-haiku-4-5")).toBe("openrouter/anthropic · Claude Haiku 4.5");
  });
});
