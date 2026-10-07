<script lang="ts">
  import { focus } from "../lib/focus";
  import { tokens } from "../lib/format";
  import type { Model } from "../lib/types";

  interface Props {
    load: () => Promise<Model[]>;
    efforts: (model: string) => Promise<string[]>;
    current?: string;
    choose: (model: string, effort: string) => void;
  }
  let { load, efforts: effortsOf, current = "", choose }: Props = $props();
  let models = $state<Model[] | null>(null);
  let error = $state("");
  let search = $state("");
  let picked = $state("");
  let efforts = $state<string[] | null>(null);

  $effect(() => {
    load().then((list) => (models = list)).catch((e) => (error = String(e.message ?? e)));
  });

  const shown = $derived((models ?? []).filter((m) => m.id.toLowerCase().includes(search.toLowerCase())));

  async function pick(model: string) {
    picked = model;
    efforts = null;
    efforts = await effortsOf(model).catch(() => []);
  }

  function facts(m: Model) {
    const parts: string[] = [];
    if (m.context) parts.push(tokens(m.context) + " ctx");
    if (m.input || m.output) parts.push(`$${(m.input ?? 0).toFixed(2)} / $${(m.output ?? 0).toFixed(2)}`);
    if (m.reasoning) parts.push("reasoning");
    if (m.vision) parts.push("vision");
    return parts.join(" · ");
  }
</script>

{#if picked}
  <p class="label">[ reasoning effort · {picked} ]</p>
  {#if efforts === null}<p class="empty">Loading…</p>{/if}
  <div class="rows">
    {#each ["", ...(efforts ?? [])] as effort (effort)}
      <button class="list-row" onclick={() => choose(picked, effort)}>{effort || "Default"}</button>
    {/each}
  </div>
  <button class="btn ghost small" onclick={() => (picked = "")}>← Models</button>
{:else if error}
  <p class="error">{error}</p>
{:else if models === null}
  <p class="empty">Loading models…</p>
{:else}
  <input class="field" placeholder="Filter models" bind:value={search} use:focus />
  <div class="rows">
    {#each shown as model (model.id)}
      <button class="list-row" class:active={model.id === current} onclick={() => pick(model.id)}>
        <span class="mono id">{model.id}</span><span class="soft facts">{facts(model)}</span>
      </button>
    {:else}<p class="empty">No models</p>{/each}
  </div>
{/if}

<style>
  .rows { margin: 10px 0; display: grid; gap: 2px; max-height: 50vh; overflow-y: auto; }
  .id { color: var(--text-strong); flex: 1; overflow: hidden; text-overflow: ellipsis; }
  .facts { font-size: 12px; white-space: nowrap; }
  .error { color: var(--error); }
</style>
