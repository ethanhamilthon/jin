<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { baseName } from "../lib/format";
  import Icon from "./Icon.svelte";

  let { sidebar = $bindable() }: { sidebar: boolean } = $props();
  const where = $derived(app.project || app.dir);
  const tasks = $derived(app.live().reduce((n, s) => n + (s.tasks ?? 0), 0));
</script>

<header>
  <button class="btn ghost small" onclick={() => (sidebar = !sidebar)} title="Toggle sidebar"><Icon name="menu" /></button>
  <span class="word serif">jin</span>
  <span class="project" title={where}>
    <span class="name">{baseName(where)}</span>
    <span class="path mono">{where}</span>
  </span>
  {#if !app.connected}<span class="pill warn">reconnecting</span>{/if}
  <span class="spacer"></span>
  <button class="btn ghost small" onclick={() => app.open("tasks")} title="Background tasks">
    <Icon name="tasks" />Tasks{#if tasks}<span class="count">{tasks}</span>{/if}
  </button>
  <button class="btn ghost small" onclick={() => app.open("settings")} title="Settings"><Icon name="gear" /></button>
</header>

<style>
  header {
    grid-area: top; display: flex; align-items: center; gap: 12px; padding: 0 12px;
    background: var(--void); border-bottom: 1px solid var(--raised);
  }
  .word { font-size: 22px; line-height: 1; margin-right: 4px; }
  .project { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
  .name { color: var(--text-strong); font-size: 13px; white-space: nowrap; }
  .path { color: var(--text-muted); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .spacer { flex: 1; }
  .count { color: var(--accent); font-family: var(--mono); font-size: 11px; }
</style>
