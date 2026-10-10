<script lang="ts">
  import { get, post, query } from "../lib/api";
  import { fail } from "../lib/app.svelte";
  import type { Model } from "../lib/types";
  import ModelList from "./ModelList.svelte";
  import SubscriptionPick from "./SubscriptionPick.svelte";

  let { done }: { done: () => void } = $props();
  const kinds = [
    { id: "responses", name: "OpenAI Responses", path: "/responses", url: "https://api.openai.com/v1" },
    { id: "openai", name: "OpenAI Chat Completions", path: "/chat/completions", url: "https://api.openai.com/v1" },
    { id: "anthropic", name: "Anthropic", path: "/v1/messages", url: "https://api.anthropic.com" },
  ];
  let kind = $state("");
  let name = $state("");
  let baseURL = $state("");
  let key = $state("");
  let probing = $state(false);
  let mode = $state<"" | "api" | "subscription">("");

  async function pickKind(id: string) {
    kind = id;
    baseURL = kinds.find((k) => k.id === id)!.url;
    name = (await get<{ name: string }>("/api/providers/name" + query({ kind: id })).catch(() => ({ name: id }))).name;
  }

  const body = $derived({ kind, name, base_url: baseURL, api_key: key });

  async function save(model: string, effort: string) {
    try {
      await post("/api/providers", { ...body, model, effort });
      done();
    } catch (err) {
      fail(err);
    }
  }
</script>

{#if !mode}
  <div class="kinds">
    <button class="list-row kind" onclick={() => (mode = "api")}><span class="name">API</span><span class="soft">Your own key and base URL</span></button>
    <button class="list-row kind" onclick={() => (mode = "subscription")}><span class="name">Subscription</span><span class="soft">Claude, Codex or Antigravity through CLIProxyAPI</span></button>
  </div>
{:else if mode === "subscription"}
  <SubscriptionPick {done} back={() => (mode = "")} />
{:else if !kind}
  <div class="kinds">
    {#each kinds as k, i (k.id)}
      <button class="list-row kind" onclick={() => pickKind(k.id)}>
        <span class="mono n">{i + 1}</span><span class="name">{k.name}</span><span class="soft mono">{k.path}</span>
      </button>
    {/each}
  </div>
  <button class="btn ghost small" onclick={() => (mode = "")}>← Back</button>
{:else if !probing}
  <form class="form" onsubmit={(e) => { e.preventDefault(); probing = true; }}>
    <label><span class="label">Name</span><input class="field" bind:value={name} required /></label>
    <label><span class="label">Base URL</span><input class="field mono" bind:value={baseURL} required /></label>
    <label><span class="label">API key</span><input class="field mono" type="password" bind:value={key} autocomplete="off" /></label>
    <div class="actions">
      <button type="button" class="btn ghost" onclick={() => (kind = "")}>← Kinds</button>
      <button class="btn primary">Choose a model</button>
    </div>
  </form>
{:else}
  <ModelList
    load={() => post<Model[]>("/api/providers/probe", body)}
    efforts={(model) => post<string[]>("/api/providers/probe-efforts", { ...body, model })}
    choose={save}
  />
  <button class="btn ghost small" onclick={() => (probing = false)}>← Connection</button>
{/if}

<style>
  .kinds { display: grid; gap: 4px; }
  .kind { padding: 12px; border: 1px solid var(--raised); }
  .n { color: var(--accent); }
  .name { flex: 1; color: var(--text-strong); }
  .form { display: grid; gap: 12px; }
  label { display: grid; gap: 6px; }
  .actions { display: flex; justify-content: space-between; }
</style>
