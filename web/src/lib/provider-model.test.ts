import { expect, it } from "vitest";
import { providerModelKey } from "./provider-model";

it("keeps identical model names distinct across providers", () => {
  const a = { provider: "a", provider_name: "A", id: "shared" };
  const b = { provider: "b", provider_name: "B", id: "shared" };
  expect(providerModelKey(a)).not.toBe(providerModelKey(b));
});

it("does not collide on separators in provider or model ids", () => {
  expect(providerModelKey({ provider: "a:b", provider_name: "", id: "c" })).not.toBe(
    providerModelKey({ provider: "a", provider_name: "", id: "b:c" }),
  );
});
