<script lang="ts">
  import type { Intro } from "../lib/types";

  let { intro, loading }: { intro: Intro; loading: string[] } = $props();
  const rows = $derived([
    ["Tools", intro.tools.join(", ") || "no tools enabled"],
    ["Context", intro.context?.join("\n") || "no AGENTS.md found"],
    ["Hooks", intro.hooks?.join(", ") || "no hooks enabled"],
  ]);
</script>

<div class="intro">
  <p class="label">[ jin {intro.version} ]</p>
  <h1 class="serif">What should we <span class="accent-word">build</span> today?</h1>
  <dl>
    {#each rows as [name, value] (name)}
      <dt class="label">{name}</dt><dd>{value}</dd>
    {/each}
    <dt class="label">Prompts</dt>
    <dd>
      {#each intro.prompts ?? [] as name (name)}
        <span class="prompt" class:loading={loading.includes(name)}>#{name}</span>
      {:else}no prompts enabled{/each}
    </dd>
    {#if intro.system_prompt}<dt class="label">System prompt</dt><dd>custom ({intro.system_prompt})</dd>{/if}
    {#if intro.update}<dt class="label">Update</dt><dd>jin {intro.update} is available · run <code>jin update</code> in a shell</dd>{/if}
  </dl>
</div>

<style>
  .intro { padding: 24px 0 8px; }
  h1 { font-size: 40px; line-height: 1.05; letter-spacing: -0.025em; margin: 6px 0 20px; }
  dl { display: grid; grid-template-columns: 120px 1fr; gap: 8px 16px; margin: 0; }
  dt { padding-top: 3px; }
  dd { margin: 0; color: var(--text-dim); white-space: pre-wrap; font-size: 13px; }
  .prompt { font-family: var(--mono); font-size: 12px; margin-right: 10px; color: var(--text); }
  .prompt.loading { color: var(--text-muted); animation: blink 1.1s infinite; }
  code { font-family: var(--mono); }
</style>
