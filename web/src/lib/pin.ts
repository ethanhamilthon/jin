export const edge = 24;
const intentMs = 300;

// Pin decides whether the chat follows new content. Only the reader's own
// input moves it, and only one way: once the reader has scrolled up, the chat
// never follows again until reset, which happens when the reader sends a
// message or opens another session. A scroll the browser causes (content that
// shrinks, our own jump to the end) never changes it.
export class Pin {
  following = true;
  private input = -Infinity;

  touch(now: number) {
    this.input = now;
  }

  up(now: number) {
    this.input = now;
    this.following = false;
  }

  scrolled(gap: number, now: number) {
    if (now - this.input <= intentMs && gap >= edge) this.following = false;
  }

  reset() {
    this.following = true;
  }
}
