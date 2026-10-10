<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, query } from "../../lib/api";
  import { ago, shortPath } from "../../lib/format";
  import { groupRows } from "../../lib/buckets";
  import { newSession, switchSession } from "../../lib/actions";
  import type { Project, SessionRow } from "../../lib/types";
  import Dialog from "../Dialog.svelte";
  import Icon from "../Icon.svelte";
  import SwitchList from "./SwitchList.svelte";

  const week = 7 * 24 * 3600 * 1000;

  let projects = $state<Project[]>([]);
  let chosen = $state<Project | null>(null);
  let rows = $state<SessionRow[]>([]);
  let loaded = $state(false);
  let search = $state("");

  const active = $derived(projects.filter((p) => !p.archived));
  const shown = $derived(rows.filter((row) => Date.now() - new Date(row.updated_at).getTime() < week && (row.title || "").toLowerCase().includes(search.trim().toLowerCase())));
  const groups = $derived(groupRows(shown, (row) => !!app.sessions[row.id]?.state.busy, Date.now()));

  $effect(() => {
    void app.projectsRev, app.sessionsRev;
    get<Project[]>("/api/projects").then((list) => (projects = list)).catch(fail);
  });

  $effect(() => {
    void app.sessionsRev;
    const path = chosen?.path;
    if (!path) return;
    let stale = false;
    const timer = setTimeout(() => {
      get<SessionRow[]>("/api/sessions" + query({ path }))
        .then((list) => { if (!stale) ((rows = list), (loaded = true)); })
        .catch(fail);
    }, 200);
    return () => { stale = true; clearTimeout(timer); };
  });

  function open(project: Project) {
    rows = [];
    loaded = false;
    chosen = project;
  }

  function back() {
    chosen = null;
    search = "";
  }

  async function pick(id: string) {
    await switchSession(id);
    app.dialog = null;
  }

  async function startNew(path: string) {
    await newSession(path);
    app.dialog = null;
  }
</script>

<Dialog title={chosen ? "Sessions" : "Projects"} label={chosen ? shortPath(chosen.path, app.home) : "switch"}>
  {#if chosen}
    {@const project = chosen}
    <div class="bar">
      <button class="btn ghost small" onclick={back}><Icon name="back" size={14} />Projects</button>
      <input class="field" placeholder="Search titles" aria-label="Search titles" bind:value={search} />
    </div>
    <div class="rows">
      <button class="list-row" onclick={() => startNew(project.path)}><Icon name="plus" size={14} /><span class="title">New session</span></button>
      {#each groups as group (group.name)}
        <p class="label group">{group.name}</p>
        {#each group.rows as row (row.id)}
          {@const state = app.sessions[row.id]?.state}
          <button class="list-row" title={row.title} onclick={() => pick(row.id)}>
            {#if state?.busy}<span class="dot blink"></span>{:else if state ? state.unread : row.unread}<span class="dot"></span>{/if}
            <span class="title">{row.title || "Untitled"}</span>
            <span class="time">{ago(row.updated_at)}</span>
          </button>
        {/each}
      {:else}
        {#if loaded}<p class="empty">{search ? "No matches" : "No recent sessions"}</p>{/if}
      {/each}
    </div>
  {:else}
    <SwitchList projects={active} choose={open} />
  {/if}
</Dialog>

<style>
  .bar { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
  .bar .field { flex: 1; min-width: 0; }
  .rows { display: grid; gap: 2px; }
  .group { margin: 12px 10px 4px; }
  .list-row :global(svg) { flex: none; }
  .title { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .time { flex: none; font-size: 12px; color: var(--text-muted); }
</style>
