<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { get, query } from "../lib/api";
  import { focus } from "../lib/focus";
  import { tokens } from "../lib/format";
  import { providerModelKey, type Catalog, type ProviderModel } from "../lib/provider-model";

  let { provider = "", current = "", picked = $bindable(null), choose }: { provider?: string; current?: string; picked?: ProviderModel | null; choose: (model: ProviderModel, effort: string) => void } = $props();
  let catalog = $state<Catalog | null>(null);
  let error = $state("");
  let filter = $state("");
  let efforts = $state<string[] | null>(null);
  const revision = $derived(app.config.providers);
  const shown = $derived((catalog?.models ?? []).filter((m) => `${m.id} ${m.provider_name}`.toLowerCase().includes(filter.toLowerCase())));

  $effect(() => {
    void revision;
    let active = true;
    catalog = null; error = "";
    get<Catalog>("/api/models").then((value) => { if (active) catalog = value; }).catch((e) => { if (active) error = String(e.message ?? e); });
    return () => { active = false; };
  });

  async function pick(model: ProviderModel) {
    picked = model; efforts = null;
    const list = await get<string[]>(`/api/providers/${model.provider}/efforts` + query({ model: model.id })).catch(() => []);
    if (picked === model) efforts = [...new Set(list.filter(Boolean))];
  }
</script>

{#if picked}
  <p class="label">[ {picked.id} ]</p>
  {#if efforts === null}<p class="empty">Loading…</p>
  {:else}
    {#each ["", ...efforts] as effort (effort)}
      <button class="list-row" onclick={() => picked && choose(picked, effort)}>{effort || "Default"}</button>
    {/each}
  {/if}
{:else if error}<p class="empty">{error}</p>
{:else if catalog === null}<p class="empty">Loading models…</p>
{:else}
  <input class="field" placeholder="Filter models or providers" bind:value={filter} use:focus />
  <div class="rows">
    {#each shown as model (providerModelKey(model))}
      <button class="list-row" class:active={model.provider === provider && model.id === current} onclick={() => pick(model)}>
        <span class="mono id">{model.id}</span><span class="soft source">{model.provider_name}{#if model.context}<small>{tokens(model.context)} ctx</small>{/if}{#if model.input || model.output}<small>${(model.input ?? 0).toFixed(2)} / ${(model.output ?? 0).toFixed(2)}</small>{/if}</span>
      </button>
    {:else}<p class="empty">No models. Enable a provider or connect a subscription in Settings.</p>{/each}
  </div>
  {#each catalog.errors as failure (failure.provider)}<p class="soft failure">{failure.provider}: {failure.message}</p>{/each}
{/if}

<style>
  .rows { margin: 10px 0; display: grid; gap: 2px; max-height: 50vh; overflow-y: auto; }
  .id { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; }
  .source { font-size: 11px; display: grid; text-align: right; }
  .failure { font-size: 12px; }
</style>
