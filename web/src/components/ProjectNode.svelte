<script lang="ts">
  import { app, fail } from "../lib/app.svelte";
  import { get, post, query } from "../lib/api";
  import { ago, shortPath } from "../lib/format";
  import { newSession, openSession } from "../lib/actions";
  import { openPanel } from "../lib/panels";
  import { tree } from "../lib/sidebar.svelte";
  import type { Project, SessionRow } from "../lib/types";
  import Icon from "./Icon.svelte";

  let { project, search, mark, choose }: {
    project: Project; search: string; mark: string; choose: (project: Project) => void;
  } = $props();

  async function archive(project: Project) {
    await post("/api/projects/archive", { path: project.path, archived: true }).catch(fail);
    app.toast("Project archived. Restore it in Settings, Archived projects");
  }

  let rows = $state<SessionRow[]>([]);
  let loaded = $state(false);
  const open = $derived(!!search || tree.isOpen(project.path));

  $effect(() => {
    void app.sessionsRev;
    const path = project.path, q = search;
    if (!open) return;
    const timer = setTimeout(() => {
      get<SessionRow[]>("/api/sessions" + query({ path, q }))
        .then((list) => ((rows = list), (loaded = true)))
        .catch(fail);
    }, q ? 200 : 0);
    return () => clearTimeout(timer);
  });
</script>

{#if !search || !loaded || rows.length}
  <div class="project">
    <div class="list-row head" class:active={project.path === app.project}>
      <button class="chevron" class:open onclick={() => tree.toggle(project.path)} title={open ? "Collapse" : "Expand"} aria-label="Toggle sessions"><Icon name="chevron" size={12} /></button>
      <button class="name" onclick={() => (tree.open(project.path), choose(project))} title={project.path}>
        <Icon name="folder" size={14} />
        <span class="path"><bdi>{shortPath(project.path, app.home)}</bdi></span>
        {#if mark}<span class="dot {mark}" class:blink={mark !== "unread"}></span>{/if}
      </button>
      <button class="add" onclick={() => openPanel("project", undefined, project.path)} title="Project settings" aria-label="Project settings"><Icon name="gear" size={14} /></button>
      <button class="add" onclick={() => archive(project)} title="Archive project" aria-label="Archive project"><Icon name="archive" size={14} /></button>
      <button class="add" onclick={() => (tree.open(project.path), newSession(project.path))} title="New session (Alt+N)" aria-label="New session"><Icon name="plus" size={14} /></button>
    </div>
    {#if open}
      {#each rows as row (row.id)}
        {@const state = app.sessions[row.id]?.state}
        <button class="list-row session" class:active={app.chat?.session === row.id} onclick={() => openSession(row.id)} title={row.title}>
          {#if state?.busy}<span class="dot blink"></span>
          {:else if state ? state.unread : row.unread}<span class="dot"></span>{/if}
          <span class="title">{row.title || "Untitled"}</span>
          <span class="time">{ago(row.updated_at)}</span>
        </button>
      {:else}
        {#if loaded && !search}<p class="empty">No sessions yet</p>{/if}
      {/each}
    {/if}
  </div>
{/if}

<style>
  .head { padding: 0; color: var(--text-dim); }
  .head.active { color: var(--text-strong); background: transparent; box-shadow: none; }
  .head button { background: none; border: 0; color: inherit; cursor: pointer; display: flex; align-items: center; gap: 8px; padding: 6px 4px; }
  .chevron { transition: transform 0.12s; padding-left: 8px !important; color: var(--text-muted); }
  .chevron.open { transform: rotate(90deg); }
  .name { flex: 1; min-width: 0; text-align: left; }
  .name span:not(.dot) { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .path { direction: rtl; text-align: left; }
  .name :global(svg) { flex: none; }
  .add { opacity: 0; padding-right: 10px !important; color: var(--text-muted); }
  @media (max-width: 700px) { .head button { min-height: 44px; } }
  .head:hover .add, .add:focus-visible { opacity: 1; }
  .add:hover { color: var(--accent); }
  .dot.tasks { background: var(--text-soft); }
  .session { width: calc(100% - 24px); margin-left: 24px; padding: 7px 10px; }
  .title { flex: 1; min-width: 0; color: var(--text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .time { flex: none; font-size: 12px; color: var(--text-muted); }
  .empty { padding: 4px 34px; margin: 0; font-size: 12px; color: var(--text-muted); }
</style>
