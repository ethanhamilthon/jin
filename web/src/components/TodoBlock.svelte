<script lang="ts">
  import type { TodoItem } from "../lib/types";

  let { todos }: { todos: TodoItem[] } = $props();
  const done = $derived(todos.filter((t) => t.status === "done").length);
  const marks = { pending: "○", in_progress: "◐", done: "●" };
</script>

<div class="todo">
  <div class="head"><span class="label">[ todo ]</span><span class="label">{done}/{todos.length}</span></div>
  <ul>
    {#each todos as item, i (i)}
      <li class={item.status}><span class="mark">{marks[item.status]}</span>{item.text}</li>
    {/each}
  </ul>
</div>

<style>
  .todo { background: var(--card); border: 1px solid var(--raised); padding: 10px 14px; max-height: 190px; overflow-y: auto; }
  .head { display: flex; justify-content: space-between; margin-bottom: 6px; }
  ul { list-style: none; margin: 0; padding: 0; display: grid; gap: 3px; }
  li { display: flex; gap: 10px; color: var(--text-dim); font-size: 13px; }
  .mark { color: var(--text-muted); width: 12px; flex: none; }
  .in_progress { color: var(--text-strong); }
  .in_progress .mark { color: var(--accent); }
  .done { color: var(--text-muted); text-decoration: line-through; }
</style>
