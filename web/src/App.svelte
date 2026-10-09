<script lang="ts">
  import { onMount } from "svelte";
  import { app, fail } from "./lib/app.svelte";
  import { loadState, newSession } from "./lib/actions";
  import { connect } from "./lib/events";
  import { keys } from "./lib/keys";
  import { restoreLayout, saveLayout, saveSidebar } from "./lib/layout";
  import { followViewport, mobile } from "./lib/mobile.svelte";
  import { onBack, syncBackLayer } from "./lib/back";
  import TopBar from "./components/TopBar.svelte";
  import Sidebar from "./components/Sidebar.svelte";
  import Workspace from "./components/Workspace.svelte";
  import Dialogs from "./components/Dialogs.svelte";
  import Toasts from "./components/Toasts.svelte";
  import Onboarding from "./components/Onboarding.svelte";
  import Stopped from "./components/Stopped.svelte";

  onMount(async () => {
    followViewport();
    try {
      await connect();
      await loadState();
      if (app.config.ready) await restoreLayout();
    } catch (err) {
      fail(err);
    }
    app.restored = true;
  });

  $effect(() => {
    if (app.restored) saveLayout();
  });
  $effect(() => saveSidebar(app.sidebar));

  // On a phone the drawer closes when a session is chosen, and starts closed.
  $effect(() => {
    void app.chat?.session;
    if (mobile.on) app.sidebar = false;
  });
  $effect(syncBackLayer);

  let creating = false;
  $effect(() => {
    if (!app.restored || !app.config.ready || app.chat?.session || creating) return;
    creating = true;
    newSession(app.project || app.dir).finally(() => (creating = false));
  });
</script>

<svelte:window onkeydown={keys} onpopstate={onBack} />

{#if app.stopped}
  <Stopped />
{:else if !app.loaded}
  <div class="boot label">[ connecting ]</div>
{:else if !app.config.ready}
  <Onboarding />
{:else}
  <div class="shell" class:collapsed={!app.sidebar}>
    <TopBar />
    {#if app.sidebar}
      {#if mobile.on}<button class="scrim" aria-label="Close sidebar" onclick={() => (app.sidebar = false)}></button>{/if}
      <Sidebar />
    {/if}
    <Workspace />
  </div>
{/if}
<Dialogs />
<Toasts />

<style>
  .shell {
    display: grid; height: 100%; overflow: clip;
    grid-template: "top top" 48px "side main" 1fr / 320px 1fr;
  }
  .shell.collapsed { grid-template: "top" 48px "main" 1fr / 1fr; }
  .scrim { position: fixed; inset: 0; z-index: 14; border: 0; background: rgb(0 0 0 / 0.6); }
  @media (max-width: 700px) {
    .shell, .shell.collapsed { grid-template: "top" 48px "main" 1fr / 1fr; }
  }
  .boot { display: grid; place-items: center; height: 100%; }
</style>
