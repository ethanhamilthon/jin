<script lang="ts">
  import { app } from "../lib/app.svelte";
  import PaneHeader from "./PaneHeader.svelte";
  import Chat from "./Chat.svelte";
  import AskBlock from "./AskBlock.svelte";
  import Composer from "./Composer.svelte";
  import NoProvider from "./NoProvider.svelte";
  import StatusGlow from "./StatusGlow.svelte";
  import { noEnabledProvider } from "../lib/provider-access";

  let { index, session }: { index: number; session: string } = $props();
  const view = $derived(app.sessions[session]);
  const focused = $derived(app.focused === index);
  const live = $derived(view?.state);
  const busy = $derived(!!live?.busy && !live?.shell && !live?.reloading && !!live?.ready);
  const tasks = $derived((live?.tasks ?? 0) > 0);
  let dock = $state(0);
  let root: HTMLElement;

  $effect(() => {
    if (focused && !root.contains(document.activeElement)) root.focus();
  });
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<section bind:this={root} tabindex="-1" onfocusin={() => (app.focused = index)}>
  {#if view}
    <PaneHeader {index} {view} />
    {#if busy || tasks}<StatusGlow {busy} {tasks} />{/if}
    <Chat {view} bottom={dock} />
    <div class="dock" bind:clientHeight={dock}>
      {#if view.state.ask?.length}<AskBlock id={session} questions={view.state.ask} />{/if}
      {#if noEnabledProvider(app.config.providers, view.state.provider)}<NoProvider />{/if}
      <Composer {view} {focused} />
    </div>
  {:else}
    <div class="loading label">[ starting ]</div>
  {/if}
</section>

<style>
  section {
    display: flex; flex-direction: column; min-height: 0; min-width: 0; background: var(--canvas);
    position: relative; outline: none;
  }
  .dock {
    position: absolute; left: 0; right: 0; bottom: 0; z-index: 3; pointer-events: none;
    width: 100%; max-width: calc(var(--column) + 144px); margin: 0 auto; padding: 0 24px 10px; display: grid; gap: 8px;
  }
  .dock > :global(*) { pointer-events: auto; }
  @media (max-width: 700px) { .dock { max-width: none; padding: 0; gap: 0; } }
  .loading { display: grid; place-items: center; flex: 1; }
</style>
