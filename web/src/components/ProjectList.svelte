<script lang="ts">
  import { app, fail } from "../lib/app.svelte";
  import { get } from "../lib/api";
  import { openSession, newSession } from "../lib/actions";
  import { focus } from "../lib/focus";
  import type { Project } from "../lib/types";
  import Icon from "./Icon.svelte";
  import ProjectNode from "./ProjectNode.svelte";

  let projects = $state<Project[]>([]);
  let searching = $state(false);
  let search = $state("");

  $effect(() => {
    void app.projectsRev, app.sessionsRev;
    get<Project[]>("/api/projects").then((list) => (projects = list)).catch(fail);
  });

  function mark(path: string) {
    const live = app.live().filter((s) => s.path === path);
    if (live.some((s) => s.busy)) return "working";
    if (live.some((s) => s.unread) || projects.find((p) => p.path === path)?.unread) return "unread";
    if (live.some((s) => s.tasks)) return "tasks";
    return "";
  }

  async function choose(project: Project) {
    app.project = project.path;
    const shown = app.panes.findIndex((p) => p.kind === "chat" && app.sessions[p.session]?.state.path === project.path);
    if (shown >= 0) return void (app.focused = shown);
    if (project.last_session) await openSession(project.last_session);
    else await newSession(project.path);
  }

  function toggleSearch() {
    searching = !searching;
    if (!searching) search = "";
  }
</script>

<section>
  <div class="head">
    <span class="label">Projects</span>
    <span class="tools">
      <button class="btn ghost small" class:on={searching} onclick={toggleSearch} title="Search sessions"><Icon name="search" /></button>
      <button class="btn ghost small" onclick={() => app.open("addproject")} title="Add a project"><Icon name="plus" /></button>
    </span>
  </div>
  {#if searching}
    <div class="search">
      <input class="field" placeholder="Search titles and messages" bind:value={search} use:focus onkeydown={(e) => e.key === "Escape" && toggleSearch()} />
    </div>
  {/if}
  <div class="rows">
    {#each projects as project (project.id)}
      <ProjectNode {project} {search} mark={mark(project.path)} {choose} />
    {/each}
  </div>
</section>

<style>
  section { flex: 1; min-height: 0; display: flex; flex-direction: column; padding: 12px 10px 8px; }
  .head { display: flex; align-items: center; justify-content: space-between; padding: 0 4px 6px; }
  .tools { display: flex; gap: 2px; }
  .on { color: var(--accent); }
  .search { padding: 0 4px 8px; }
  .rows { overflow-y: auto; flex: 1; }
</style>
