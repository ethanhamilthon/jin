<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { ProviderModel } from "../../lib/provider-model";
  import type { TitleSettings } from "../../lib/types";
  import AllModels from "../AllModels.svelte";
  import Switch from "../Switch.svelte";

  const title = $derived(app.config.title);
  const target = $derived(title.model ? { provider: title.provider, model: title.model } : { provider: app.config.active, model: app.config.model });
  const providerName = $derived(app.config.providers.find((p) => p.id === title.provider)?.name ?? title.provider);
  let picking = $state(false);
  let efforts = $state<string[]>([]);
  let prompt = $state(app.config.title.prompt);

  $effect(() => {
    const { provider, model } = target;
    efforts = [];
    if (provider && model) {
      get<string[]>(`/api/providers/${provider}/efforts` + query({ model })).then((list) => (efforts = list.filter(Boolean))).catch(() => {});
    }
  });

  async function save(change: Partial<TitleSettings>) {
    await post("/api/settings/title", { ...title, ...change }).catch(fail);
  }

  function choose(model: ProviderModel, effort: string) {
    picking = false;
    save({ provider: model.provider, model: model.id, effort });
  }
</script>

<div class="set-hd"><span class="label">Session titles</span><span class="hint">named by a model after the first message</span></div>
<div class="set-row">
  <div class="set-text"><span class="set-name">Model</span><span class="set-desc">{title.model ? `${providerName} · ${title.model}` : "The session's model"}</span></div>
  <div class="set-ctl">
    {#if title.model}<button class="btn ghost small" onclick={() => save({ provider: "", model: "" })}>Session model</button>{/if}
    <button class="btn small" class:on={picking} onclick={() => (picking = !picking)}>{picking ? "Close" : "Change"}</button>
  </div>
</div>
{#if picking}
  <div class="set-pad"><AllModels provider={title.provider} current={title.model} {choose} /></div>
{/if}
<div class="set-row">
  <div class="set-text"><span class="set-name">Effort</span><span class="set-desc">for the model above; Default is the provider's</span></div>
  <div class="set-ctl">
    <select aria-label="Title effort" value={title.effort} onchange={(e) => save({ effort: e.currentTarget.value })}>
      {#each ["", ...efforts] as level (level)}<option value={level}>{level || "Default"}</option>{/each}
    </select>
  </div>
</div>
<div class="set-row">
  <div class="set-text"><span class="set-name">After</span><span class="set-desc">titled when you have sent this many messages; 0 turns it off</span></div>
  <div class="set-ctl"><input class="field num" type="number" min="0" max="50" aria-label="Messages before the title" value={title.after} onchange={(e) => save({ after: Number(e.currentTarget.value) })} /></div>
</div>
<div class="set-row">
  <div class="set-text"><span class="set-name">Rename at message 4</span><span class="set-desc">names the session again once you have sent four messages</span></div>
  <div class="set-ctl"><Switch label="Rename at message 4" checked={title.refresh} onchange={(v) => save({ refresh: v })} /></div>
</div>
<div class="set-pad"><textarea class="field mono" rows="4" aria-label="Title prompt" spellcheck="false" bind:value={prompt}></textarea></div>
<div class="set-pad actions">
  <span class="set-note">Sent after the conversation. Empty uses the built-in prompt.</span>
  <button class="btn small" onclick={() => save({ prompt })}>Save prompt</button>
</div>

<style>
  .num { width: 84px; }
  textarea { min-height: 96px; resize: vertical; font-size: 12.5px; line-height: 1.55; display: block; }
  .actions { display: flex; align-items: center; gap: 10px; }
  .actions .set-note { flex: 1; padding: 0; }
  .on { color: var(--accent); }
</style>
