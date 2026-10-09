<script lang="ts">
  import { app, type SessionView } from "../lib/app.svelte";
  import { closePane, split } from "../lib/actions";
  import { shortPath } from "../lib/format";
  import Icon from "./Icon.svelte";
  import { openPanel } from "../lib/panels";

  let { index, view }: { index: number; view: SessionView } = $props();
  const state = $derived(view.state);
</script>

<div class="bar">
  <div class="title">
    <span class="serif name" title={state.title}>{state.title || "New session"}</span>
    {#if state.path}
      <button class="path mono" title="{state.path} · project settings" onclick={() => openPanel("project", undefined, state.path)}><bdi>{shortPath(state.path, app.home)}</bdi></button>
    {/if}
  </div>
  {#if state.read_only}<span class="pill warn">read-only · pid {state.read_only}</span>{/if}
  {#if state.provider_missing}<span class="pill error">provider deleted</span>{/if}
  {#if state.busy}<span class="pill"><span class="dot blink"></span>{state.shell ? "shell" : state.reloading ? "reloading" : !state.ready ? "starting" : "working"}</span>{/if}
  {#if state.tasks}<span class="pill violet"><span class="dot violet blink"></span>{state.tasks} task{state.tasks > 1 ? "s" : ""}</span>{/if}
  <button class="btn ghost small" onclick={split} title="Split (Alt+\)" disabled={app.panes.length >= 4}><Icon name="split" /></button>
  {#if app.panes.length > 1}
    <button class="btn ghost small" onclick={(e) => { e.stopPropagation(); closePane(index); }} title="Close pane (Alt+W)"><Icon name="close" /></button>
  {/if}
</div>

<style>
  .bar { display: flex; align-items: center; gap: 8px; padding: 8px 12px; border-bottom: 1px solid var(--raised); min-width: 0; }
  .title { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
  .name { font-size: 17px; line-height: 1.3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .path { background: none; border: 0; padding: 0; cursor: pointer; font-size: 11px; line-height: 1.4; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl; text-align: left; }
</style>
