import { describe, expect, it } from "vitest";
import { Marked } from "marked";
import { alertQuote, highlight, localImage } from "./markdown_ext";
import { footnotes, notesHtml, resetNotes } from "./footnotes";

describe("markdown extensions", () => {
  it("makes alerts", () => {
    expect(alertQuote("<p>[!WARNING]\nCareful</p>")).toContain('class="alert alert-warning"');
    expect(alertQuote("<p>plain</p>")).toBeNull();
  });

  it("highlights ==text==", () => {
    expect(new Marked(highlight).parse("a ==b== c", { async: false })).toContain("<mark>b</mark>");
  });

  it("routes local images through the server", () => {
    expect(localImage("https://x/y.png", "/p")).toBe("https://x/y.png");
    expect(localImage("shot.png", "/p")).toBe("/api/local-image?path=shot.png&dir=%2Fp");
    expect(localImage("file:///a/b.png", "")).toBe("/api/local-image?path=%2Fa%2Fb.png&dir=");
  });

  it("numbers footnotes by first use", () => {
    resetNotes();
    const marked = new Marked(footnotes);
    const html = marked.parse("Hi[^b] and[^a].\n\n[^a]: First\n[^b]: Second\n", { async: false }) as string;
    expect(html).toContain('href="#fn-b">1<');
    const list = notesHtml((t) => t);
    expect(list.indexOf("Second")).toBeLessThan(list.indexOf("First"));
  });
});
