export const edge = 24;
const intentMs = 300;

// Pin decides whether the chat follows new content. Only the reader's own
// input moves it: a scroll the browser causes (content that shrinks, our own
// jump to the end) never changes it, so a reader is not pulled down.
export class Pin {
  following = true;
  private input = -Infinity;
  private top = 0;

  touch(now: number) {
    this.input = now;
  }

  up(now: number) {
    this.input = now;
    this.following = false;
  }

  scrolled(gap: number, top: number, now: number) {
    const down = top > this.top;
    this.top = top;
    if (now - this.input > intentMs) return;
    if (gap >= edge) this.following = false;
    else if (down) this.following = true;
  }

  reset() {
    this.following = true;
  }
}
