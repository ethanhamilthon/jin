<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { ago, shortID } from "../../lib/format";
  import type { Task } from "../../lib/types";
  import Dialog from "../Dialog.svelte";

  const maxAgeMs = 60 * 60 * 1000;

  let tasks = $state<Task[]>([]);
  let shown = $state("");
  let output = $state("");

  const isRecent = (task: Task) => Date.now() - Date.parse(task.started) <= maxAgeMs;
  const runningFirst = (a: Task, b: Task) => Number(b.status === "running") - Number(a.status === "running");
  const newestFirst = (a: Task, b: Task) => Date.parse(b.started) - Date.parse(a.started);

  const visible = $derived(tasks.filter(isRecent).sort((a, b) => runningFirst(a, b) || newestFirst(a, b)));
  const hidden = $derived(tasks.length - visible.length);

  $effect(() => {
    void app.tasksRev;
    get<Task[]>("/api/tasks").then((list) => (tasks = list)).catch(fail);
    if (shown) show(shown);
  });

  async function show(id: string) {
    shown = id;
    const res = await get<{ text: string; truncated: boolean }>(`/api/tasks/${id}/output`).catch((e) => ({ text: String(e), truncated: false }));
    output = (res.truncated ? "…\n" : "") + (res.text || "(no output yet)");
  }

  const stop = (id: string) => post(`/api/tasks/${id}/stop`).catch(fail);
</script>

<Dialog title="Background tasks" label="owned by jin" wide>
  <div class="rows">
    {#each visible as task (task.id)}
      <div class="row" class:active={task.id === shown} title={task.id}>
        <div class="text">
          <button class="cmd mono" onclick={() => show(task.id)} title={task.command}>{task.command}</button>
          <div class="meta">
            <span class="pill status-{task.status}">{task.status}</span>
            <span>{ago(task.started)} · {shortID(task.owner)}</span>
          </div>
        </div>
        {#if task.status === "running"}<button class="btn danger small" onclick={() => stop(task.id)}>Stop</button>{/if}
      </div>
    {:else}<p class="empty">No background tasks</p>{/each}
  </div>
  {#if hidden}<p class="hidden soft">{hidden} {hidden === 1 ? "task" : "tasks"} older than an hour hidden</p>{/if}
  {#if shown}
    <div class="head"><span class="label">[ output · {shown} ]</span><button class="btn ghost small" onclick={() => show(shown)}>Refresh</button></div>
    <pre class="out">{output}</pre>
  {/if}
</Dialog>

<style>
  .rows { display: grid; grid-template-columns: minmax(0, 1fr); gap: 4px; }
  .row { display: flex; gap: 12px; align-items: center; padding: 8px 10px; border: 1px solid var(--raised); }
  .row.active { border-color: var(--accent-deep); }
  .text { flex: 1; display: grid; gap: 4px; min-width: 0; }
  .cmd { width: 100%; min-width: 0; text-align: left; background: none; border: 0; padding: 0; color: var(--text-strong); cursor: pointer; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
  .meta { display: flex; align-items: center; gap: 8px; color: var(--text-muted); font-size: 11px; white-space: nowrap; }
  .pill.status-done { border-color: var(--ok); color: var(--ok); }
  .pill.status-failed { border-color: var(--error); color: var(--error); }
  .pill.status-stopped { border-color: var(--warn); color: var(--warn); }
  .hidden { margin: 8px 2px 0; font-size: 12px; }
  .head { display: flex; justify-content: space-between; align-items: center; margin: 14px 0 6px; }
  .out { margin: 0; background: var(--void); border: 1px solid var(--raised); padding: 10px 12px; font: 12px/1.5 var(--mono); max-height: 40vh; overflow: auto; white-space: pre-wrap; color: var(--text-dim); }
</style>
