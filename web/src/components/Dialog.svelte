<script lang="ts">
  import type { Snippet } from "svelte";
  import { app } from "../lib/app.svelte";

  let { title, label = "", wide = false, children }: { title: string; label?: string; wide?: boolean; children: Snippet } = $props();

  function close() {
    app.dialog = null;
  }
</script>

<svelte:window onkeydown={(e) => e.key === "Escape" && close()} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="backdrop" onclick={close}>
  <div class="dialog" class:wide role="dialog" aria-modal="true" aria-label={title} tabindex="-1" onclick={(e) => e.stopPropagation()}>
    <div class="head">
      <div>
        {#if label}<p class="label">[ {label} ]</p>{/if}
        <h2 class="serif">{title}</h2>
      </div>
      <button class="btn ghost small" onclick={close} aria-label="Close">✕</button>
    </div>
    <div class="body">{@render children()}</div>
  </div>
</div>

<style>
  .backdrop { position: fixed; inset: 0; background: rgb(0 0 0 / 0.6); display: grid; place-items: start center; padding-top: 10vh; z-index: 20; }
  .dialog {
    width: min(560px, calc(100vw - 32px)); max-height: 78vh; display: flex; flex-direction: column;
    background: var(--card); border: 1px solid var(--raised); outline: none;
  }
  .dialog.wide { width: min(880px, calc(100vw - 32px)); }
  .head { display: flex; justify-content: space-between; align-items: flex-start; padding: 18px 20px 10px; }
  .head p { margin: 0 0 4px; }
  h2 { font-size: 26px; margin: 0; line-height: 1.15; }
  .body { padding: 6px 20px 20px; overflow-y: auto; overflow-x: hidden; }
</style>
