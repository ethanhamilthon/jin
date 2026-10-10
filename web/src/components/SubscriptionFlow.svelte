<script lang="ts">
  import { get, post, query } from "../lib/api";
  import { fail } from "../lib/app.svelte";
  import type { Model } from "../lib/types";
  import ModelList from "./ModelList.svelte";
  import SubscriptionInstall from "./SubscriptionInstall.svelte";
  import SubscriptionPick from "./SubscriptionPick.svelte";
  import SubscriptionSignIn from "./SubscriptionSignIn.svelte";

  let { done, back, profile = "" }: { done: () => void; back?: () => void; profile?: string } = $props();
  let step = $state<"install" | "profile" | "signin" | "model">("install");
  let chosen = $state("");
  const id = $derived("subscription-" + chosen);

  function installed() {
    chosen = profile;
    step = chosen ? "signin" : "profile";
  }
  function picked(name: string) {
    chosen = name;
    step = "signin";
  }
  async function activate(model: string, effort: string) {
    try { await post(`/api/providers/${id}/activate`, { model, effort }); done(); }
    catch (error) { fail(error); }
  }
</script>

{#if step === "install"}
  <SubscriptionInstall ready={installed} {back} />
{:else if step === "profile"}
  <SubscriptionPick {picked} {back} />
{:else if step === "signin"}
  <SubscriptionSignIn profile={chosen} signedIn={() => (step = "model")} />
{:else}
  <ModelList
    load={() => get<Model[]>(`/api/providers/${id}/models`)}
    efforts={(model) => get<string[]>(`/api/providers/${id}/efforts` + query({ model }))}
    choose={activate}
  />
{/if}
