import { describe, expect, it } from "vitest";
import { fileBody } from "./images";

describe("fileBody", () => {
  it("encodes the bytes as base64", async () => {
    const body = await fileBody(new Blob([new Uint8Array([0, 1, 2, 250, 255])]));
    expect(body).toEqual({ data: "AAEC+v8=" });
  });

  it("encodes a file larger than one chunk", async () => {
    const bytes = new Uint8Array(100_000).map((_, i) => i % 251);
    const { data } = await fileBody(new Blob([bytes]));
    expect(Uint8Array.from(atob(data), (c) => c.charCodeAt(0))).toEqual(bytes);
  });

  it("encodes an empty file as an empty string", async () => {
    expect(await fileBody(new Blob([]))).toEqual({ data: "" });
  });
});
