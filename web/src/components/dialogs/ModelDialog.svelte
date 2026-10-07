<script lang="ts">
  import { app } from "../../lib/app.svelte";
  import { act } from "../../lib/actions";
  import { get, query } from "../../lib/api";
  import type { Model } from "../../lib/types";
  import Dialog from "../Dialog.svelte";
  import ModelList from "../ModelList.svelte";

  let { id }: { id: string } = $props();
  const state = $derived(app.sessions[id]?.state);

  async function choose(model: string, effort: string) {
    if (await act(id, "model", { model, effort })) app.dialog = null;
  }
</script>

<Dialog title="Model" label={state?.provider ?? "provider"}>
  {#if state?.provider_missing || !state}
    <p class="soft">This session has no working provider. Pick one in <button class="btn small" onclick={() => app.open("providers")}>Providers</button></p>
  {:else}
    <ModelList
      current={state.model}
      load={() => get<Model[]>(`/api/providers/${state.provider}/models`)}
      efforts={(model) => get<string[]>(`/api/providers/${state.provider}/efforts` + query({ model }))}
      {choose}
    />
    <p class="more"><button class="btn ghost small" onclick={() => app.open("providers")}>Providers</button></p>
  {/if}
</Dialog>

<style>
  .more { margin: 12px 0 0; }
</style>
