<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { ago, shortID } from "../../lib/format";
  import type { Task } from "../../lib/types";
  import Dialog from "../Dialog.svelte";

  let tasks = $state<Task[]>([]);
  let shown = $state("");
  let output = $state("");

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
    {#each tasks as task (task.id)}
      <div class="row" class:active={task.id === shown}>
        <span class="pill" class:dim={task.status !== "running"}>{task.status}</span>
        <button class="cmd mono" onclick={() => show(task.id)} title={task.command}>{task.id} · {task.command}</button>
        <span class="soft meta">{ago(task.started)} · {shortID(task.owner)}</span>
        {#if task.status === "running"}<button class="btn danger small" onclick={() => stop(task.id)}>Stop</button>{/if}
      </div>
    {:else}<p class="empty">No background tasks</p>{/each}
  </div>
  {#if shown}
    <div class="head"><span class="label">[ output · {shown} ]</span><button class="btn ghost small" onclick={() => show(shown)}>Refresh</button></div>
    <pre class="out">{output}</pre>
  {/if}
</Dialog>

<style>
  .rows { display: grid; gap: 4px; }
  .row { display: flex; gap: 10px; align-items: center; padding: 6px 8px; border: 1px solid var(--raised); }
  .row.active { border-color: var(--accent-deep); }
  .cmd { flex: 1; text-align: left; background: none; border: 0; color: var(--text-strong); cursor: pointer; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
  .meta { font-size: 12px; white-space: nowrap; }
  .head { display: flex; justify-content: space-between; align-items: center; margin: 14px 0 6px; }
  .out { margin: 0; background: var(--void); border: 1px solid var(--raised); padding: 10px 12px; font: 12px/1.5 var(--mono); max-height: 40vh; overflow: auto; white-space: pre-wrap; color: var(--text-dim); }
</style>
