<script lang="ts">
  import type { AttachedFile } from "../lib/draft";

  let { files, remove }: { files: AttachedFile[]; remove: (path: string) => void } = $props();
</script>

<div class="chips">
  {#each files as file (file.path)}
    <span class="chip mono" title={file.path}>
      <span class="name">{file.name}</span>
      <button onclick={() => remove(file.path)} aria-label="Remove {file.name}">✕</button>
    </span>
  {/each}
</div>

<style>
  .chips { display: flex; gap: 6px; flex-wrap: wrap; }
  .chip { display: inline-flex; align-items: center; gap: 6px; max-width: 240px; padding: 2px 4px 2px 8px; font-size: 12px; color: var(--text-dim); background: var(--card); border: 1px solid var(--raised); border-radius: var(--radius); }
  .name { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  button { border: 0; background: transparent; color: var(--text-muted); cursor: pointer; font-size: 10px; padding: 2px 4px; }
  button:hover { color: var(--error); }
</style>
