<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";
  import { devices, others } from "../../lib/devices.svelte";
  import { openPanel } from "../../lib/panels";
  import Dialog from "../Dialog.svelte";

  let pairing = $state<{ qr: string; url: string; expires: number } | null>(null);
  let timer: ReturnType<typeof setTimeout> | undefined;

  async function renew() {
    clearTimeout(timer);
    if (!devices.remote) return;
    pairing = await post<{ qr: string; url: string; expires: number }>("/api/devices/pair");
    timer = setTimeout(() => renew().catch(fail), Math.max(1000, pairing.expires * 1000 - Date.now() - 5000));
  }

  // A scan or a closed page changes the devices; the code that was used is gone.
  $effect(() => {
    void devices.list.length;
    renew().catch(fail);
    return () => clearTimeout(timer);
  });

  function manage() {
    app.dialog = null;
    openPanel("settings", "devices");
  }
</script>

<Dialog title="Connect a phone" label="remote access">
  {#if !devices.remote}
    <p class="text">Remote access is off. Install Tailscale on this computer and your phone, sign in to both, then enable it in Settings / Remote access. No restart is needed.</p>
  {:else}
    <div class="qr">{#if pairing}<img src={pairing.qr} alt="Pairing QR code" width="280" height="280" />{/if}</div>
    <p class="text">Scan with the phone camera. The code works once and expires in 5 minutes; it renews itself.</p>
    <p class="text soft">{others() ? `${others()} device${others() === 1 ? "" : "s"} connected now.` : "No other device is connected."}</p>
  {/if}
  <button class="btn small" onclick={manage}>Remote access</button>
</Dialog>

<style>
  .qr { display: grid; place-items: center; padding: 12px; background: #fff; width: 304px; height: 304px; margin: 0 auto 12px; }
  .text { margin: 0 0 10px; }
</style>
