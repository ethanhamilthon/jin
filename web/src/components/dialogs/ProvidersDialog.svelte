<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { Model, Provider } from "../../lib/types";
  import Dialog from "../Dialog.svelte";
  import ModelList from "../ModelList.svelte";
  import ProviderForm from "../ProviderForm.svelte";

  let mode = $state<"list" | "add" | Provider>("list");
  let confirm = $state("");

  async function activate(provider: Provider, model: string, effort: string) {
    try {
      await post(`/api/providers/${provider.id}/activate`, { model, effort });
      mode = "list";
    } catch (err) {
      fail(err);
    }
  }

  async function remove(id: string) {
    await post(`/api/providers/${id}/delete`).catch(fail);
    confirm = "";
  }
</script>

<Dialog title={mode === "add" ? "Add a provider" : typeof mode === "object" ? mode.name : "Providers"} label="default for new sessions">
  {#if mode === "add"}
    <ProviderForm done={() => (mode = "list")} />
  {:else if typeof mode === "object"}
    {@const provider = mode}
    <ModelList
      load={() => get<Model[]>(`/api/providers/${provider.id}/models`)}
      efforts={(model) => get<string[]>(`/api/providers/${provider.id}/efforts` + query({ model }))}
      choose={(model, effort) => activate(provider, model, effort)}
    />
    <button class="btn ghost small" onclick={() => (mode = "list")}>← Providers</button>
  {:else}
    <div class="rows">
      {#each app.config.providers as p (p.id)}
        <div class="row">
          <span class="dot" class:off={p.id !== app.config.active}></span>
          <span class="text"><span class="name">{p.name}</span><span class="soft mono">{p.kind} · {p.base_url}</span></span>
          {#if confirm === p.id}
            <button class="btn danger small" onclick={() => remove(p.id)}>Delete</button>
            <button class="btn ghost small" onclick={() => (confirm = "")}>Keep</button>
          {:else}
            <button class="btn small" onclick={() => (mode = p)}>{p.id === app.config.active ? "Model" : "Use"}</button>
            <button class="btn ghost small" onclick={() => (confirm = p.id)}>Delete</button>
          {/if}
        </div>
      {:else}<p class="empty">No providers yet</p>{/each}
    </div>
    <button class="btn primary" onclick={() => (mode = "add")}>Add a provider</button>
  {/if}
</Dialog>

<style>
  .rows { display: grid; gap: 4px; margin-bottom: 14px; }
  .row { display: flex; align-items: center; gap: 10px; padding: 8px 10px; border: 1px solid var(--raised); }
  .dot.off { background: transparent; border: 1px solid var(--line); }
  .text { flex: 1; display: grid; min-width: 0; }
  .text .mono { font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .name { color: var(--text-strong); }
</style>
