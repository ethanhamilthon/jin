<script lang="ts">
  import Icon from "./Icon.svelte";

  let { pick, disabled = false }: { pick: (files: File[]) => void; disabled?: boolean } = $props();
  let input: HTMLInputElement;

  function chosen() {
    pick([...(input.files ?? [])]);
    input.value = "";
  }
</script>

<input bind:this={input} type="file" multiple hidden onchange={chosen} />
<button class="btn ghost small" {disabled} onclick={() => input.click()} title="Attach files" aria-label="Attach files"><Icon name="clip" size={15} /></button>
