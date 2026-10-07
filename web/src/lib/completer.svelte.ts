import { accept, trigger, type Item, type Trigger } from "./composer";
import { items } from "./complete";

// Completer is the completion list of one draft: prompts or paths.
export class Completer {
  open = $state<Trigger | null>(null);
  list = $state<Item[]>([]);
  selected = $state(0);
  private asked = 0;

  get shown(): boolean {
    return !!this.open && this.list.length > 0;
  }

  async refresh(text: string, cursor: number, dir: string) {
    const t = trigger(text, cursor);
    const ask = ++this.asked;
    this.open = t;
    if (!t) return;
    const found = await items(t, dir).catch(() => []);
    if (ask === this.asked) [this.list, this.selected] = [found, 0];
  }

  move(delta: number) {
    this.selected = (this.selected + delta + this.list.length) % this.list.length;
  }

  close() {
    this.open = null;
  }

  // apply puts the chosen item into text and says where the cursor goes.
  apply(text: string, cursor: number, item = this.list[this.selected]) {
    return this.open ? accept(text, cursor, this.open, item) : { text, cursor };
  }
}
