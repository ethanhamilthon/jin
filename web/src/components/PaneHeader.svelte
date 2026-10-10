<script lang="ts">
  import { app, type SessionView } from "../lib/app.svelte";
  import { closePane, split } from "../lib/actions";
  import { shortPath } from "../lib/format";
  import Icon from "./Icon.svelte";
  import { mobile } from "../lib/mobile.svelte";
  import { openPanel } from "../lib/panels";

  let { index, view }: { index: number; view: SessionView } = $props();
  const state = $derived(view.state);
  const busy = $derived(!!state.busy && !state.shell && !state.reloading && state.ready);
</script>

<div class="bar">
  <button class="switch" onclick={() => app.open("switch")} title="Switch session" aria-label="Switch session"><Icon name="switch" size={14} /></button>
  <div class="title">
    <span class="serif name" title={state.title}>{state.title || "New session"}</span>
    {#if state.path}
      <button class="path mono" title="{state.path} · project settings" onclick={() => openPanel("project", undefined, state.path)}><bdi>{shortPath(state.path, app.home)}</bdi></button>
    {/if}
  </div>
  {#if busy || state.tasks}<span class="dots">
    {#if busy}<span class="dot blink"></span>{/if}
    {#if state.tasks}<span class="dot violet blink" class:alt={busy}></span>{/if}
  </span>{/if}
  {#if state.read_only}<span class="pill warn">read-only · pid {state.read_only}</span>{/if}
  {#if state.provider_missing}<span class="pill error">provider deleted</span>{/if}
  {#if !mobile.on}
  <button class="btn ghost small" onclick={split} title="Split (Alt+\)" disabled={app.panes.length >= 4}><Icon name="split" /></button>
  {#if app.panes.length > 1}
    <button class="btn ghost small" onclick={(e) => { e.stopPropagation(); closePane(index); }} title="Close pane (Alt+W)"><Icon name="close" /></button>
  {/if}
  {/if}
</div>

<style>
  .bar { display: flex; align-items: center; gap: 8px; padding: 8px 12px; border-bottom: 1px solid var(--raised); min-width: 0; }
  .switch {
    flex: none; display: inline-flex; align-items: center; min-height: 44px; padding: 0 8px 0 0;
    background: none; border: 0; color: var(--text-muted); cursor: pointer;
  }
  .switch:hover { color: var(--accent); }
  .dots { flex: none; display: inline-flex; align-items: center; gap: 4px; }
  .dots .alt { animation-delay: 0.55s; }
  .title { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
  .name { font-size: 17px; line-height: 1.3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .path { background: none; border: 0; padding: 0; cursor: pointer; font-size: 11px; line-height: 1.4; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl; text-align: left; }
</style>
