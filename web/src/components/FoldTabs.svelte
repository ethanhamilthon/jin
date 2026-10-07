<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { toggleTools, toolsShown } from "../lib/fold";

  const shown = $derived(toolsShown(app.config.fold));
</script>

<button class:on={shown} aria-pressed={shown} title="Show tool calls and outputs in the chat (Ctrl+O)" onclick={toggleTools}>tool outputs</button>

<style>
  button {
    background: transparent; border: 1px solid transparent; border-radius: 2px; cursor: pointer;
    padding: 0 6px; color: var(--text-muted); font: inherit; font-size: 11px;
  }
  button:hover { color: var(--text-dim); }
  button.on { color: var(--accent); border-color: var(--accent-deep); }
</style>
