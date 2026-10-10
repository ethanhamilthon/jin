<script lang="ts">
  import type { Folders } from "../../lib/types";
  import Icon from "../Icon.svelte";

  let { folders, open }: { folders: Folders; open: (path: string) => void } = $props();
</script>

<div class="rows">
  {#if folders.parent}
    <button class="list-row" onclick={() => open(folders.parent)}><Icon name="back" size={14} /><span class="name">Parent folder</span></button>
  {/if}
  {#each folders.dirs as name (name)}
    <button class="list-row" onclick={() => open(`${folders.path.replace(/\/$/, "")}/${name}`)}>
      <Icon name="folder" size={14} /><span class="name">{name}</span><Icon name="chevron" size={14} />
    </button>
  {:else}<p class="empty">No folders here</p>{/each}
</div>

<style>
  .rows { display: grid; gap: 2px; max-height: 40vh; overflow-y: auto; }
  .name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .list-row :global(svg) { flex: none; }
</style>
