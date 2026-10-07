<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { act } from "../lib/actions";
  import { exportMarkdown } from "../lib/export";
  import { toggleTools, toolsShown } from "../lib/fold";
  import Icon from "./Icon.svelte";

  let { id }: { id: string } = $props();
  let open = $state(false);
  let root: HTMLDivElement;

  const items: { label: string; hint?: string; run: () => void; checked?: () => boolean; gap?: boolean }[] = [
    { label: "Show tool outputs", hint: "Ctrl+O", run: toggleTools, checked: () => toolsShown(app.config.fold) },
    { label: "Context", run: () => app.open("context", id) },
    { label: "Compact", run: () => act(id, "compact"), gap: true },
    { label: "Handoff", run: () => act(id, "handoff") },
    { label: "Rewind", run: () => app.open("rewind", id) },
    { label: "Undo", run: () => app.open("undo", id) },
    { label: "Reload prompts", run: () => act(id, "reload"), gap: true },
    { label: "Export as Markdown", run: () => exportMarkdown(id).catch((e) => app.toast(String(e), true)) },
  ];

  function choose(item: (typeof items)[number]) {
    open = false;
    item.run();
  }
</script>

<svelte:window
  onmousedown={(e) => open && !root.contains(e.target as Node) && (open = false)}
  onkeydown={(e) => open && e.key === "Escape" && (open = false)}
/>

<div class="more" bind:this={root}>
  <button class="btn ghost small" class:on={open} onclick={() => (open = !open)} title="More" aria-label="More" aria-expanded={open}><Icon name="dots" size={15} /></button>
  {#if open}
    <div class="menu" role="menu">
      {#each items as item (item.label)}
        {#if item.gap}<hr />{/if}
        <button role="menuitem" onclick={() => choose(item)}>
          <span class="check">{#if item.checked?.()}<Icon name="check" size={13} />{/if}</span>
          <span class="label-text">{item.label}</span>
          {#if item.hint}<span class="hint mono">{item.hint}</span>{/if}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style>
  .more { position: relative; }
  .on { color: var(--accent); }
  .menu {
    position: absolute; right: 0; bottom: calc(100% + 6px); z-index: 20; min-width: 220px; padding: 4px;
    background: var(--card); border: 1px solid var(--line); border-radius: var(--radius); box-shadow: 0 8px 24px rgb(0 0 0 / 0.5);
  }
  .menu button { display: flex; align-items: center; gap: 8px; width: 100%; padding: 6px 8px; border: 0; background: transparent; border-radius: 2px; color: var(--text); font-size: 13px; text-align: left; cursor: pointer; }
  .menu button:hover { background: var(--hover); }
  .check { width: 14px; display: inline-flex; color: var(--accent); }
  .label-text { flex: 1; }
  .hint { color: var(--text-muted); font-size: 11px; }
  hr { border: 0; border-top: 1px solid var(--raised); margin: 4px 0; }
</style>
