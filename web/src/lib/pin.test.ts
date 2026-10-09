import { describe, expect, it } from "vitest";
import { Pin } from "./pin";

describe("Pin", () => {
  it("follows by default", () => {
    expect(new Pin().following).toBe(true);
  });

  it("lets go when the reader scrolls up", () => {
    const pin = new Pin();
    pin.up(0);
    pin.scrolled(300, 10);
    expect(pin.following).toBe(false);
  });

  it("lets go when the reader scrolls away from the end by any input", () => {
    const pin = new Pin();
    pin.touch(0);
    pin.scrolled(120, 10);
    expect(pin.following).toBe(false);
  });

  it("ignores scrolls without input, like the browser clamping a shrunk page", () => {
    const pin = new Pin();
    pin.scrolled(300, 5000);
    expect(pin.following).toBe(true);
  });

  it("ignores its own jump to the end", () => {
    const pin = new Pin();
    pin.scrolled(0, 5000);
    expect(pin.following).toBe(true);
  });

  it("never follows again when the reader scrolls back down to the end", () => {
    const pin = new Pin();
    pin.up(0);
    pin.scrolled(300, 10);
    pin.touch(1000);
    pin.scrolled(0, 1010);
    expect(pin.following).toBe(false);
  });

  it("stays released after a small scroll up near the end", () => {
    const pin = new Pin();
    pin.up(100);
    pin.scrolled(15, 110);
    expect(pin.following).toBe(false);
  });

  it("follows again after reset, when the reader sends a message", () => {
    const pin = new Pin();
    pin.up(0);
    pin.reset();
    expect(pin.following).toBe(true);
  });
});
