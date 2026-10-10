<script lang="ts">
  import { app } from "../../lib/app.svelte";
  import { act } from "../../lib/actions";
  import type { ProviderModel } from "../../lib/provider-model";
  import Dialog from "../Dialog.svelte";
  import AllModels from "../AllModels.svelte";

  let { id }: { id: string } = $props();
  const state = $derived(app.sessions[id]?.state);

  async function choose(model: ProviderModel, effort: string) {
    if (await act(id, "model", { provider: model.provider, model: model.id, effort })) app.dialog = null;
  }
</script>

<Dialog title="Model" label="all enabled providers">
  {#if state}
    <AllModels provider={state.provider} current={state.model} {choose} />
    <p class="more"><button class="btn ghost small" onclick={() => app.open("providers")}>Providers</button></p>
  {/if}
</Dialog>

<style>
  .more { margin: 12px 0 0; }
</style>
