<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import TextEditor from "./TextEditor.svelte";
  import { closeSection } from "../../lib/panels";

  const sections = [
    { id: "system", label: "System" },
    { id: "compact", label: "Compact" },
    { id: "handoff", label: "Handoff" },
    { id: "all", label: "All three" },
  ];

  let file = $state<{ path: string; content: string } | null>(null);
  let armed = $state("");
  let busy = $state("");

  $effect(() => {
    get<{ path: string; content: string }>("/api/sysprompt").then((f) => (file = f)).catch(fail);
  });

  async function save(text: string) {
    await post("/api/sysprompt", { content: text }).then(() => app.toast("Saved · new sessions use it")).catch(fail);
  }

  async function reset(section: string) {
    if (armed !== section) {
      armed = section;
      return;
    }
    armed = "";
    busy = section;
    await post<{ path: string; content: string }>("/api/sysprompt/reset", { section })
      .then((f) => ((file = f), app.toast("Reset to the latest from git · new sessions use it")))
      .catch(fail)
      .finally(() => (busy = ""));
  }
</script>

<p class="soft">The system, compaction and handoff prompts in one file. Changes reach new sessions; Reload applies them to an open one.</p>
{#if file}
  <div class="reset">
    <span class="label">[ reset to the latest from git ]</span>
    {#each sections as section}
      <button class="btn ghost" class:armed={armed === section.id} disabled={!!busy} onclick={() => reset(section.id)} onblur={() => (armed = "")}>
        {busy === section.id ? "Fetching…" : armed === section.id ? `Replace ${section.label.toLowerCase()}? Click again` : section.label}
      </button>
    {/each}
  </div>
  <TextEditor title="system-prompt.md" text={file.content} path={file.path} {save} back={closeSection} />
{:else}<p class="empty">Loading…</p>{/if}

<style>
  .reset { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin: 8px 0; }
  .reset .label { margin-right: 4px; }
  .armed { color: var(--accent); border-color: var(--accent); }
</style>
