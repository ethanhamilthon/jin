<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get } from "../../lib/api";
  import { cost, percent, tokens } from "../../lib/format";
  import type { ContextReport } from "../../lib/types";
  import Dialog from "../Dialog.svelte";
  import Icon from "../Icon.svelte";

  let { id }: { id: string } = $props();
  let report = $state<ContextReport | null>(null);

  $effect(() => {
    get<ContextReport>(`/api/sessions/${id}/context`).then((r) => (report = r)).catch(fail);
  });

  const usage = $derived(app.sessions[id]?.state.usage);
  const prompt = $derived((report?.prompt ?? []).reduce((n, p) => n + p.tokens, 0));
  const promptHtml = $derived((report?.prompt ?? []).map((p) => {
    const text = escapeHtml(p.text ?? "");
    return p.name.startsWith("command: ") ? `<span class="cmd">${text}</span>` : text;
  }).join(""));

  function escapeHtml(s: string): string {
    return s.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;");
  }
  const cache = $derived(report?.cache ?? null);
  const meta = $derived([
    ...(usage ? [`Spent ${cost(usage.Cost)}`, `In ${tokens(usage.Input)}`, `Out ${tokens(usage.Output)}`] : []),
    ...(cache !== null ? [`Cache ${cache}%`] : []),
  ]);
</script>

<Dialog title="Context" label="token counts are approximate" wide>
  {#if report}
    <div class="used">
      <div class="left">
        <span class="serif big">{tokens(report.used)}</span>
        {#if report.window}<span class="soft">of {tokens(report.window)} · {percent(report.used, report.window)}%</span>{/if}
      </div>
      {#if meta.length}<div class="meta">{meta.join(" · ")}</div>{/if}
    </div>
    {#if report.window}<div class="bar"><div style="width: {Math.min(100, percent(report.used, report.window))}%"></div></div>{/if}
    <details class="sys">
      <summary><span>System prompt</span><span class="right"><span class="soft">~{tokens(prompt)}</span><span class="chev"><Icon name="chevron" size={14} /></span></span></summary>
      <pre class="full">{@html promptHtml}</pre>
    </details>
    <dl>
      <dt>Tool schemas</dt><dd>~{tokens(report.tool_schemas)}</dd>
      <dt>Conversation</dt><dd>{report.messages} messages · ~{tokens(report.conversation)}</dd>
    </dl>
  {:else}<p class="empty">Measuring…</p>{/if}
</Dialog>

<style>
  .used { display: flex; align-items: baseline; justify-content: space-between; gap: 16px; }
  .left { display: flex; align-items: baseline; gap: 10px; min-width: 0; }
  .left .soft { white-space: nowrap; }
  .big { font-size: 40px; }
  .meta { font-family: var(--mono); font-size: 12px; color: var(--text-muted); text-align: right; }
  @media (max-width: 560px) {
    .used { flex-direction: column; align-items: stretch; gap: 4px; }
    .big { font-size: 34px; }
    .meta { text-align: left; }
  }
  .bar { height: 4px; background: var(--raised); margin: 8px 0 16px; }
  .bar div { height: 100%; background: var(--accent); box-shadow: var(--glow); }
  .sys { border-top: 1px solid var(--raised); }
  .sys summary { display: flex; justify-content: space-between; align-items: center; gap: 12px; padding: 8px 0; cursor: pointer; }
  .sys .right { display: inline-flex; align-items: center; gap: 8px; }
  .sys .soft { font-family: var(--mono); font-size: 12px; }
  .chev { display: inline-flex; transition: transform 0.15s; color: var(--text-muted); }
  .sys[open] .chev { transform: rotate(90deg); }
  .full :global(.cmd) { color: var(--accent); }
  dl { display: grid; grid-template-columns: 1fr auto; gap: 6px 16px; margin: 8px 0 0; padding-top: 8px; border-top: 1px solid var(--raised); }
  pre { margin: 0 0 10px; padding: 10px 12px; background: var(--void); border: 1px solid var(--raised); font: 12px/1.55 var(--mono); white-space: pre-wrap; overflow-wrap: anywhere; max-height: 40vh; overflow: auto; }
  dd { margin: 0; text-align: right; font-family: var(--mono); font-size: 12px; color: var(--text-dim); }
</style>
