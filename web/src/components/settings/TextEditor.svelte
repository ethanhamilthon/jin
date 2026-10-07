<script lang="ts">
  let { title, text, path = "", save, back }: { title: string; text: string; path?: string; save: (text: string) => Promise<unknown>; back: () => void } = $props();
  let value = $state("");
  let saving = $state(false);
  $effect(() => {
    value = text;
  });

  async function submit() {
    saving = true;
    await save(value).finally(() => (saving = false));
  }
</script>

<div class="editor">
  <div class="head">
    <span class="label">[ {title} ]</span>
    {#if path}<span class="mono soft path">{path}</span>{/if}
  </div>
  <textarea class="field mono" bind:value spellcheck="false" onkeydown={(e) => (e.metaKey || e.ctrlKey) && e.key === "s" && (e.preventDefault(), submit())}></textarea>
  <div class="actions">
    <button class="btn ghost" onclick={back}>← Back</button>
    <button class="btn primary" onclick={submit} disabled={saving}>Save</button>
  </div>
</div>

<style>
  .editor { display: grid; gap: 8px; }
  .head { display: flex; gap: 10px; align-items: baseline; min-width: 0; }
  .path { font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  textarea { min-height: 44vh; resize: vertical; font-size: 12.5px; line-height: 1.55; }
  .actions { display: flex; justify-content: space-between; }
</style>
