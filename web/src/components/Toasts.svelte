<script lang="ts">
  import { app } from "../lib/app.svelte";
</script>

<div class="toasts">
  {#each app.toasts as toast (toast.id)}
    <div class="toast" class:error={toast.error}>{toast.text}</div>
  {/each}
</div>

<style>
  .toasts { position: fixed; right: 16px; bottom: 16px; display: grid; gap: 8px; z-index: 50; max-width: 420px; }
  .toast {
    background: var(--hover); border: 1px solid var(--line); border-radius: var(--radius);
    padding: 10px 14px; color: var(--text); white-space: pre-wrap;
  }
  .toast.error { border-color: var(--error); color: var(--text-strong); }
</style>
