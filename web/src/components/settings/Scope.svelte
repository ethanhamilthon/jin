<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { Model } from "../../lib/types";
  import Switch from "../Switch.svelte";

  let models = $state<Model[] | null>(null);
  let scope = $state<string[]>([]);
  let filter = $state("");
  const provider = $derived(app.config.active);
  const shown = $derived((models ?? []).filter((m) => m.id.toLowerCase().includes(filter.trim().toLowerCase())));

  async function load() {
    try {
      models = await get<Model[]>(`/api/providers/${provider}/models?all=1`);
      scope = await get<string[]>("/api/settings/scope" + query({ provider }));
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

<div class="set-hd"><span class="label">Models</span><span class="hint">default provider: {provider || "none"}</span></div>
{#if !provider}<p class="empty">No provider yet</p>
{:else if models === null}<p class="empty">Loading models…</p>
{:else}
  <div class="set-pad"><input class="field" placeholder="Filter models" bind:value={filter} /></div>
  {#each shown as model (model.id)}
    <div class="set-row">
      <div class="set-text"><span class="set-name mono">{model.id}</span></div>
      <div class="set-ctl"><Switch label={model.id} checked={!scope.length || scope.includes(model.id)} onchange={() => toggle(model.id)} /></div>
    </div>
  {:else}<p class="empty">No model matches</p>{/each}
  <p class="set-note">With none switched on, all models are offered.</p>
{/if}
