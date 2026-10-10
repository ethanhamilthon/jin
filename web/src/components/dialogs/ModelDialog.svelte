<script lang="ts">
  import { app } from "../../lib/app.svelte";
  import { act } from "../../lib/actions";
  import type { ProviderModel } from "../../lib/provider-model";
  import Dialog from "../Dialog.svelte";
  import AllModels from "../AllModels.svelte";

  let { id }: { id: string } = $props();
  const session = $derived(app.sessions[id]?.state);
  let picked = $state<ProviderModel | null>(null);

  async function choose(model: ProviderModel, effort: string) {
    if (await act(id, "model", { provider: model.provider, model: model.id, effort })) app.dialog = null;
  }
</script>

<Dialog title={picked ? "Effort" : "Model"} label="all enabled providers">
  {#if session}
    <AllModels provider={session.provider} current={session.model} bind:picked {choose} />
  {/if}
</Dialog>
