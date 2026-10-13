<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { preview, type Tone } from "../../lib/sound";

  let tones = $state<Tone[]>([]);
  $effect(() => {
    let active = true;
    get<Tone[]>("/api/sounds").then((list) => { if (active) tones = list; }).catch(fail);
    return () => { active = false; };
  });

  async function select(tone: string) {
    const sound = { ...app.config.sound, tone };
    preview(sound);
    await post("/api/settings/sound", sound).catch(fail);
  }
</script>

<div class="set-row">
  <div class="set-text"><span class="set-name">Tone</span><span class="set-desc">one sound for questions, answers and errors</span></div>
  <div class="set-ctl">
    <select class="field" aria-label="Notification tone" value={app.config.sound.tone} onchange={(e) => select(e.currentTarget.value)}>
      {#each tones as tone (tone.id)}<option value={tone.id}>{tone.name}</option>{/each}
    </select>
    <button class="btn small" onclick={() => preview(app.config.sound)}>Preview</button>
  </div>
</div>

<style>
  select { width: 120px; }
</style>
