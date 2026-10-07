<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import TextEditor from "./TextEditor.svelte";

  let file = $state<{ path: string; content: string } | null>(null);

  $effect(() => {
    get<{ path: string; content: string }>("/api/sysprompt").then((f) => (file = f)).catch(fail);
  });

  async function save(text: string) {
    await post("/api/sysprompt", { content: text }).then(() => app.toast("Saved · new sessions use it")).catch(fail);
  }
</script>

<p class="soft">The system, compaction and handoff prompts in one file. Changes reach new sessions; Reload applies them to an open one.</p>
{#if file}
  <TextEditor title="system-prompt.md" text={file.content} path={file.path} {save} back={() => (app.dialog = null)} />
{:else}<p class="empty">Loading…</p>{/if}
