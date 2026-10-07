<script lang="ts">
  import { render } from "../lib/markdown";

  let { text }: { text: string } = $props();
  const html = $derived(render(text));
</script>

<div class="md">{@html html}</div>
