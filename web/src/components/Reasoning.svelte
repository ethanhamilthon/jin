<script lang="ts">
  let { text, open = false }: { text: string; open?: boolean } = $props();
  let expanded = $state(false);
  const shown = $derived(open || expanded);
</script>

<div class="reasoning">
  <button class="label toggle" onclick={() => (expanded = !expanded)}>[ {shown ? "−" : "+"} reasoning ]</button>
  {#if shown}<div class="text">{text}</div>{/if}
</div>

<style>
  .toggle { background: none; border: 0; padding: 0; cursor: pointer; }
  .toggle:hover { color: var(--text-dim); }
  .text { margin-top: 6px; color: var(--text-soft); font-size: 13px; white-space: pre-wrap; border-left: 1px solid var(--line); padding-left: 12px; }
</style>
