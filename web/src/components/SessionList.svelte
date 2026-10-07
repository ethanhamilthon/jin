<script lang="ts">
  import { app, fail } from "../lib/app.svelte";
  import { get, query } from "../lib/api";
  import { ago } from "../lib/format";
  import { focus } from "../lib/focus";
  import { newSession, openSession } from "../lib/actions";
  import type { SessionRow } from "../lib/types";
  import Icon from "./Icon.svelte";

  let rows = $state<SessionRow[]>([]);
  let searching = $state(false);
  let search = $state("");

  $effect(() => {
    void app.sessionsRev;
    const path = app.project, q = search;
    const timer = setTimeout(() => {
      get<SessionRow[]>("/api/sessions" + query({ path, q }))
        .then((list) => (rows = list))
        .catch(fail);
    }, q ? 200 : 0);
    return () => clearTimeout(timer);
  });

  function toggleSearch() {
    searching = !searching;
    if (!searching) search = "";
  }
</script>

<section>
  <div class="head">
    <span class="label">Sessions</span>
    <span class="tools">
      <button class="btn ghost small" class:on={searching} onclick={toggleSearch} title="Search sessions"><Icon name="search" /></button>
      <button class="btn ghost small" onclick={() => newSession(app.project)} title="New session (Alt+N)"><Icon name="plus" /></button>
    </span>
  </div>
  {#if searching}
    <div class="search">
      <input class="field" placeholder="Search titles and messages" bind:value={search} use:focus onkeydown={(e) => e.key === "Escape" && toggleSearch()} />
    </div>
  {/if}
  <div class="rows">
    {#each rows as row (row.id)}
      {@const state = app.sessions[row.id]?.state}
      <button class="list-row" class:active={app.pane.session === row.id} onclick={() => openSession(row.id)} title={row.title}>
        {#if state?.busy}<span class="dot blink"></span>
        {:else if state ? state.unread : row.unread}<span class="dot"></span>{/if}
        <span class="title">{row.title || "Untitled"}</span>
        <span class="time">{ago(row.updated_at)}</span>
      </button>
    {:else}
      <p class="empty">{search ? "Nothing found" : "No sessions yet"}</p>
    {/each}
  </div>
</section>

<style>
  section { flex: 1; min-height: 0; display: flex; flex-direction: column; padding: 12px 10px; }
  .head { display: flex; align-items: center; justify-content: space-between; padding: 0 4px 6px; }
  .tools { display: flex; gap: 2px; }
  .on { color: var(--accent); }
  .search { padding: 0 4px 8px; }
  .rows { overflow-y: auto; flex: 1; }
  .list-row { padding: 10px 12px; }
  .title { flex: 1; min-width: 0; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .time { flex: none; font-size: 12px; color: var(--text-muted); }
</style>
