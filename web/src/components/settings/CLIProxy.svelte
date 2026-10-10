<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";

  let version = $state("");
  let previous = $state("");
  let selected = $state("v8.0.23");
  let versions = $state<string[]>([]);
  let busy = $state(false);
  let loaded = $state(false);
  async function load() {
    const status = await get<{ installation: { version: string; previous?: string } }>("/api/cliproxy");
    version = status.installation.version ?? ""; previous = status.installation.previous ?? ""; loaded = true;
  }
  $effect(() => { void app.config.providers.length; load().catch(fail); });
  async function refresh() { versions = await get<string[]>("/api/cliproxy/versions").catch((e) => { fail(e); return []; }); }
  async function change(target = selected) {
    busy = true;
    try { await post(`/api/cliproxy/${version ? "update" : "install"}`, { version: target }); await load(); }
    catch (error) { fail(error); }
    finally { busy = false; }
  }
</script>

<div class="set-row">
  <div class="set-text"><span class="set-name">CLIProxyAPI</span><span class="set-desc">{!loaded ? "Loading…" : version ? `Installed: ${version}` : "Not installed"}</span></div>
  <div class="set-ctl"><button class="btn small" disabled={busy} onclick={refresh}>Check versions</button></div>
</div>
<div class="set-row">
  <div class="set-text"><label class="set-name" for="proxy-version">Version</label><span class="set-desc">Separate binary. Compatible v8.0.x releases only.</span></div>
  <div class="set-ctl">
    {#if versions.length}<select id="proxy-version" bind:value={selected}>{#each versions as v (v)}<option value={v}>{v}</option>{/each}</select>
    {:else}<input id="proxy-version" class="field mono" bind:value={selected} />{/if}
    <button class="btn small" disabled={busy || !loaded} onclick={() => change()}>{busy ? "Working…" : version ? "Update" : "Install"}</button>
  </div>
</div>
{#if previous}<div class="set-pad"><button class="btn" disabled={busy} onclick={() => change(previous)}>Roll back to {previous}</button></div>{/if}
<p class="set-note">Downloads official releases and verifies checksums. Updates wait for active requests. Jin itself is not replaced.</p>
<p class="set-note">Provider terms apply. A subscription is not a general-purpose API license.</p>

<style>input, select { width: 110px; }</style>
