<script lang="ts">
  import type { Item } from "../lib/composer";

  let { items, selected, choose }: { items: Item[]; selected: number; choose: (item: Item) => void } = $props();
</script>

<div class="menu" role="listbox">
  {#each items as item, i (item.insert)}
    <button
      class="row" class:on={i === selected} role="option" aria-selected={i === selected}
      onmousedown={(e) => { e.preventDefault(); choose(item); }}
    >
      <span class="name mono">{item.label}</span>
      {#if item.detail}<span class="detail">{item.detail}</span>{/if}
    </button>
  {/each}
</div>

<style>
  .menu {
    position: absolute; left: 0; right: 0; bottom: calc(100% + 6px); z-index: 10; max-height: 300px; overflow-y: auto;
    background: var(--hover); border: 1px solid var(--line); border-radius: var(--radius); padding: 4px;
  }
  .row { display: flex; gap: 12px; width: 100%; text-align: left; border: 0; background: transparent; padding: 6px 10px; border-radius: 2px; cursor: pointer; }
  .row.on { background: var(--raised); box-shadow: inset 2px 0 0 var(--accent); }
  .name { color: var(--text-strong); white-space: nowrap; }
  .detail { color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 12px; }
</style>
