<script lang="ts">
  import { post } from "../lib/api";
  import { fail } from "../lib/app.svelte";

  let { done, back }: { done: () => void; back: () => void } = $props();
  const profiles = [
    { id: "claude", name: "Claude", detail: "Anthropic subscription" },
    { id: "codex", name: "Codex", detail: "OpenAI subscription" },
    { id: "antigravity", name: "Antigravity", detail: "Google subscription" },
  ];

  async function add(profile: string) {
    try { await post("/api/cliproxy/providers", { profile }); done(); }
    catch (error) { fail(error); }
  }
</script>

<div class="kinds">
  {#each profiles as p, i (p.id)}
    <button class="list-row kind" onclick={() => add(p.id)}>
      <span class="mono n">{i + 1}</span><span class="name">{p.name}</span><span class="soft">{p.detail}</span>
    </button>
  {/each}
</div>
<p class="soft">Install CLIProxyAPI in Settings, then sign in on this computer. Provider terms apply.</p>
<button class="btn ghost small" onclick={back}>← Back</button>

<style>
  .kinds { display: grid; gap: 4px; }
  .kind { padding: 12px; border: 1px solid var(--raised); }
  .n { color: var(--accent); }
  .name { flex: 1; color: var(--text-strong); }
</style>
