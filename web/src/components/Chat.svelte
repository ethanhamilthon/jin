<script lang="ts">
  import { tick } from "svelte";
  import { app, type SessionView } from "../lib/app.svelte";
  import { shows } from "../lib/fold";
  import { Pin } from "../lib/pin";
  import EntryView from "./EntryView.svelte";
  import Intro from "./Intro.svelte";

  let { view, bottom = 0 }: { view: SessionView; bottom?: number } = $props();
  let scroller: HTMLDivElement;
  const pin = new Pin();
  const keys = new Set(["PageUp", "PageDown", "Home", "End", "ArrowUp", "ArrowDown", " "]);

  function onScroll() {
    pin.scrolled(scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight, scroller.scrollTop, performance.now());
  }

  let seen = { id: "", users: 0 };
  $effect(() => {
    const users = view.entries.filter((entry) => entry.kind === "user").length;
    if (view.state.id !== seen.id || users > seen.users) pin.reset();
    seen = { id: view.state.id, users };
  });

  $effect(() => {
    void view.entries.length, view.entries[view.entries.length - 1]?.text, view.state.id;
    if (pin.following) tick().then(() => pin.following && scroller && (scroller.scrollTop = scroller.scrollHeight));
  });
</script>

<svelte:window onkeydown={(e) => keys.has(e.key) && pin.touch(performance.now())} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  class="scroll" bind:this={scroller} onscroll={onScroll}
  onwheel={(e) => (e.deltaY < 0 ? pin.up(performance.now()) : pin.touch(performance.now()))}
  ontouchmove={() => pin.touch(performance.now())}
  onpointerdown={() => pin.touch(performance.now())}
  onpointermove={(e) => e.buttons && pin.touch(performance.now())}
>
  <div class="column" style:padding-bottom="{bottom + 12}px">
    {#if view.intro && !view.state.persisted}
      <Intro intro={view.intro} loading={view.state.loading ?? []} />
    {/if}
    {#each view.entries as entry, i (i)}
      {#if shows(app.config.fold, entry.kind)}
        <EntryView {entry} session={view.state.id} index={i} />
      {/if}
    {/each}
  </div>
</div>

<style>
  .scroll { flex: 1; overflow-y: auto; min-height: 0; }
  .column { grid-template-columns: minmax(0, 1fr); max-width: calc(var(--column) + 48px); margin: 0 auto; padding: 20px 24px 12px; display: grid; gap: 12px; }
</style>
