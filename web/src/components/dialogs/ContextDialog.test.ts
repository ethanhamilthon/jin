import { parse } from "svelte/compiler";
import { expect, it } from "vitest";
import source from "./ContextDialog.svelte?raw";

it("keys both prompt lists by position, allowing repeated names and token counts", () => {
  const ast = parse(source, { modern: true });
  const lists: Record<string, unknown>[] = [];

  function visit(value: unknown) {
    if (!value || typeof value !== "object") return;
    if (Array.isArray(value)) return value.forEach(visit);
    const node = value as Record<string, unknown>;
    if (node.type === "EachBlock") {
      const expression = node.expression as { left?: { property?: { name?: string } } };
      if (expression.left?.property?.name === "prompt") lists.push(node);
    }
    Object.values(node).forEach(visit);
  }

  visit(ast.fragment);
  expect(lists).toHaveLength(2);
  for (const list of lists) {
    expect(list.index).toEqual(expect.any(String));
    expect(list.key).toMatchObject({ type: "Identifier", name: list.index });
  }
});
