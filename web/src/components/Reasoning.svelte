<script lang="ts">
  import Markdown from "./Markdown.svelte";

  let { text }: { text: string } = $props();
</script>

<div class="reasoning"><Markdown {text} /></div>

<style>
  .reasoning { --text: var(--text-soft); --text-strong: var(--text-dim); }
</style>
