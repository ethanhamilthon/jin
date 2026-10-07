<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { Hook } from "../../lib/types";
  import TextEditor from "./TextEditor.svelte";

  const dir = $derived(app.current?.state.path ?? app.dir);
  let hooks = $state<Hook[]>([]);
  let trust = $state(0);
  let editing = $state<{ hook: Hook; text: string; path: string } | null>(null);
  let name = $state("");
  let project = $state(false);
  let confirm = $state("");

  $effect(() => {
    void app.configRev;
    get<{ hooks: Hook[]; trust: number }>("/api/hooks" + query({ dir }))
      .then((r) => ([hooks, trust] = [r.hooks, r.trust]))
      .catch(fail);
  });

  async function edit(hook: Hook) {
    const t = await get<{ path: string; content: string }>("/api/hooks/text" + query({ name: hook.name, project: hook.project ? "1" : "", dir })).catch(() => ({ path: "", content: "" }));
    editing = { hook, text: t.content, path: t.path };
  }

  const ref = (h: Hook) => ({ name: h.name, project: h.project, dir });
  const save = (text: string) => post("/api/hooks", { ...ref(editing!.hook), content: text }).then(() => (editing = null)).catch(fail);
  const toggle = (h: Hook) => post("/api/hooks/toggle", ref(h)).catch(fail);
  const remove = (h: Hook) => post("/api/hooks/delete", ref(h)).catch(fail).finally(() => (confirm = ""));
  const setTrust = (trusted: boolean) => post("/api/hooks/trust", { dir, trusted }).catch(fail);
  const key = (h: Hook) => (h.project ? "project:" : "") + h.name;
</script>

{#if editing}
  <TextEditor title={editing.hook.name} text={editing.text} path={editing.path} {save} back={() => (editing = null)} />
{:else}
  <p class="soft">Hooks add instructions to the system prompt of new sessions. Project hooks live in <span class="mono">.jin/hooks</span> of {dir}.</p>
  <div class="trust">
    <span>Project hooks: <strong>{trust === 1 ? "trusted" : trust === 2 ? "not trusted" : "not decided"}</strong></span>
    {#if trust !== 1}<button class="btn small" onclick={() => setTrust(true)}>Trust</button>{/if}
    {#if trust !== 2}<button class="btn ghost small" onclick={() => setTrust(false)}>Distrust</button>{/if}
  </div>
  <form class="add" onsubmit={(e) => { e.preventDefault(); if (name.trim()) { editing = { hook: { name: name.trim(), project, dir, enabled: true, preview: "" }, text: "", path: "" }; name = ""; } }}>
    <input class="field mono" placeholder="new-hook" bind:value={name} />
    <label class="soft"><input type="checkbox" bind:checked={project} /> project</label>
    <button class="btn">Add</button>
  </form>
  {#each hooks as h (key(h))}
    <div class="row">
      <input type="checkbox" checked={h.enabled} onchange={() => toggle(h)} disabled={h.project && trust !== 1} />
      <span class="text"><span class="mono name">{h.name}</span>{#if h.project}<span class="pill dim">project</span>{/if}<span class="soft preview">{h.preview}</span></span>
      {#if confirm === key(h)}
        <button class="btn danger small" onclick={() => remove(h)}>Delete</button>
        <button class="btn ghost small" onclick={() => (confirm = "")}>Keep</button>
      {:else}
        <button class="btn small" onclick={() => edit(h)}>Edit</button>
        <button class="btn ghost small" onclick={() => (confirm = key(h))}>Delete</button>
      {/if}
    </div>
  {:else}<p class="empty">No hooks yet</p>{/each}
{/if}

<style>
  .trust { display: flex; gap: 10px; align-items: center; margin: 8px 0; }
  .add { display: flex; gap: 8px; margin: 8px 0 12px; align-items: center; }
  .add label { display: flex; gap: 6px; white-space: nowrap; }
  .row { display: flex; gap: 10px; align-items: center; padding: 6px 0; border-bottom: 1px solid var(--raised); }
  .text { flex: 1; min-width: 0; display: flex; gap: 10px; align-items: center; }
  .name { color: var(--text-strong); white-space: nowrap; }
  .preview { font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
