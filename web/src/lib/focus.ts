// focus puts the cursor into an element once it is on the page.
export function focus(node: HTMLElement) {
  requestAnimationFrame(() => node.focus());
}
