<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";
  import { newSession } from "../../lib/actions";
  import Dialog from "../Dialog.svelte";

  let { dir }: { dir: string } = $props();

  async function answer(trusted: boolean) {
    await post("/api/hooks/trust", { dir, trusted }).catch(fail);
    app.dialog = null;
    const view = app.current;
    if (trusted && view && view.state.path === dir && !view.state.persisted && !view.state.busy) await newSession(dir);
  }
</script>

<Dialog title="Run this project's hooks?" label=".jin/hooks">
  <p>The folder <span class="mono">{dir}</span> has its own hooks. Their <span class="mono">{"{{commands}}"}</span> run on your machine.</p>
  <p class="soft">New sessions of this project pick up the answer. Change it later in Settings → Hooks.</p>
  <div class="actions">
    <button class="btn ghost" onclick={() => answer(false)}>No, keep them off</button>
    <button class="btn primary" onclick={() => answer(true)}>Yes, I trust this repository</button>
  </div>
</Dialog>

<style>
  .actions { display: flex; justify-content: flex-end; gap: 8px; }
</style>
