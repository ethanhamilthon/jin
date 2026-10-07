<script lang="ts">
  import { fail } from "../../lib/app.svelte";
  import { get } from "../../lib/api";
  import { percent, tokens } from "../../lib/format";
  import type { ContextReport } from "../../lib/types";
  import Dialog from "../Dialog.svelte";

  let { id }: { id: string } = $props();
  let report = $state<ContextReport | null>(null);

  $effect(() => {
    get<ContextReport>(`/api/sessions/${id}/context`).then((r) => (report = r)).catch(fail);
  });

  const prompt = $derived((report?.prompt ?? []).reduce((n, p) => n + p.tokens, 0));
</script>

<Dialog title="Context" label="token counts are approximate">
  {#if report}
    <div class="used">
      <span class="serif big">{tokens(report.used)}</span>
      {#if report.window}<span class="soft">of {tokens(report.window)} · {percent(report.used, report.window)}%</span>{/if}
    </div>
    {#if report.window}<div class="bar"><div style="width: {Math.min(100, percent(report.used, report.window))}%"></div></div>{/if}
    <dl>
      <dt>System prompt</dt><dd>~{tokens(prompt)}</dd>
      {#each report.prompt ?? [] as part (part.name)}<dt class="sub">{part.name}</dt><dd>~{tokens(part.tokens)}</dd>{/each}
      <dt>Tool schemas</dt><dd>~{tokens(report.tool_schemas)}</dd>
      <dt>Conversation</dt><dd>{report.messages} messages · ~{tokens(report.conversation)}</dd>
      {#if report.results?.length}<dt>Largest tool results</dt><dd></dd>{/if}
      {#each report.results ?? [] as part, i (i)}<dt class="sub mono">{part.name}</dt><dd>~{tokens(part.tokens)}</dd>{/each}
      {#if report.cache !== undefined && report.cache !== null}<dt>Cache</dt><dd>{report.cache}% of the last request</dd>{/if}
    </dl>
  {:else}<p class="empty">Measuring…</p>{/if}
</Dialog>

<style>
  .used { display: flex; align-items: baseline; gap: 10px; }
  .big { font-size: 40px; }
  .bar { height: 4px; background: var(--raised); margin: 8px 0 16px; }
  .bar div { height: 100%; background: var(--accent); box-shadow: var(--glow); }
  dl { display: grid; grid-template-columns: 1fr auto; gap: 6px 16px; margin: 0; }
  dt { color: var(--text); }
  dt.sub { padding-left: 14px; color: var(--text-soft); font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  dd { margin: 0; text-align: right; font-family: var(--mono); font-size: 12px; color: var(--text-dim); }
</style>
