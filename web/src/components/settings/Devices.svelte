<script lang="ts">
  import { fail } from "../../lib/app.svelte";
  import { ago } from "../../lib/format";
  import { patch, del } from "../../lib/api";
  import { devices, loadDevices } from "../../lib/devices.svelte";
  import type { Device } from "../../lib/types";

  const seen = (d: Device) => (d.online ? "online now" : "seen " + ago(new Date(d.lastSeen * 1000).toISOString()));
  const reload = () => loadDevices().catch(fail);

  function rename(d: Device, name: string) {
    name = name.trim();
    if (name && name !== d.name) patch(`/api/devices/${d.id}`, { name }).then(reload).catch(fail);
  }

  const revoke = (d: Device) => del(`/api/devices/${d.id}`).then(reload).catch(fail);
</script>

<div class="set-hd"><span class="label">Devices</span><span class="hint">browsers that can open this jin web; revoke one to lock it out</span></div>
{#each devices.list as d (d.id)}
  <div class="set-row">
    <span class="dot" class:dim={!d.online}></span>
    <div class="set-text">
      <input class="field name" value={d.name} aria-label="Device name" onchange={(e) => rename(d, e.currentTarget.value)} />
      <span class="set-desc">{seen(d)}{d.current ? " · this browser" : ""}</span>
    </div>
    <div class="set-ctl"><button class="btn danger small" onclick={() => revoke(d)}>Revoke</button></div>
  </div>
{:else}<p class="empty">No devices</p>{/each}

<style>
  .dot.dim { background: var(--line); }
  .name { width: 100%; }
</style>
