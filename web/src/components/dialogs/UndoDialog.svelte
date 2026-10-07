<script lang="ts">
  import { app } from "../../lib/app.svelte";
  import { get } from "../../lib/api";
  import { act } from "../../lib/actions";
  import Dialog from "../Dialog.svelte";

  let { id }: { id: string } = $props();
  let preview = $state<{ skipped: string[] | null; scope: string } | null>(null);
  let error = $state("");

  $effect(() => {
    get<{ skipped: string[] | null; scope: string }>(`/api/sessions/${id}/undo`)
      .then((p) => (preview = p))
      .catch((e) => (error = e.message));
  });

  async function undo() {
    if (await act(id, "undo")) app.dialog = null;
  }
</script>

<Dialog title="Undo the last turn" label="edit and write changes">
  {#if error}
    <p class="soft">{error}</p>
  {:else if preview}
    <p class="soft">{preview.scope}</p>
    {#if preview.skipped?.length}
      <p>These files changed after the agent wrote them and stay as they are:</p>
      <ul class="mono">{#each preview.skipped as path (path)}<li>{path}</li>{/each}</ul>
    {/if}
    <div class="actions">
      <button class="btn ghost" onclick={() => (app.dialog = null)}>Cancel</button>
      <button class="btn primary" onclick={undo}>Undo</button>
    </div>
  {:else}<p class="empty">Checking…</p>{/if}
</Dialog>

<style>
  .actions { display: flex; justify-content: flex-end; gap: 8px; margin-top: 12px; }
  ul { font-size: 12px; color: var(--warn); }
</style>
