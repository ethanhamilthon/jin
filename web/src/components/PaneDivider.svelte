<script lang="ts">
  let { orientation, pos, from = 0, onDrag, onRelease }: {
    orientation: "col" | "row"; pos: number; from?: number;
    onDrag: (delta: number) => void; onRelease: () => void;
  } = $props();
  let root: HTMLButtonElement;
  let size = 1;
  let last = 0;

  function down(event: PointerEvent) {
    event.preventDefault();
    root.setPointerCapture(event.pointerId);
    const rect = root.closest("main")?.getBoundingClientRect();
    size = orientation === "col" ? (rect?.width ?? 1) : (rect?.height ?? 1);
    last = orientation === "col" ? event.clientX : event.clientY;
  }

  function move(event: PointerEvent) {
    if (!event.buttons) return;
    const at = orientation === "col" ? event.clientX : event.clientY;
    onDrag((at - last) / size);
    last = at;
  }
</script>

<button
  bind:this={root}
  class="div"
  class:row={orientation === "row"}
  style={orientation === "col" ? `left: ${pos * 100}%;` : `top: ${pos * 100}%; left: ${from * 100}%;`}
  aria-label={orientation === "col" ? "Resize columns" : "Resize rows"}
  onpointerdown={down}
  onpointermove={move}
  onpointerup={onRelease}
  onpointercancel={onRelease}
></button>

<style>
  .div {
    position: absolute; top: 0; bottom: 0; left: 0; width: 13px; transform: translateX(-50%);
    z-index: 5; padding: 0; background: transparent; border: 0; cursor: col-resize; touch-action: none;
  }
  .row {
    top: 0; bottom: auto; left: 0; right: 0; width: auto; height: 13px; transform: translateY(-50%);
    cursor: row-resize;
  }
  .div:hover, .div:active { background: color-mix(in srgb, var(--accent) 22%, transparent); }
</style>
