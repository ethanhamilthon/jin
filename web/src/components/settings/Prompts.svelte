<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import type { PromptInfo } from "../../lib/types";
  import Switch from "../Switch.svelte";
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
  <div class="set-edit"><TextEditor title={"#" + editing.name} text={editing.text} path={editing.path} {save} back={() => (editing = null)} /></div>
{:else}
  <div class="set-hd"><span class="label">Prompts</span><span class="hint">type #name in a message; {"{{commands}}"} run when a session starts</span></div>
  <form class="set-form" onsubmit={(e) => { e.preventDefault(); if (name.trim()) { editing = { name: name.trim(), text: "", path: "" }; name = ""; } }}>
    <input class="field mono" placeholder="new-prompt (folders with /)" bind:value={name} />
    <button class="btn">Add</button>
  </form>
  {#each list as p (p.name)}
    <div class="set-row">
      <div class="set-text">
        <span class="set-name mono">#{p.name}{#if p.system}<span class="pill dim">system</span>{/if}</span>
        <span class="set-desc" title={p.preview}>{p.preview}</span>
      </div>
      <div class="set-ctl">
        {#if !p.system}
          {#if confirm === p.name}
            <button class="btn danger small" onclick={() => remove(p.name)}>Delete</button>
            <button class="btn ghost small" onclick={() => (confirm = "")}>Keep</button>
          {:else}
            <button class="btn small" onclick={() => edit(p.name)}>Edit</button>
            <button class="btn ghost small" onclick={() => (confirm = p.name)}>Delete</button>
          {/if}
        {/if}
        <Switch label={"#" + p.name} checked={p.enabled} onchange={() => toggle(p.name)} />
      </div>
    </div>
  {/each}
{/if}
