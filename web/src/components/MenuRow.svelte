<script lang="ts">
  import Icon from "./Icon.svelte";

  let { icon, label, kind, on = false, hint, pick }: {
    icon: string; label: string; kind: "open" | "run" | "toggle"; on?: boolean; hint?: string; pick: () => void;
  } = $props();
</script>

<button role={kind === "toggle" ? "menuitemcheckbox" : "menuitem"} aria-checked={kind === "toggle" ? on : undefined} onclick={pick}>
  <span class="icon"><Icon name={icon} size={15} /></span>
  <span class="text">{label}</span>
  {#if hint}<span class="hint mono">{hint}</span>{/if}
  {#if kind === "toggle"}
    <span class="switch" class:on aria-hidden="true"><span class="thumb"></span></span>
  {:else}
    <span class="kind"><Icon name={kind === "open" ? "chevron" : "play"} size={kind === "open" ? 14 : 12} /></span>
  {/if}
</button>

<style>
  button { display: flex; align-items: center; gap: 10px; width: 100%; padding: 7px 8px; border: 0; background: transparent; border-radius: 2px; color: var(--text); font-size: 13px; text-align: left; cursor: pointer; }
  button:hover { background: var(--hover); }
  .icon { display: inline-flex; width: 16px; color: var(--text-soft); }
  button:hover .icon { color: var(--text-strong); }
  .text { flex: 1; min-width: 0; white-space: nowrap; }
  .hint { color: var(--text-muted); font-size: 11px; }
  .kind { display: inline-flex; width: 28px; justify-content: flex-end; color: var(--text-muted); }
  .switch { position: relative; width: 28px; height: 14px; border: 1px solid var(--line); background: var(--canvas); }
  .thumb { position: absolute; top: 1px; left: 1px; width: 10px; height: 10px; background: var(--text-muted); transition: transform 120ms; }
  .switch.on { border-color: var(--accent); }
  .switch.on .thumb { transform: translateX(14px); background: var(--accent); }
  @media (prefers-reduced-motion: reduce) { .thumb { transition: none; } }
</style>
