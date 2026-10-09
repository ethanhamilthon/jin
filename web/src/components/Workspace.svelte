<script lang="ts">
  import { app } from "../lib/app.svelte";
  import Pane from "./Pane.svelte";
  import SettingsPane from "./SettingsPane.svelte";
  import ProjectPane from "./ProjectPane.svelte";
  import FilesPane from "./FilesPane.svelte";

  const layout = $derived(["one", "two", "three", "four"][app.panes.length - 1]);
</script>

<main class={layout}>
  {#each app.panes as pane, i (pane.key)}
    {#if pane.kind === "settings"}<SettingsPane index={i} />
    {:else if pane.kind === "project"}<ProjectPane index={i} />
    {:else if pane.kind === "files"}<FilesPane index={i} />
    {:else}<Pane index={i} session={pane.session} />{/if}
  {/each}
</main>

<style>
  main { grid-area: main; display: grid; gap: 1px; background: var(--raised); min-height: 0; min-width: 0; }
  .one { grid-template: 1fr / 1fr; }
  .two { grid-template: 1fr / 1fr 1fr; }
  .three { grid-template: 1fr 1fr / 1fr 1fr; }
  .three > :global(:first-child) { grid-row: span 2; }
  .four { grid-template: 1fr 1fr / 1fr 1fr; }
</style>
