<script lang="ts">
  import { onMount } from "svelte";
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { devices, loadDevices } from "../../lib/devices.svelte";
  import Devices from "./Devices.svelte";

  let remote = $state<{ enabled: boolean; url?: string; error?: string }>({ enabled: false });
  let busy = $state(false);
  const load = async () => { remote = await get("/api/remote"); };
  onMount(() => { load().catch(fail); });
  $effect(() => { void devices.remote; load().catch(fail); });

  async function toggle() {
    busy = true;
    try {
      remote = await post("/api/remote", { enabled: !remote.enabled });
      await loadDevices();
    } catch (error) { fail(error); await load(); }
    finally { busy = false; }
  }
</script>

<div class="set-hd"><span class="label">Remote access</span><span class="hint">Tailscale access stays available while the daemon runs, even without local clients</span></div>
<div class="set-row">
  <div class="set-text"><span>Enable remote access</span><span class="set-desc">Install Tailscale and sign in on this computer and your phone. HTTPS must be enabled.</span></div>
  <div class="set-ctl"><button class="btn small" disabled={busy} onclick={toggle}>{busy ? "Updating…" : remote.enabled ? "Disable" : "Enable"}</button></div>
</div>
{#if remote.error}<p class="empty">{remote.error}</p>{/if}
{#if remote.enabled}
  <div class="set-row"><div class="set-text"><span class="set-desc">{remote.url}</span></div><button class="btn small" onclick={() => (app.dialog = { name: "pair" })}>Connect a phone</button></div>
{/if}
<Devices />
