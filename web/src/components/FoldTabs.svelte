<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { folds, setFold } from "../lib/fold";
</script>

<div class="tabs" role="tablist" aria-label="What the chat shows (Ctrl+O)">
  {#each folds as fold (fold.value)}
    <button
      role="tab" aria-selected={app.config.fold === fold.value} class:on={app.config.fold === fold.value}
      title={fold.title + " · Ctrl+O"} onclick={() => setFold(fold.value)}
    >{fold.label}</button>
  {/each}
</div>

<style>
  .tabs { display: flex; gap: 2px; }
  button {
    background: transparent; border: 1px solid transparent; border-radius: 2px; cursor: pointer;
    padding: 2px 8px; color: var(--text-muted); font: 500 10px/1.5 var(--sans); letter-spacing: 0.12em; text-transform: uppercase;
  }
  button:hover { color: var(--text-dim); }
  button.on { color: var(--accent); border-color: var(--accent-deep); }
</style>
