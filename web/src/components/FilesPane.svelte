<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { insertIntoChat } from "../lib/files";
  import { shortPath } from "../lib/format";
  import PanelFrame from "./PanelFrame.svelte";
  import FileTree from "./FileTree.svelte";
  import FileView from "./FileView.svelte";
  import Icon from "./Icon.svelte";

  let { index }: { index: number } = $props();
  const pane = $derived(app.panes[index]);
  const dir = $derived(app.projectPath);
  let filter = $state("");
  let hidden = $state(false);
  const file = $derived(pane.file ?? "");
  const folder = $derived(file.includes("/") ? file.slice(0, file.lastIndexOf("/")) : "");

  // A file of another project does not exist in this one.
  let project = "";
  $effect(() => {
    if (project && project !== dir) pane.file = undefined;
    project = dir;
  });
</script>

<PanelFrame {index} title={file ? file.split("/").pop()! : "Files"} subtitle={file ? folder || "." : shortPath(dir, app.home)} back={file ? () => (pane.file = undefined) : undefined}>
  {#snippet actions()}
    {#if file}<button class="btn small primary" onclick={() => insertIntoChat(file)} title="Insert @path into the message">@ Insert</button>{/if}
  {/snippet}
  {#if file}
    <FileView {dir} path={file} insert={() => insertIntoChat(file)} />
  {:else}
    <div class="filter">
      <input class="field" placeholder="Filter files" bind:value={filter} />
      <button class="btn ghost eye" class:on={hidden} onclick={() => (hidden = !hidden)} title={hidden ? "Hide ignored files" : "Show hidden and ignored files"} aria-label="Show hidden files" aria-pressed={hidden}><Icon name="eye" size={20} /></button>
    </div>
    <FileTree {dir} {hidden} {filter} selected="" open={(path) => (pane.file = path)} />
  {/if}
</PanelFrame>

<style>
  .filter { display: flex; gap: 6px; padding: 10px 12px; border-bottom: 1px solid var(--raised); }
  .eye { padding: 6px 9px; }
  .eye.on { color: var(--accent); }
</style>
