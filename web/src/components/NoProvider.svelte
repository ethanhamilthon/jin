<script lang="ts">
  import { openPanel } from "../lib/panels";
</script>

<div class="banner" role="status">
  <span>No enabled provider.</span>
  <button class="btn small" onclick={() => openPanel("settings", "providers")}>Open Providers</button>
</div>

<style>
  .banner {
    display: flex; gap: 8px; align-items: center; padding: 8px 12px; font-size: 13px;
    background: var(--card); border: 1px solid var(--warn); border-left-width: 2px; border-radius: var(--radius);
  }
  .banner span { flex: 1; }
</style>
