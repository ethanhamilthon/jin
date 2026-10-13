import { describe, expect, it } from "vitest";
import { providerKinds } from "./provider-kind";

describe("provider setup kinds", () => {
  it("offers only automatic OpenAI-compatible and Anthropic APIs", () => {
    expect(providerKinds.map(({ id }) => id)).toEqual(["auto", "anthropic"]);
  });
});
