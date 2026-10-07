<script lang="ts">
  import { app } from "../lib/app.svelte";
  import Icon from "./Icon.svelte";
  import { toggleTools, toolsShown } from "../lib/fold";

  const shown = $derived(toolsShown(app.config.fold));
</script>

<button class:on={shown} aria-pressed={shown} title="Tool outputs in the chat (Ctrl+O)" aria-label="Tool outputs" onclick={toggleTools}><Icon name="terminal" size={14} /></button>

<style>
  button {
    background: transparent; border: 1px solid transparent; border-radius: 2px; cursor: pointer;
    padding: 3px 6px; display: inline-flex; color: var(--text-muted);
  }
  button:hover { color: var(--text-dim); }
  button.on { color: var(--accent); border-color: var(--accent-deep); }
</style>
