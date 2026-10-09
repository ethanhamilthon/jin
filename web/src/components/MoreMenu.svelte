<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { act } from "../lib/actions";
  import { exportMarkdown } from "../lib/export";
  import { toggleTools, toolsShown } from "../lib/fold";
  import { place } from "../lib/menu_place";
  import Icon from "./Icon.svelte";
  import MenuRow from "./MenuRow.svelte";

  let { id }: { id: string } = $props();
  let open = $state(false);
  let root: HTMLDivElement;
  let button: HTMLButtonElement;
  let spot = $state({ left: 0, bottom: 0, width: 0 });

  type Item = { icon: string; label: string; kind: "open" | "run" | "toggle"; run: () => void; hint?: string; gap?: boolean };
  const items: Item[] = [
    { icon: "eye", label: "Show tool outputs", kind: "toggle", hint: "Ctrl+O", run: toggleTools },
    { icon: "context", label: "Context", kind: "open", run: () => app.open("context", id) },
    { icon: "compact", label: "Compact", kind: "run", run: () => act(id, "compact"), gap: true },
    { icon: "handoff", label: "Handoff", kind: "run", run: () => act(id, "handoff") },
    { icon: "rewind", label: "Rewind", kind: "open", run: () => app.open("rewind", id) },
    { icon: "undo", label: "Undo", kind: "open", run: () => app.open("undo", id) },
    { icon: "reload", label: "Reload prompts", kind: "run", run: () => act(id, "reload"), gap: true },
    { icon: "export", label: "Export as Markdown", kind: "run", run: () => exportMarkdown(id).catch((e) => app.toast(String(e), true)) },
  ];

  function toggle() {
    if (!open) {
      const at = button.getBoundingClientRect();
      const pane = root.closest("section")?.getBoundingClientRect() ?? { left: 0, right: innerWidth };
      spot = { ...place(at, pane, 270), bottom: innerHeight - at.top + 6 };
    }
    open = !open;
  }

  function choose(item: Item) {
    if (item.kind !== "toggle") open = false;
    item.run();
  }
</script>

<svelte:window
  onmousedown={(e) => open && !root.contains(e.target as Node) && (open = false)}
  onkeydown={(e) => open && e.key === "Escape" && (open = false)}
  onresize={() => (open = false)}
/>

<div class="more" bind:this={root}>
  <button bind:this={button} class="btn ghost small" class:on={open} onclick={toggle} title="More" aria-label="More" aria-expanded={open}><Icon name="dots" size={15} /></button>
  {#if open}
    <div class="menu" role="menu" style="left: {spot.left}px; bottom: {spot.bottom}px; width: {spot.width}px">
      {#each items as item (item.label)}
        {#if item.gap}<hr />{/if}
        <MenuRow icon={item.icon} label={item.label} kind={item.kind} hint={item.hint} on={item.kind === "toggle" && toolsShown(app.config.fold)} pick={() => choose(item)} />
      {/each}
    </div>
  {/if}
</div>

<style>
  .on { color: var(--accent); }
  .menu {
    position: fixed; z-index: 20; padding: 4px;
    background: var(--card); border: 1px solid var(--line); border-radius: var(--radius); box-shadow: 0 8px 24px rgb(0 0 0 / 0.5);
  }
  hr { border: 0; border-top: 1px solid var(--raised); margin: 4px 0; }
</style>
