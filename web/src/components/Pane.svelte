<script lang="ts">
  import { app } from "../lib/app.svelte";
  import PaneHeader from "./PaneHeader.svelte";
  import Chat from "./Chat.svelte";
  import AskBlock from "./AskBlock.svelte";
  import TodoBlock from "./TodoBlock.svelte";
  import Composer from "./Composer.svelte";

  let { index, session }: { index: number; session: string } = $props();
  const view = $derived(app.sessions[session]);
  const focused = $derived(app.focused === index);
  let dock = $state(0);
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<section class:focused class:working={view?.state.busy} onclick={() => (app.focused = index)}>
  {#if view}
    <PaneHeader {index} {view} />
    <Chat {view} bottom={dock} />
    <div class="dock" bind:clientHeight={dock}>
      {#if view.state.todos.length}<TodoBlock todos={view.state.todos} />{/if}
      {#if view.state.ask?.length}<AskBlock id={session} questions={view.state.ask} />{/if}
      <Composer {view} {focused} />
    </div>
  {:else}
    <div class="loading label">[ starting ]</div>
  {/if}
</section>

<style>
  section {
    display: flex; flex-direction: column; min-height: 0; min-width: 0; background: var(--canvas);
    position: relative;
  }
  section::after {
    content: ""; position: absolute; inset: 0; pointer-events: none; border: 1px solid transparent; z-index: 2;
  }
  :global(main:not(.one)) > section.focused::after { border-color: var(--accent-deep); }
  section.working::after { border-color: var(--accent); box-shadow: inset var(--glow); }
  .dock {
    position: absolute; left: 0; right: 0; bottom: 0; z-index: 3; pointer-events: none;
    width: 100%; max-width: calc(var(--column) + 144px); margin: 0 auto; padding: 0 24px 10px; display: grid; gap: 8px;
  }
  .dock > :global(*) { pointer-events: auto; }
  .loading { display: grid; place-items: center; flex: 1; }
</style>
