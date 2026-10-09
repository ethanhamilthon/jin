export type Box = { left: number; right: number };

const gap = 8;

// place keeps a menu of the wanted width inside its pane: it starts at the
// anchor and shifts left or shrinks when the pane has no room.
export function place(anchor: Box, pane: Box, wanted: number): { left: number; width: number } {
  const width = Math.max(0, Math.min(wanted, pane.right - pane.left - 2 * gap));
  const left = Math.min(Math.max(anchor.left, pane.left + gap), pane.right - width - gap);
  return { left, width };
}
