<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { Model } from "../../lib/types";

  let models = $state<Model[] | null>(null);
  let scope = $state<string[]>([]);
  const provider = $derived(app.config.active);

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

<p class="soft">Models of the default provider that the model picker offers. With none ticked, all are offered.</p>
{#if !provider}<p class="empty">No provider yet</p>
{:else if models === null}<p class="empty">Loading models…</p>
{:else}
  <div class="rows">
    {#each models as model (model.id)}
      <label class="row">
        <input type="checkbox" checked={!scope.length || scope.includes(model.id)} onchange={() => toggle(model.id)} />
        <span class="mono">{model.id}</span>
      </label>
    {/each}
  </div>
{/if}

<style>
  .rows { max-height: 52vh; overflow-y: auto; }
  .row { display: flex; gap: 10px; padding: 4px 0; align-items: center; }
</style>
