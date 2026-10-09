<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";
  import Switch from "../Switch.svelte";
  const toggle = (name: string, enabled: boolean) => post("/api/settings/tools", { name, enabled }).catch(fail);
</script>

<div class="set-hd"><span class="label">Tools</span><span class="hint">changes reach new sessions</span></div>
{#each app.config.tools as tool (tool.name)}
  <div class="set-row">
    <div class="set-text"><span class="set-name mono">{tool.name}</span><span class="set-desc" title={tool.description}>{tool.description}</span></div>
    <div class="set-ctl"><Switch label={tool.name} checked={tool.enabled} onchange={(v) => toggle(tool.name, v)} /></div>
  </div>
{/each}
