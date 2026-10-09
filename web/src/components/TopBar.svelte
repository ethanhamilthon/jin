<script lang="ts">
  import { app } from "../lib/app.svelte";
  import Icon from "./Icon.svelte";
  import { openPanel } from "../lib/panels";

  import { devices, others } from "../lib/devices.svelte";

  const online = $derived(others());
  const tasks = $derived(app.live().reduce((n, s) => n + (s.tasks ?? 0), 0));
</script>

<header>
  <button class="btn ghost small" onclick={() => (app.sidebar = !app.sidebar)} title="Toggle sidebar"><Icon name="menu" /></button>
  <span class="word serif">jin</span>
  {#if !app.connected}<span class="pill warn">reconnecting</span>{/if}
  <button class="pill remote" class:dim={!devices.remote} class:warn={devices.remote && !online} onclick={() => app.open("pair")} title="Remote access">
    {devices.remote ? (online ? `${online} connected` : "not connected") : "remote off"}
  </button>
  <span class="spacer"></span>
  <button class="btn ghost small" onclick={() => app.open("tasks")} title="Background tasks">
    <Icon name="tasks" /><span class="lbl">Tasks</span>{#if tasks}<span class="count">{tasks}</span>{/if}
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
  .remote { background: transparent; cursor: pointer; }
  .spacer { flex: 1; }
  @media (max-width: 700px) {
    header { gap: 0; padding: 0 6px; }
    .remote, .lbl { display: none; }
    .word { margin-left: 4px; }
  }
  .count { color: var(--accent); font-family: var(--mono); font-size: 11px; }
</style>
