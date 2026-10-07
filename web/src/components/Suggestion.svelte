<script lang="ts">
  let { text, send, edit }: { text: string; send: () => void; edit: () => void } = $props();
</script>

<div class="suggest">
  <span class="mark">›</span>
  <span class="text" title={text}>{text}</span>
  <button class="btn secondary small" onclick={send}>Send</button>
  <button class="btn ghost small" onclick={edit}>Edit</button>
</div>

<style>
  .suggest { display: flex; align-items: center; gap: 10px; padding: 0 2px; min-width: 0; }
  .mark { color: var(--accent); }
  .text { flex: 1; color: var(--text-soft); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
