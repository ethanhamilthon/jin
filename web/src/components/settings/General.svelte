<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";
  import { accents, applyAccent, defaultAccent } from "../../lib/accent";
  import { ring } from "../../lib/sound";
  import type { Sound } from "../../lib/types";

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

<section>
  <h3 class="label">Accent</h3>
  <div class="swatches">
    {#each accents as a (a.value)}
      <button class="swatch" class:on={accent === a.value} style="--c: {a.value}" title={a.name} aria-label={a.name} onclick={() => setAccent(a.value)}></button>
    {/each}
    <label class="custom" title="Custom color">
      <input type="color" value={accent} onchange={(e) => setAccent(e.currentTarget.value)} />
      <span class="mono">{accent}</span>
    </label>
  </div>
</section>

<section>
  <h3 class="label">Notification sound</h3>
  <label class="row"><input type="checkbox" checked={app.config.sound.enabled} onchange={(e) => setSound({ enabled: e.currentTarget.checked })} /> Play a sound when an answer is done or the agent asks</label>
  <label class="row"><input type="checkbox" checked={app.config.sound.only_blur} onchange={(e) => setSound({ only_blur: e.currentTarget.checked })} /> Only when the page is not focused</label>
  <label class="row">Volume
    <input type="range" min="10" max="100" step="5" value={app.config.sound.volume} onchange={(e) => setSound({ volume: Number(e.currentTarget.value) })} />
    <span class="mono">{app.config.sound.volume}%</span>
  </label>
</section>

<style>
  section { margin-bottom: 24px; }
  h3 { margin: 0 0 10px; }
  .swatches { display: flex; gap: 10px; align-items: center; flex-wrap: wrap; }
  .swatch { width: 28px; height: 28px; border-radius: var(--radius); background: var(--c); border: 2px solid transparent; cursor: pointer; }
  .swatch.on { border-color: var(--text-strong); box-shadow: 0 0 8px var(--c); }
  .custom { display: flex; align-items: center; gap: 8px; color: var(--text-soft); }
  .custom input { width: 32px; height: 28px; border: 0; background: none; padding: 0; }
  .row { display: flex; gap: 10px; align-items: center; margin: 6px 0; }
  input[type="range"] { accent-color: var(--accent); }
</style>
