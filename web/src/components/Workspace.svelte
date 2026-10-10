<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { mobile } from "../lib/mobile.svelte";
  import { splitOf, setSplit, persistSizes } from "../lib/paneSizes.svelte";
  import Pane from "./Pane.svelte";
  import PaneDivider from "./PaneDivider.svelte";
  import SettingsPane from "./SettingsPane.svelte";
  import ProjectPane from "./ProjectPane.svelte";
  import FilesPane from "./FilesPane.svelte";

  const layout = $derived(["one", "two", "three", "four"][app.panes.length - 1]);
  const split = $derived(splitOf(layout));
  const grid = $derived(mobile.on || layout === "one" ? "" :
    `grid-template-columns: ${split.col}fr ${1 - split.col}fr;` +
    (layout === "two" ? "" : ` grid-template-rows: ${split.row}fr ${1 - split.row}fr;`));
  const dividers = $derived(!mobile.on && layout !== "one");
</script>

<main class={layout} style={grid}>
  {#each app.panes as pane, i (pane.key)}
    {#if !mobile.on || i === app.focused}
      {#if pane.kind === "settings"}<SettingsPane index={i} />
      {:else if pane.kind === "project"}<ProjectPane index={i} />
      {:else if pane.kind === "files"}<FilesPane index={i} />
      {:else}<Pane index={i} session={pane.session} />{/if}
    {/if}
  {/each}
  {#if dividers}
    <PaneDivider orientation="col" pos={split.col} onDrag={(d) => setSplit(layout, split.col + d, split.row)} onRelease={persistSizes} />
    {#if layout !== "two"}
      <PaneDivider orientation="row" pos={split.row} from={layout === "three" ? split.col : 0} onDrag={(d) => setSplit(layout, split.col, split.row + d)} onRelease={persistSizes} />
    {/if}
  {/if}
</main>

<style>
  main { position: relative; grid-area: main; display: grid; gap: 1px; background: var(--raised); min-height: 0; min-width: 0; }
  .one { grid-template: 1fr / 1fr; }
  .two { grid-template: 1fr / 1fr 1fr; }
  .three { grid-template: 1fr 1fr / 1fr 1fr; }
  .three > :global(:first-child) { grid-row: span 2; }
  .four { grid-template: 1fr 1fr / 1fr 1fr; }
  @media (max-width: 700px) {
    main.one, main.two, main.three, main.four { grid-template: 1fr / 1fr; }
    .three > :global(:first-child) { grid-row: auto; }
  }
</style>
