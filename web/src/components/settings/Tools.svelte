<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";

  const toggle = (name: string, enabled: boolean) => post("/api/settings/tools", { name, enabled }).catch(fail);
</script>

<p class="soft">Tools of the agent. Changes reach new sessions.</p>
{#each app.config.tools as tool (tool.name)}
  <label class="row">
    <input type="checkbox" checked={tool.enabled} onchange={(e) => toggle(tool.name, e.currentTarget.checked)} />
    <span class="mono name">{tool.name}</span>
    <span class="soft">{tool.description}</span>
  </label>
{/each}

<style>
  .row { display: flex; gap: 12px; align-items: baseline; padding: 6px 0; border-bottom: 1px solid var(--raised); }
  .name { color: var(--text-strong); min-width: 80px; }
  .soft { font-size: 13px; }
</style>
