<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import { prettyModel } from "../../lib/format";
  import type { Model } from "../../lib/types";
  import Switch from "../Switch.svelte";

  let models = $state<Model[] | null>(null);
  let scope = $state<string[]>([]);
  let filter = $state("");
  let selected = $state("");
  const provider = $derived(selected || app.config.active);
  let loadRevision = 0;
  const shown = $derived((models ?? []).filter((m) => m.id.toLowerCase().includes(filter.trim().toLowerCase())));

  async function load() {
    try {
      const source = provider; const revision = ++loadRevision;
      const [list, selectedScope] = await Promise.all([get<Model[]>(`/api/providers/${source}/models?all=1`), get<string[]>("/api/settings/scope" + query({ provider: source }))]);
      if (source !== provider || revision !== loadRevision) return;
      models = list; scope = selectedScope;
    } catch (err) {
      fail(err);
    }
  }

  $effect(() => {
    if (provider) load();
  });

  async function toggle(model: string) {
    await post("/api/settings/scope", { provider, model }).catch(fail);
    scope = await get<string[]>("/api/settings/scope" + query({ provider }));
  }
</script>

<div class="set-hd"><span class="label">Models</span></div>
<div class="set-pad"><select aria-label="Provider model scope" value={provider} onchange={(e) => { selected = e.currentTarget.value; models = null; }}>
  {#each app.config.providers.filter((p) => p.enabled !== false) as source (source.id)}<option value={source.id}>{source.name}</option>{/each}
</select></div>
{#if !provider}<p class="empty">No provider yet</p>
{:else if models === null}<p class="empty">Loading models…</p>
{:else}
  <div class="set-pad"><input class="field" placeholder="Filter models" bind:value={filter} /></div>
  {#each shown as model (model.id)}
    <div class="set-row">
      <div class="set-text"><span class="set-name mono">{prettyModel(model.id)}</span></div>
      <div class="set-ctl"><Switch label={prettyModel(model.id)} checked={!scope.length || scope.includes(model.id)} onchange={() => toggle(model.id)} /></div>
    </div>
  {:else}<p class="empty">No model matches</p>{/each}
  <p class="set-note">With none switched on, all models are offered.</p>
{/if}
