<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";
  import { accents, applyAccent, defaultAccent } from "../../lib/accent";
  import { ring } from "../../lib/sound";
  import type { Sound } from "../../lib/types";
  import Switch from "../Switch.svelte";

  const accent = $derived(app.config.accent || defaultAccent);

  async function setAccent(value: string) {
    applyAccent(value);
    await post("/api/settings/accent", { accent: value }).catch(fail);
  }

  async function setSound(change: Partial<Sound>) {
    const sound = { ...app.config.sound, ...change };
    await post("/api/settings/sound", sound).catch(fail);
    if (change.volume !== undefined) ring({ ...sound, only_blur: false }, true);
  }
</script>

<div class="set-hd"><span class="label">Accent</span><span class="hint">one color drives the whole interface</span></div>
<div class="swatches">
  {#each accents as a (a.value)}
    <button class="swatch" class:on={accent === a.value} style="--c: {a.value}" title={a.name} aria-label={a.name} onclick={() => setAccent(a.value)}></button>
  {/each}
  <label class="custom" title="Custom color">
    <input type="color" value={accent} onchange={(e) => setAccent(e.currentTarget.value)} />
    <span class="mono">{accent}</span>
  </label>
</div>

<div class="set-hd"><span class="label">Notifications</span></div>
<div class="set-row">
  <div class="set-text"><span class="set-name">Sound</span><span class="set-desc">when an answer is done or the agent asks</span></div>
  <div class="set-ctl"><Switch label="Sound" checked={app.config.sound.enabled} onchange={(v) => setSound({ enabled: v })} /></div>
</div>
<div class="set-row">
  <div class="set-text"><span class="set-name">Only when the page is not focused</span></div>
  <div class="set-ctl"><Switch label="Only when the page is not focused" checked={app.config.sound.only_blur} onchange={(v) => setSound({ only_blur: v })} /></div>
</div>
<div class="set-row">
  <div class="set-text"><span class="set-name">Volume</span></div>
  <div class="set-ctl">
    <input type="range" min="10" max="100" step="5" value={app.config.sound.volume} onchange={(e) => setSound({ volume: Number(e.currentTarget.value) })} />
    <span class="mono">{app.config.sound.volume}%</span>
  </div>
</div>

<style>
  .swatches { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; padding: 0 16px 14px; }
  .swatch { width: 26px; height: 26px; border-radius: var(--radius); background: var(--c); border: 2px solid transparent; cursor: pointer; }
  .swatch.on { border-color: var(--text-strong); box-shadow: 0 0 8px var(--c); }
  .custom { display: flex; align-items: center; gap: 8px; color: var(--text-soft); margin-left: 6px; }
  .custom input { width: 32px; height: 28px; border: 0; background: none; padding: 0; }
  input[type="range"] { accent-color: var(--accent); width: 120px; }
</style>
