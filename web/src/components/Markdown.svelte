<script lang="ts">
  import { render } from "../lib/markdown";
  import { app } from "../lib/app.svelte";

  let { text }: { text: string } = $props();
  const html = $derived(render(text, app.current?.state.path ?? app.dir));

  function picture(event: MouseEvent): boolean {
    const target = event.target as HTMLElement;
    if (target.tagName !== "IMG" || target.closest("a")) return false;
    window.open((target as HTMLImageElement).src, "_blank", "noopener");
    return true;
  }

  function anchor(event: MouseEvent): boolean {
    const link = (event.target as HTMLElement).closest<HTMLAnchorElement>('a[href^="#"]');
    if (!link) return false;
    event.preventDefault();
    const id = decodeURIComponent(link.getAttribute("href")!.slice(1));
    [...(event.currentTarget as HTMLElement).querySelectorAll("[id]")].find((e) => e.id === id)?.scrollIntoView({ block: "center" });
    return true;
  }

  async function copy(event: MouseEvent) {
    if (picture(event) || anchor(event)) return;
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
