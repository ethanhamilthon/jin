<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { PromptInfo } from "../../lib/types";
  import TextEditor from "./TextEditor.svelte";

  let list = $state<PromptInfo[]>([]);
  let editing = $state<{ name: string; text: string; path: string } | null>(null);
  let name = $state("");
  let confirm = $state("");

  $effect(() => {
    void app.configRev;
    get<PromptInfo[]>("/api/prompts").then((l) => (list = l)).catch(fail);
  });

  async function edit(prompt: string) {
    const t = await get<{ path: string; content: string }>("/api/prompts/text" + query({ name: prompt })).catch(() => ({ path: "", content: "" }));
    editing = { name: prompt, text: t.content, path: t.path };
  }

  async function save(text: string) {
    await post("/api/prompts", { name: editing!.name, content: text }).then(() => (editing = null)).catch(fail);
  }

  const toggle = (prompt: string) => post("/api/prompts/toggle", { name: prompt }).catch(fail);
  const remove = (prompt: string) => post("/api/prompts/delete", { name: prompt }).catch(fail).finally(() => (confirm = ""));
</script>

{#if editing}
  <TextEditor title={"#" + editing.name} text={editing.text} path={editing.path} {save} back={() => (editing = null)} />
{:else}
  <p class="soft">Reusable prompts: type <span class="mono">#name</span> in a message. They live in ~/.jin/prompts; {"{{commands}}"} in them run when a session starts.</p>
  <form class="add" onsubmit={(e) => { e.preventDefault(); if (name.trim()) { editing = { name: name.trim(), text: "", path: "" }; name = ""; } }}>
    <input class="field mono" placeholder="new-prompt (folders with /)" bind:value={name} />
    <button class="btn">Add</button>
  </form>
  {#each list as p (p.name)}
    <div class="row">
      <input type="checkbox" checked={p.enabled} onchange={() => toggle(p.name)} title="On or off" />
      <span class="text"><span class="mono name">#{p.name}</span>{#if p.system}<span class="pill dim">system</span>{/if}<span class="soft preview">{p.preview}</span></span>
      {#if !p.system}
        {#if confirm === p.name}
          <button class="btn danger small" onclick={() => remove(p.name)}>Delete</button>
          <button class="btn ghost small" onclick={() => (confirm = "")}>Keep</button>
        {:else}
          <button class="btn small" onclick={() => edit(p.name)}>Edit</button>
          <button class="btn ghost small" onclick={() => (confirm = p.name)}>Delete</button>
        {/if}
      {/if}
    </div>
  {/each}
{/if}

<style>
  .add { display: flex; gap: 8px; margin: 8px 0 12px; }
  .row { display: flex; gap: 10px; align-items: center; padding: 6px 0; border-bottom: 1px solid var(--raised); }
  .text { flex: 1; min-width: 0; display: flex; gap: 10px; align-items: center; }
  .name { color: var(--text-strong); white-space: nowrap; }
  .preview { font-size: 12px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
