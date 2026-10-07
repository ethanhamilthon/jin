<script lang="ts">
  import { app, fail } from "../lib/app.svelte";
  import { get, query } from "../lib/api";
  import { ago, baseName } from "../lib/format";
  import { newSession, openSession } from "../lib/actions";
  import type { SessionRow } from "../lib/types";
  import Icon from "./Icon.svelte";

  let rows = $state<SessionRow[]>([]);
  let search = $state("");
  let all = $state(false);

  $effect(() => {
    void app.sessionsRev;
    const path = app.project, q = search, everywhere = all;
    const timer = setTimeout(() => {
      get<SessionRow[]>("/api/sessions" + query({ path, q, all: everywhere ? "1" : "" }))
        .then((list) => (rows = list))
        .catch(fail);
    }, q ? 200 : 0);
    return () => clearTimeout(timer);
  });

  function live(id: string) {
    return app.sessions[id]?.state;
  }
</script>

<section>
  <div class="head">
    <span class="label">Sessions</span>
    <button class="btn secondary small" onclick={() => newSession(app.project)}><Icon name="plus" size={13} />New</button>
  </div>
  <div class="search">
    <input class="field" placeholder="Search titles and messages" bind:value={search} />
    <label class="soft"><input type="checkbox" bind:checked={all} /> all projects</label>
  </div>
  <div class="rows">
    {#each rows as row (row.id)}
      {@const state = live(row.id)}
      <button class="list-row" class:active={app.pane.session === row.id} onclick={() => openSession(row.id)}>
        <span class="text">
          <span class="title">{row.title || "Untitled"}</span>
          <span class="meta">{ago(row.updated_at)}{all ? " · " + baseName(row.path) : ""}{row.model ? " · " + row.model : ""}</span>
          {#if row.snippet}<span class="meta snippet">{row.snippet}</span>{/if}
        </span>
        {#if state?.busy}<span class="dot blink"></span>
        {:else if state ? state.unread : row.unread}<span class="dot"></span>{/if}
      </button>
    {:else}
      <p class="empty">{search ? "Nothing found" : "No sessions yet"}</p>
    {/each}
  </div>
</section>

<style>
  section { flex: 1; min-height: 0; display: flex; flex-direction: column; padding: 12px 10px; }
  .head { display: flex; align-items: center; justify-content: space-between; padding: 0 4px 8px; }
  .search { display: grid; gap: 6px; padding: 0 4px 8px; }
  .search label { font-size: 12px; display: flex; gap: 6px; align-items: center; }
  .rows { overflow-y: auto; flex: 1; }
  .text { flex: 1; min-width: 0; display: grid; }
  .title { color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .meta { font-size: 12px; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .snippet { color: var(--text-soft); }
</style>
