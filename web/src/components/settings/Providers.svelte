<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import ModelList from "../ModelList.svelte";
  import type { Model } from "../../lib/types";
  import type { Provider } from "../../lib/types";
  import Switch from "../Switch.svelte";
  import ProviderForm from "../ProviderForm.svelte";
  import SubscriptionAccount from "./SubscriptionAccount.svelte";

  let adding = $state(false);
  let confirm = $state("");
  let picking = $state<Provider | null>(null);
  async function setDefault(model: string, effort: string) {
    if (!picking) return;
    try { await post(`/api/providers/${picking.id}/activate`, { model, effort }); picking = null; }
    catch (error) { fail(error); }
  }
  async function enable(provider: Provider, enabled: boolean) {
    await post(`/api/providers/${provider.id}/enabled`, { enabled }).catch(fail);
  }
  async function remove(id: string) {
    await post(`/api/providers/${id}/delete`).catch(fail); confirm = "";
  }
</script>

{#if picking}
  {@const chosen = picking}
  <div class="set-pad"><ModelList load={() => get<Model[]>(`/api/providers/${chosen.id}/models`)} efforts={(model) => get<string[]>(`/api/providers/${chosen.id}/efforts` + query({ model }))} choose={setDefault} /></div>
  <button class="btn ghost small back" onclick={() => (picking = null)}>← Providers</button>
{:else if adding}
  <div class="set-pad"><ProviderForm done={() => (adding = false)} /></div>
  <button class="btn ghost small back" onclick={() => (adding = false)}>← Providers</button>
{:else}
  {#each app.config.providers as provider (provider.id)}
    <div class="set-row">
      <div class="set-text"><span class="set-name">{provider.name}</span><span class="set-desc">{provider.source === "cliproxy" ? "CLIProxyAPI subscription" : `${provider.kind} · ${provider.base_url}`}</span></div>
      <div class="set-ctl">
        <button class="btn small" disabled={provider.enabled === false} onclick={() => (picking = provider)}>{provider.id === app.config.active ? "Default" : "Set default"}</button>
        {#if confirm === provider.id}
          <button class="btn danger small" onclick={() => remove(provider.id)}>Delete</button>
          <button class="btn ghost small" onclick={() => (confirm = "")}>Keep</button>
        {:else}<button class="btn ghost small" onclick={() => (confirm = provider.id)}>Delete</button>{/if}
        <Switch label={provider.name} checked={provider.enabled !== false} onchange={(value) => enable(provider, value)} />
      </div>
    </div>
    {#if provider.source === "cliproxy" && provider.profile}<SubscriptionAccount profile={provider.profile} />{/if}
  {:else}<p class="empty">No providers yet</p>{/each}
  <div class="set-pad"><button class="btn" onclick={() => (adding = true)}>Add a provider</button></div>
  <p class="set-note">Enabled providers share the model picker. Turning one off lets its current request finish.</p>
{/if}

<style>.back { margin: 0 16px 16px; }</style>
