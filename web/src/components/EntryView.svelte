<script lang="ts">
  import type { Entry } from "../lib/types";
  import Markdown from "./Markdown.svelte";
  import Reasoning from "./Reasoning.svelte";
  import ToolOutput from "./ToolOutput.svelte";

  let { entry, session, index }: { entry: Entry; session: string; index: number } = $props();
  const src = $derived(`/api/sessions/${session}/entries/${index}/image`);
  const label = $derived((entry.tool ?? "").replace("_", " "));
</script>

{#if entry.kind === "user"}
  <div class="user"><span>{entry.text}</span>
    {#if entry.pictures}
      <div class="shots">
        {#each Array(entry.pictures) as _, n}
          <a href="{src}?n={n}" target="_blank" rel="noreferrer noopener"><img src="{src}?n={n}" alt="image {n + 1}" loading="lazy" /></a>
        {/each}
      </div>
    {/if}
  </div>
{:else if entry.kind === "assistant"}
  <Markdown text={entry.text} />
{:else if entry.kind === "reasoning"}
  <Reasoning text={entry.text} />
{:else if entry.kind === "tool_call"}
  <div class="call"><span class="label">[ {label} ]</span><span class="mono summary">{entry.text}</span></div>
{:else if entry.kind === "tool_result"}
  <ToolOutput lines={entry.lines ?? []} tool={entry.tool ?? ""} />
{:else if entry.kind === "compacted"}
  <div class="divider"><span class="label">[ {entry.text} ]</span></div>
{:else if entry.kind === "ask"}
  <div class="card"><pre>{entry.text}</pre></div>
{:else if entry.picture}
  <figure class="picture">
    <a href={src} target="_blank" rel="noreferrer noopener"><img {src} alt={entry.text} loading="lazy" /></a>
    <figcaption class="info">{entry.text}</figcaption>
  </figure>
{:else if entry.tool === "shell"}
  <pre class="shell" class:error={entry.kind === "error"}>{entry.text}</pre>
{:else if entry.tool === "task"}
  <div class="info"><span class="label">[ task ]</span> {entry.text}</div>
{:else}
  <div class={entry.kind === "error" ? "error" : "info"}>{entry.text}</div>
{/if}

<style>
  .user {
    background: var(--card); border: 1px solid var(--raised); border-left: 2px solid var(--accent);
    padding: 10px 14px; white-space: pre-wrap; overflow-wrap: anywhere; color: var(--text-strong);
  }
  .shots { display: flex; flex-wrap: wrap; gap: 8px; margin-top: 8px; }
  .shots img { display: block; max-width: 100%; max-height: 240px; border: 1px solid var(--line); border-radius: var(--radius); }
  .call { display: flex; gap: 10px; align-items: baseline; min-width: 0; }
  .call .label { white-space: nowrap; color: var(--text-soft); }
  .summary { color: var(--text-dim); overflow-wrap: anywhere; font-size: 12.5px; }
  .divider { display: flex; align-items: center; gap: 12px; }
  .divider::before, .divider::after { content: ""; flex: 1; border-top: 1px solid var(--raised); }
  .card { background: var(--card); border: 1px solid var(--raised); padding: 10px 14px; }
  pre { margin: 0; white-space: pre-wrap; font: inherit; }
  .shell {
    font-family: var(--mono); font-size: 12.5px; background: var(--void); border: 1px solid var(--raised);
    padding: 8px 12px; border-radius: var(--radius); max-height: 360px; overflow: auto; color: var(--text-dim);
  }
  .shell.error { border-color: color-mix(in srgb, var(--error) 50%, transparent); }
  .picture { margin: 0; display: grid; gap: 4px; justify-items: start; }
  .picture img { display: block; max-width: 100%; max-height: 420px; border: 1px solid var(--raised); border-radius: var(--radius); }
  .info { color: var(--text-muted); font-size: 13px; white-space: pre-wrap; }
  .error { color: var(--error); white-space: pre-wrap; }
</style>
