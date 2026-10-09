<script lang="ts">
  import hljs from "highlight.js/lib/common";
  import { fail } from "../lib/app.svelte";
  import { get, query } from "../lib/api";
  import { escape } from "../lib/escape";
  import { languageOf, size, type FileView } from "../lib/files";
  import { every } from "../lib/poll";
  import Markdown from "./Markdown.svelte";

  let { dir, path, insert }: { dir: string; path: string; insert: () => void } = $props();
  let view = $state<FileView | null>(null);
  let source = $state(false);

  let gone = $state(false);

  // fetch reads the file. A refresh changes the view only when the file changed.
  async function fetchView(target: string, refresh = false) {
    try {
      const next = await get<FileView>("/api/tree/file" + query({ dir, path: target }));
      if (target !== path) return;
      gone = false;
      if (!refresh || next.size !== view?.size || next.content !== view?.content || next.kind !== view?.kind) view = next;
    } catch (err) {
      if (!refresh) fail(err);
      else if (target === path) gone = true;
    }
  }

  $effect(() => {
    const target = path;
    view = null;
    gone = false;
    source = false;
    fetchView(target);
    return every(() => view?.kind !== "image" && fetchView(target, true));
  });

  const html = $derived.by(() => {
    if (!view?.content) return "";
    const language = languageOf(view.name);
    return language && hljs.getLanguage(language) ? hljs.highlight(view.content, { language }).value : escape(view.content);
  });
  const lines = $derived(view?.content ? view.content.replace(/\n$/, "").split("\n").length : 0);
  const raw = $derived("/api/tree/raw" + query({ dir, path }));
</script>

{#if !view}<p class="empty">Loading…</p>
{:else}
  {#if gone}<div class="gone">This file no longer exists.</div>{/if}
  <div class="meta"><span class="mono">{view.path} · {size(view.size)}{#if view.content !== undefined && view.kind !== "image"} · {lines} lines{/if}</span></div>
  {#if view.kind === "markdown"}
    <div class="seg"><button class:on={!source} onclick={() => (source = false)}>Preview</button><button class:on={source} onclick={() => (source = true)}>Source</button></div>
  {/if}
  {#if view.kind === "markdown" && !source}
    <div class="doc"><Markdown text={view.content ?? ""} /></div>
  {:else if view.kind === "text" || view.kind === "markdown"}
    <div class="code">
      <div class="ln" aria-hidden="true">{#each { length: lines } as _, i}{i + 1}<br />{/each}</div>
      <pre class="hljs">{@html html}</pre>
    </div>
  {:else if view.kind === "image"}
    <div class="img"><img src={raw} alt={view.name} /></div>
  {:else}
    <div class="nope">
      <p class="label">No preview</p>
      <p class="soft">{view.kind === "large" ? "This file is larger than 1 MB." : "This is not a text file."} The agent can still read it with the read tool.</p>
      <button class="btn small primary" onclick={insert}>@ Insert path</button>
    </div>
  {/if}
{/if}

<style>
  .gone { padding: 8px 12px; border-bottom: 1px solid var(--warn); color: var(--warn); font-size: 12px; }
  .meta { padding: 8px 12px; border-bottom: 1px solid var(--raised); color: var(--text-soft); font-size: 12px; overflow-wrap: anywhere; }
  .seg { display: flex; padding: 8px 12px; }
  .seg button { background: none; border: 1px solid var(--raised); padding: 4px 12px; color: var(--text-soft); cursor: pointer; }
  .seg .on { background: var(--raised); color: var(--text-strong); }
  .doc { padding: 4px 16px 20px; }
  .code { display: flex; font: 12.5px/1.6 var(--mono); }
  .ln { padding: 10px 10px 10px 12px; text-align: right; color: var(--text-muted); border-right: 1px solid var(--raised); user-select: none; }
  pre { margin: 0; padding: 10px 14px; overflow-x: auto; flex: 1; min-width: 0; background: none; color: var(--text-dim); }
  .img { display: grid; place-items: center; padding: 16px; }
  .img img { max-width: 100%; max-height: 70vh; border: 1px solid var(--line); background: repeating-conic-gradient(var(--card) 0 25%, var(--canvas) 0 50%) 0 0 / 16px 16px; }
  .nope { padding: 24px 16px; display: grid; gap: 10px; justify-items: start; }
</style>
