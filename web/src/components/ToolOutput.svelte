<script lang="ts">
  import type { Line } from "../lib/types";

  let { lines, tool }: { lines: Line[]; tool: string } = $props();
  const fold = 8;
  let expanded = $state(false);
  const hidden = $derived(Math.max(0, lines.length - fold));
  const shown = $derived(expanded ? lines : lines.slice(hidden));
  const diff = $derived(tool === "edit" || tool === "write");
</script>

{#if lines.length && !(lines.length === 1 && !lines[0].text)}
  <div class="out" class:diff>
    {#if hidden && !expanded}
      <button class="more" onclick={() => (expanded = true)}>… {hidden} more lines</button>
    {/if}
    {#each shown as line, i (i)}
      <div class="line {line.op === '+' ? 'add' : line.op === '-' ? 'del' : ''}"><span class="op">{line.op}</span>{line.text}</div>
    {/each}
    {#if expanded && hidden}<button class="more" onclick={() => (expanded = false)}>show less</button>{/if}
  </div>
{/if}

<style>
  .out {
    font-family: var(--mono); font-size: 12px; line-height: 1.55; background: var(--void);
    border: 1px solid var(--raised); border-radius: var(--radius); padding: 6px 0; overflow-x: auto;
    max-height: 480px; overflow-y: auto; margin-top: -6px;
  }
  .line { white-space: pre; padding: 0 12px; color: var(--text-dim); min-width: max-content; }
  .op { display: inline-block; width: 16px; color: var(--text-muted); user-select: none; }
  .add { background: color-mix(in srgb, var(--ok) 12%, transparent); color: var(--text); }
  .add .op { color: var(--ok); }
  .del { background: color-mix(in srgb, var(--error) 12%, transparent); color: var(--text); }
  .del .op { color: var(--error); }
  .more { background: none; border: 0; color: var(--text-muted); cursor: pointer; padding: 0 12px 2px; font: inherit; }
  .more:hover { color: var(--accent); }
</style>
