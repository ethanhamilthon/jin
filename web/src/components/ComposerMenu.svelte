<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { act } from "../lib/actions";
  import { exportMarkdown } from "../lib/export";
  import { toggleTools, toolsShown } from "../lib/fold";
  import { place } from "../lib/menu_place";
  import { generateTitle } from "../lib/title";
  import Icon from "./Icon.svelte";
  import MenuRow from "./MenuRow.svelte";

  let { id, model, effort, disabled, attach }: {
    id: string; model: string; effort: string; disabled: boolean; attach: (files: File[]) => void;
  } = $props();
  let open = $state(false);
  let root: HTMLDivElement;
  let button: HTMLButtonElement;
  let files: HTMLInputElement;
  let spot = $state({ left: 0, bottom: 0, width: 0 });
  const shown = $derived(toolsShown(app.config.fold));

  function toggle() {
    if (!open) {
      const at = button.getBoundingClientRect();
      const pane = root.closest("section")?.getBoundingClientRect() ?? { left: 0, right: innerWidth };
      spot = { ...place(at, pane, 270), bottom: innerHeight - at.top + 6 };
    }
    open = !open;
  }

  function pick(fn: () => void) {
    open = false;
    fn();
  }
</script>

<svelte:window
  onmousedown={(e) => open && !root.contains(e.target as Node) && (open = false)}
  onkeydown={(e) => open && e.key === "Escape" && (open = false)}
  onresize={() => (open = false)}
/>

<div class="more" bind:this={root}>
  <input bind:this={files} type="file" multiple hidden onchange={() => { attach([...(files.files ?? [])]); files.value = ""; }} />
  <button bind:this={button} class="btn ghost small" class:on={open} onclick={toggle} title="Menu" aria-label="Menu" aria-expanded={open}><Icon name="menu" size={18} /></button>
  {#if open}
    <div class="menu" role="menu" style="left: {spot.left}px; bottom: {spot.bottom}px; width: {spot.width}px">
      <MenuRow icon="command" label="Model" hint={effort ? `${model} · ${effort}` : model} kind="open" pick={() => pick(() => app.open("model", id))} />
      <MenuRow icon="clip" label="Attach files" kind="run" pick={() => { open = false; if (!disabled) files.click(); }} />
      <MenuRow icon="context" label="Context" kind="open" pick={() => pick(() => app.open("context", id))} />
      <MenuRow icon="eye" label="Chat details" kind="toggle" on={shown} pick={() => pick(toggleTools)} />
      <hr />
      <MenuRow icon="compact" label="Compact" kind="run" pick={() => pick(() => act(id, "compact"))} />
      <MenuRow icon="handoff" label="Handoff" kind="run" pick={() => pick(() => act(id, "handoff"))} />
      <MenuRow icon="edit" label="Generate title" kind="run" pick={() => pick(() => generateTitle(id))} />
      <MenuRow icon="rewind" label="Rewind" kind="open" pick={() => pick(() => app.open("rewind", id))} />
      <hr />
      <MenuRow icon="reload" label="Reload prompts" kind="run" pick={() => pick(() => act(id, "reload"))} />
      <MenuRow icon="export" label="Export as Markdown" kind="run" pick={() => pick(() => exportMarkdown(id).catch((e) => app.toast(String(e), true)))} />
    </div>
  {/if}
</div>

<style>
  .on { color: var(--accent); }
  .menu {
    position: fixed; z-index: 20; padding: 4px;
    background: var(--card); border: 1px solid var(--line); border-radius: var(--radius); box-shadow: 0 8px 24px rgb(0 0 0 / 0.5);
  }
  @media (max-width: 700px) {
    .menu {
      left: 0 !important; bottom: 0 !important; width: 100% !important; padding: 8px 0 calc(8px + env(safe-area-inset-bottom));
      border-width: 1px 0 0; border-radius: 0; box-shadow: 0 0 0 100vmax rgb(0 0 0 / 0.6);
    }
  }
  hr { border: 0; border-top: 1px solid var(--raised); margin: 4px 0; }
</style>
