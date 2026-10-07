import { app } from "./app.svelte";

class Tree {
  expanded = $state<Record<string, boolean>>({});

  // The project of the focused session is open unless it was closed by hand.
  isOpen(path: string): boolean {
    return this.expanded[path] ?? path === app.project;
  }

  toggle(path: string) {
    this.expanded[path] = !this.isOpen(path);
  }

  open(path: string) {
    this.expanded[path] = true;
  }
}

export const tree = new Tree();
