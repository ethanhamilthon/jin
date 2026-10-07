<script lang="ts">
  import { tick } from "svelte";
  import { app, type SessionView } from "../lib/app.svelte";
  import { shows } from "../lib/fold";
  import EntryView from "./EntryView.svelte";
  import Intro from "./Intro.svelte";

  let { view }: { view: SessionView } = $props();
  let scroller: HTMLDivElement;
  let pinned = $state(true);

  function onScroll() {
    pinned = scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 40;
  }

  $effect(() => {
    void view.entries.length, view.entries[view.entries.length - 1]?.text, view.state.id;
    if (pinned) tick().then(() => scroller && (scroller.scrollTop = scroller.scrollHeight));
  });

  $effect(() => {
    void view.state.id;
    pinned = true;
  });
</script>

<div class="scroll" bind:this={scroller} onscroll={onScroll}>
  <div class="column">
    {#if view.intro && !view.state.persisted}
      <Intro intro={view.intro} loading={view.state.loading ?? []} />
    {/if}
    {#each view.entries as entry, i (i)}
      {#if shows(app.config.fold, entry.kind)}
        <EntryView {entry} streaming={view.state.working && i === view.entries.length - 1} />
      {/if}
    {/each}
  </div>
</div>
{#if !pinned}
  <button class="jump btn small" onclick={() => { pinned = true; scroller.scrollTop = scroller.scrollHeight; }}>↓ Latest</button>
{/if}

<style>
  .scroll { flex: 1; overflow-y: auto; min-height: 0; }
  .column { max-width: calc(var(--column) + 48px); margin: 0 auto; padding: 20px 24px 12px; display: grid; gap: 12px; }
  .jump { position: absolute; right: 24px; bottom: 150px; z-index: 3; }
</style>
