<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { closeDesktopNotice, desktopNotice } from "../lib/desktop-notice.svelte";
  import Icon from "./Icon.svelte";

  const notice = $derived(desktopNotice.value);

  async function copy() {
    if (!notice?.line) return;
    try {
      await navigator.clipboard.writeText(notice.line);
      app.toast("Copied");
    } catch {
      app.toast("Copy failed: select the line and copy it");
    }
  }
</script>

{#if notice}
  <div class="notice" role="status">
    {#if notice.kind === "path"}
      <div class="text">
        <p>The <code>jin</code> command is linked in <code>{notice.dir}</code>, but that folder is not in your PATH.</p>
        <p>Add this line to <code>~/.zshrc</code>, then open a new terminal:</p>
        <code class="line">{notice.line}</code>
      </div>
      <button class="btn small" onclick={copy}>Copy</button>
    {:else}
      <div class="text"><p>{notice.message}</p></div>
    {/if}
    <button class="btn ghost small" onclick={closeDesktopNotice} title="Close" aria-label="Close"><Icon name="close" /></button>
  </div>
{/if}

<style>
  .notice {
    position: fixed; left: 16px; bottom: 16px; z-index: 50; max-width: 460px; display: flex; gap: 8px; align-items: flex-start;
    background: var(--hover); border: 1px solid var(--accent); border-radius: var(--radius); padding: 10px 12px; color: var(--text);
  }
  .text { flex: 1; min-width: 0; display: grid; gap: 6px; }
  p { margin: 0; }
  .line { display: block; font-family: var(--mono); font-size: 12.5px; background: var(--void); padding: 6px 8px; border-radius: var(--radius); overflow-wrap: anywhere; }
</style>
