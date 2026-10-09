<script lang="ts">
  import { app } from "../lib/app.svelte";
  import Icon from "./Icon.svelte";
  import { openPanel } from "../lib/panels";

  const tasks = $derived(app.live().reduce((n, s) => n + (s.tasks ?? 0), 0));
</script>

<header>
  <button class="btn ghost small" onclick={() => (app.sidebar = !app.sidebar)} title="Toggle sidebar"><Icon name="menu" /></button>
  <span class="word serif">jin</span>
  {#if !app.connected}<span class="pill warn">reconnecting</span>{/if}
  <span class="spacer"></span>
  <button class="btn ghost small" onclick={() => app.open("tasks")} title="Background tasks">
    <Icon name="tasks" />Tasks{#if tasks}<span class="count">{tasks}</span>{/if}
  </button>
  <button class="btn ghost small" onclick={() => openPanel("files")} title="Files"><Icon name="folder" /></button>
  <button class="btn ghost small" onclick={() => openPanel("settings")} title="Settings"><Icon name="gear" /></button>
</header>

<style>
  header {
    grid-area: top; display: flex; align-items: center; gap: 12px; padding: 0 12px;
    background: var(--void); border-bottom: 1px solid var(--raised);
  }
  .word { font-size: 22px; line-height: 1; margin-right: 4px; }
  .spacer { flex: 1; }
  .count { color: var(--accent); font-family: var(--mono); font-size: 11px; }
</style>
