<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { shortPath } from "../../lib/format";
  import type { Project } from "../../lib/types";

  let archived = $state<Project[]>([]);

  $effect(() => {
    void app.projectsRev;
    get<Project[]>("/api/projects").then((list) => (archived = list.filter((p) => p.archived))).catch(fail);
  });

  const restore = (p: Project) => post("/api/projects/archive", { path: p.path, archived: false }).catch(fail);
</script>

<div class="set-hd"><span class="label">Archived projects</span><span class="hint">hidden from the sidebar; sessions are kept</span></div>
{#each archived as p (p.id)}
  <div class="set-row">
    <div class="set-text">
      <span class="set-name">{p.name}</span>
      <span class="set-desc" title={p.path}>{shortPath(p.path, app.home)} · {p.sessions} session{p.sessions === 1 ? "" : "s"}</span>
    </div>
    <div class="set-ctl"><button class="btn small" onclick={() => restore(p)}>Restore</button></div>
  </div>
{:else}<p class="empty">No archived projects</p>{/each}
