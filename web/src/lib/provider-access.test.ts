import { describe, expect, it } from "vitest";
import type { Provider } from "./types";
import { noEnabledProvider } from "./provider-access";

const provider = (id: string, enabled?: boolean): Provider => ({ id, name: id, kind: "api", base_url: "", enabled });

describe("noEnabledProvider", () => {
  it("is false without providers, the first-run screen handles that", () => {
    expect(noEnabledProvider([], "a")).toBe(false);
  });

  it("treats a provider without the enabled flag as enabled", () => {
    expect(noEnabledProvider([provider("a")], "a")).toBe(false);
    expect(noEnabledProvider([provider("a", true)], "a")).toBe(false);
  });

  it("is true when no provider is enabled", () => {
    expect(noEnabledProvider([provider("a", false), provider("b", false)], "a")).toBe(true);
  });

  it("is true when the session's provider is disabled, even if another one is enabled", () => {
    expect(noEnabledProvider([provider("a", false), provider("b", true)], "a")).toBe(true);
  });

  it("is false when the session's provider is enabled, even if another one is disabled", () => {
    expect(noEnabledProvider([provider("a", true), provider("b", false)], "a")).toBe(false);
  });
});
