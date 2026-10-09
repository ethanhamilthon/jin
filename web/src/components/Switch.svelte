<script lang="ts">
  let { checked, onchange, disabled = false, label = "" }: {
    checked: boolean; onchange: (value: boolean) => void; disabled?: boolean; label?: string;
  } = $props();
</script>

<button type="button" role="switch" aria-checked={checked} aria-label={label} class="switch" class:on={checked} {disabled} onclick={() => onchange(!checked)}></button>

<style>
  .switch {
    flex: none; width: 30px; height: 16px; border-radius: 2px; padding: 0; position: relative; cursor: pointer;
    background: var(--hover); border: 1px solid var(--line);
  }
  .switch::after { content: ""; position: absolute; top: 2px; left: 2px; width: 10px; height: 10px; background: var(--text-muted); }
  .switch.on { background: var(--accent-soft); border-color: var(--accent); box-shadow: var(--glow); }
  .switch.on::after { left: 16px; background: var(--accent); }
  .switch:disabled { opacity: 0.4; cursor: default; }
</style>
