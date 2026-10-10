<script lang="ts">
  import { onMount } from "svelte";
  import { get, post } from "../lib/api";

  let { ready, back }: { ready: () => void; back?: () => void } = $props();
  let installing = $state(false);
  let error = $state("");

  async function install() {
    error = "";
    try {
      const status = await get<{ installation: { version: string } }>("/api/cliproxy");
      if (!status.installation.version) {
        installing = true;
        await post("/api/cliproxy/install", {});
      }
      ready();
    } catch (err) {
      installing = false;
      error = err instanceof Error ? err.message : String(err);
    }
  }
  onMount(() => { void install(); });
</script>

{#if error}
  <p class="error">{error}</p>
  <div class="actions">
    <button class="btn primary small" onclick={install}>Retry</button>
    {#if back}<button class="btn ghost small" onclick={back}>← Back</button>{/if}
  </div>
{:else}
  <p class="soft">{installing ? "Installing CLIProxyAPI…" : "Checking CLIProxyAPI…"}</p>
{/if}

<style>
  .error { color: var(--error); }
  .actions { display: flex; gap: 8px; }
</style>
