<script lang="ts">
  import { onMount } from "svelte";
  import { app, fail } from "./lib/app.svelte";
  import { loadState, newSession } from "./lib/actions";
  import { connect } from "./lib/events";
  import { keys } from "./lib/keys";
  import { restoreLayout, saveLayout } from "./lib/layout";
  import { followViewport } from "./lib/mobile.svelte";
  import { onBack, syncBackLayer } from "./lib/back";
  import TopBar from "./components/TopBar.svelte";
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
      if (app.config.ready || app.config.providers.length) await restoreLayout();
    } catch (err) {
      fail(err);
    }
    app.restored = true;
  });

  $effect(() => {
    if (app.restored) saveLayout();
  });
  $effect(syncBackLayer);

  let creating = false;
  $effect(() => {
    if (!app.restored || !app.config.model || app.chat?.session || creating) return;
    creating = true;
    newSession(app.project || app.dir).finally(() => (creating = false));
  });
</script>

<svelte:window onkeydown={keys} onpopstate={onBack} />

{#if app.stopped}
  <Stopped />
{:else if !app.loaded}
  <div class="boot label">[ connecting ]</div>
{:else if !app.config.model}
  <Onboarding />
{:else}
  <div class="shell">
    <TopBar />
    <Workspace />
  </div>
{/if}
<Dialogs />
<Toasts />

<style>
  .shell {
    display: grid; height: 100%; overflow: clip;
    grid-template: "top" 48px "main" 1fr / 1fr;
  }
  .boot { display: grid; place-items: center; height: 100%; }
</style>
