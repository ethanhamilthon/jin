<script lang="ts">
  import type { Snippet } from "svelte";
  import { app } from "../lib/app.svelte";
  import { closePane } from "../lib/actions";
  import Icon from "./Icon.svelte";

  let { index, title, subtitle = "", back, actions, children }: {
    index: number; title: string; subtitle?: string; back?: () => void; actions?: Snippet; children: Snippet;
  } = $props();
</script>

<!-- svelte-ignore a11y_no_noninteractive_tabindex -->
<section tabindex="-1" onfocusin={() => (app.focused = index)}>
  <div class="bar">
    {#if back}<button class="btn ghost small" onclick={back} title="Back" aria-label="Back"><Icon name="back" /></button>{/if}
    <div class="title">
      <span class="serif name" {title}>{title}</span>
      {#if subtitle}<span class="sub mono" title={subtitle}><bdi>{subtitle}</bdi></span>{/if}
    </div>
    {@render actions?.()}
    <button class="btn ghost small" onclick={() => closePane(index)} title="Close pane (Alt+W)" aria-label="Close pane"><Icon name="close" /></button>
  </div>
  <div class="body">{@render children()}</div>
</section>

<style>
  section { display: flex; flex-direction: column; min-height: 0; min-width: 0; background: var(--canvas); outline: none; }
  .bar { display: flex; align-items: center; gap: 8px; padding: 8px 12px; border-bottom: 1px solid var(--raised); min-width: 0; }
  .title { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 1px; }
  .name { font-size: 17px; line-height: 1.3; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .sub { font-size: 11px; line-height: 1.4; color: var(--text-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; direction: rtl; text-align: left; }
  .body { flex: 1; min-height: 0; overflow-y: auto; overflow-x: hidden; }
</style>
