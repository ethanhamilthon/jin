<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { Hook } from "../../lib/types";
  import Switch from "../Switch.svelte";
  import TextEditor from "./TextEditor.svelte";

  const dir = $derived(app.projectPath);
  let hooks = $state<Hook[]>([]);
  let editing = $state<{ hook: Hook; text: string; path: string } | null>(null);
  let name = $state("");
  let confirm = $state("");

  $effect(() => {
    void app.configRev;
    get<{ hooks: Hook[]; trust: number }>("/api/hooks" + query({ dir }))
      .then((r) => (hooks = r.hooks.filter((h) => !h.project)))
      .catch(fail);
  });

  async function edit(hook: Hook) {
    const t = await get<{ path: string; content: string }>("/api/hooks/text" + query({ name: hook.name, dir })).catch(() => ({ path: "", content: "" }));
    editing = { hook, text: t.content, path: t.path };
  }

  const ref = (h: Hook) => ({ name: h.name, project: false, dir });
  const save = (text: string) => post("/api/hooks", { ...ref(editing!.hook), content: text }).then(() => (editing = null)).catch(fail);
  const toggle = (h: Hook) => post("/api/hooks/toggle", ref(h)).catch(fail);
  const remove = (h: Hook) => post("/api/hooks/delete", ref(h)).catch(fail).finally(() => (confirm = ""));
</script>

{#if editing}
  <div class="set-pad"><TextEditor title={editing.hook.name} text={editing.text} path={editing.path} {save} back={() => (editing = null)} /></div>
{:else}
  <div class="set-hd"><span class="label">Hooks</span><span class="hint">added to the system prompt of new sessions</span></div>
  <form class="set-form" onsubmit={(e) => { e.preventDefault(); if (name.trim()) { editing = { hook: { name: name.trim(), project: false, dir, enabled: true, preview: "" }, text: "", path: "" }; name = ""; } }}>
    <input class="field mono" placeholder="new-hook" bind:value={name} />
    <button class="btn">Add</button>
  </form>
  {#each hooks as h (h.name)}
    <div class="set-row">
      <div class="set-text">
        <span class="set-name mono">{h.name}</span>
        <span class="set-desc" title={h.preview}>{h.preview}</span>
      </div>
      <div class="set-ctl">
        {#if confirm === h.name}
          <button class="btn danger small" onclick={() => remove(h)}>Delete</button>
          <button class="btn ghost small" onclick={() => (confirm = "")}>Keep</button>
        {:else}
          <button class="btn small" onclick={() => edit(h)}>Edit</button>
          <button class="btn ghost small" onclick={() => (confirm = h.name)}>Delete</button>
        {/if}
        <Switch label={h.name} checked={h.enabled} onchange={() => toggle(h)} />
      </div>
    </div>
  {:else}<p class="empty">No hooks yet</p>{/each}
  <p class="set-note">Hooks of one project and their trust live in that project's Project pane.</p>
{/if}
