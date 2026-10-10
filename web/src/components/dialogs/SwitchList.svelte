<script lang="ts">
  import { app } from "../../lib/app.svelte";
  import { shortPath } from "../../lib/format";
  import type { Project } from "../../lib/types";
  import Icon from "../Icon.svelte";

  let { projects, choose }: { projects: Project[]; choose: (project: Project) => void } = $props();
  let queryText = $state("");

  const shown = $derived(projects.filter((p) => p.path.toLowerCase().includes(queryText.trim().toLowerCase())));

  function markOf(project: Project): string {
    const live = app.live().filter((s) => s.path === project.path);
    if (live.some((s) => s.busy)) return "dot blink";
    if (live.some((s) => s.unread) || project.unread) return "dot";
    if (live.some((s) => s.tasks)) return "dot violet";
    return "";
  }
</script>

<div class="bar">
  <input class="field" placeholder="Search projects" aria-label="Search projects" bind:value={queryText} />
  <button class="btn secondary small new" title="Add project" aria-label="Add project" onclick={() => app.open("addproject")}><Icon name="plus" size={16} /></button>
</div>
<div class="rows">
  {#each shown as project (project.id)}
    {@const mark = markOf(project)}
    <button class="list-row" title={project.path} onclick={() => choose(project)}>
      <Icon name="folder" size={14} />
      <span class="path"><bdi>{shortPath(project.path, app.home)}</bdi></span>
      {#if mark}<span class={mark}></span>{/if}
    </button>
  {:else}
    <p class="empty">{queryText ? "No matches" : "No projects yet"}</p>
  {/each}
</div>

<style>
  .bar { display: flex; gap: 8px; align-items: center; margin-bottom: 8px; }
  .bar .field { flex: 1; min-width: 0; }
  .new { flex: none; padding: 6px 8px; }
  .rows { display: grid; gap: 2px; }
  .path { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl; text-align: left; }
  .list-row :global(svg) { flex: none; }
</style>
