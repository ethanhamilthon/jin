<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { parseSections, renderSections, sectionNames, type SectionName, type Sections } from "../../lib/sysprompt";

  const labels: Record<SectionName, string> = { system: "System", compact: "Compact", handoff: "Handoff" };

  let path = $state("");
  let sections = $state<Sections | null>(null);
  let tab = $state<SectionName>("system");
  let armed = $state(false);
  let busy = $state(false);
  let saving = $state(false);

  function load(file: { path: string; content: string }) {
    path = file.path;
    sections = parseSections(file.content);
  }

  $effect(() => {
    get<{ path: string; content: string }>("/api/sysprompt").then(load).catch(fail);
  });

  async function save() {
    if (!sections) return;
    saving = true;
    await post("/api/sysprompt", { content: renderSections(sections) })
      .then(() => app.toast("Saved · new sessions use it"))
      .catch(fail)
      .finally(() => (saving = false));
  }

  async function reset() {
    if (!armed) {
      armed = true;
      return;
    }
    armed = false;
    busy = true;
    await post<{ path: string; content: string }>("/api/sysprompt/reset", { section: tab })
      .then((file) => (load(file), app.toast(`${labels[tab]} reset to the latest from git · new sessions use it`)))
      .catch(fail)
      .finally(() => (busy = false));
  }
</script>

<div class="set-hd"><span class="label">System prompt</span><span class="hint">one file: system, compact, handoff</span></div>
{#if sections}
  <div class="set-tabs">
    {#each sectionNames as name (name)}
      <button class:on={tab === name} onclick={() => ((tab = name), (armed = false))}>{labels[name]}</button>
    {/each}
  </div>
  <div class="set-pad">
    <textarea class="field mono" bind:value={sections[tab]} spellcheck="false" onkeydown={(e) => (e.metaKey || e.ctrlKey) && e.key === "s" && (e.preventDefault(), save())}></textarea>
  </div>
  <div class="actions">
    <button class="btn small" class:armed disabled={busy} onclick={reset} onblur={() => (armed = false)}>
      {busy ? "Fetching…" : armed ? `Replace ${labels[tab].toLowerCase()}? Click again` : "Reset to latest from git"}
    </button>
    <span class="mono path" title={path}>{path}</span>
    <button class="btn primary small" onclick={save} disabled={saving}>Save</button>
  </div>
{:else}<p class="empty">Loading…</p>{/if}

<style>
  textarea { min-height: 42vh; resize: vertical; font-size: 12.5px; line-height: 1.55; display: block; }
  .actions { display: flex; align-items: center; gap: 10px; padding: 0 16px 16px; }
  .path { flex: 1; min-width: 0; font-size: 11px; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl; text-align: left; }
  .armed { color: var(--accent); border-color: var(--accent); }
</style>
