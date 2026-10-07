<script lang="ts">
  import { app, fail } from "../lib/app.svelte";
  import { get } from "../lib/api";
  import { openSession, newSession } from "../lib/actions";
  import type { Project } from "../lib/types";
  import Icon from "./Icon.svelte";

  let projects = $state<Project[]>([]);

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
    const shown = app.panes.findIndex((p) => app.sessions[p.session]?.state.path === project.path);
    if (shown >= 0) return void (app.focused = shown);
    if (project.last_session) await openSession(project.last_session);
    else await newSession(project.path);
  }
</script>

<section>
  <div class="head">
    <span class="label">Projects</span>
    <button class="btn ghost small" onclick={() => app.open("project")} title="Add a project"><Icon name="plus" /></button>
  </div>
  <div class="rows">
    {#each projects as project (project.id)}
      {@const state = mark(project.path)}
      <button class="list-row" class:active={project.path === app.project} onclick={() => choose(project)} title={project.path}>
        <Icon name="folder" size={14} />
        <span class="name">{project.name}</span>
        {#if state}<span class="dot {state}" class:blink={state !== "unread"}></span>{/if}
      </button>
    {/each}
  </div>
</section>

<style>
  section { padding: 12px 10px 8px; border-bottom: 1px solid var(--raised); max-height: 40%; display: flex; flex-direction: column; }
  .head { display: flex; align-items: center; justify-content: space-between; padding: 0 4px 6px; }
  .rows { overflow-y: auto; }
  .name { flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .list-row { color: var(--text-dim); padding: 6px 10px; }
  .list-row.active { color: var(--text-strong); }
  .dot.tasks { background: var(--text-soft); }
</style>
