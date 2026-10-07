<script lang="ts">
  import { render } from "../lib/markdown";

  let { text }: { text: string } = $props();
  const html = $derived(render(text));

  async function copy(event: MouseEvent) {
    const button = (event.target as HTMLElement).closest<HTMLButtonElement>(".code-copy");
    if (!button) return;
    await navigator.clipboard.writeText(button.parentElement?.querySelector("code")?.textContent ?? "");
    button.textContent = "copied";
    setTimeout(() => (button.textContent = "copy"), 1500);
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div class="md" onclick={copy}>{@html html}</div>
