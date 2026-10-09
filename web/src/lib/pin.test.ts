import { describe, expect, it } from "vitest";
import { Pin } from "./pin";

describe("Pin", () => {
  it("follows by default", () => {
    expect(new Pin().following).toBe(true);
  });

  it("lets go when the reader scrolls up", () => {
    const pin = new Pin();
    pin.up(0);
    pin.scrolled(300, 700, 10);
    expect(pin.following).toBe(false);
  });

  it("ignores scrolls without input, like the browser clamping a shrunk page", () => {
    const pin = new Pin();
    pin.up(0);
    pin.scrolled(300, 700, 10);
    pin.scrolled(0, 1000, 5000);
    expect(pin.following).toBe(false);
  });

  it("ignores its own jump to the end", () => {
    const pin = new Pin();
    pin.scrolled(0, 1200, 5000);
    expect(pin.following).toBe(true);
  });

  it("follows again when the reader scrolls down to the end", () => {
    const pin = new Pin();
    pin.up(0);
    pin.scrolled(300, 700, 10);
    pin.touch(1000);
    pin.scrolled(10, 990, 1010);
    expect(pin.following).toBe(true);
  });

  it("stays unpinned after a small scroll up near the end", () => {
    const pin = new Pin();
    pin.touch(0);
    pin.scrolled(0, 1000, 5);
    pin.up(100);
    pin.scrolled(15, 985, 110);
    expect(pin.following).toBe(false);
  });

  it("ignores a scroll long after the last input", () => {
    const pin = new Pin();
    pin.up(0);
    pin.touch(0);
    pin.scrolled(0, 900, 1000);
    expect(pin.following).toBe(false);
  });

  it("follows again after reset", () => {
    const pin = new Pin();
    pin.up(0);
    pin.reset();
    expect(pin.following).toBe(true);
  });
});
