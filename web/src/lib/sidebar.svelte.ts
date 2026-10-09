import { app } from "./app.svelte";

const storageKey = "jin.tree";

function readExpanded(): Record<string, boolean> {
  try {
    return JSON.parse(localStorage.getItem(storageKey) ?? "{}");
  } catch {
    return {};
  }
}

class Tree {
  expanded = $state<Record<string, boolean>>(readExpanded());

  // save keeps which projects are folded or open in this browser.
  save() {
    try {
      localStorage.setItem(storageKey, JSON.stringify(this.expanded));
    } catch {
      // not kept
    }
  }

  // The project of the focused session is open unless it was closed by hand.
  isOpen(path: string): boolean {
    return this.expanded[path] ?? path === app.project;
  }

  toggle(path: string) {
    this.expanded[path] = !this.isOpen(path);
    this.save();
  }

  open(path: string) {
    this.expanded[path] = true;
    this.save();
  }
}

export const tree = new Tree();
