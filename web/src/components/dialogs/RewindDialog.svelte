<script lang="ts">
  import { app, fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";
  import { keep, show } from "../../lib/actions";
  import type { Point, Snapshot } from "../../lib/types";
  import Dialog from "../Dialog.svelte";

  let { id }: { id: string } = $props();
  let points = $state<Point[] | null>(null);

  $effect(() => {
    get<Point[]>(`/api/sessions/${id}/points`).then((p) => (points = p)).catch(fail);
  });

  async function fork(n: number) {
    try {
      const snap = await post<Snapshot>(`/api/sessions/${id}/fork`, { point: n });
      keep(snap);
      app.dialog = null;
      show(snap.state.id);
    } catch (err) {
      fail(err);
    }
  }
</script>

<Dialog title="Rewind" label="a new session from a message · files stay as they are">
  <div class="rows">
    {#each [...(points ?? [])].reverse() as point, i (point.index)}
      <button class="list-row" onclick={() => fork((points?.length ?? 0) - 1 - i)}>
        <span class="mono n">#{(points?.length ?? 0) - i}</span><span class="text">{point.text}</span>
      </button>
    {:else}<p class="empty">{points ? "Nothing to rewind" : "Loading…"}</p>{/each}
  </div>
</Dialog>

<style>
  .rows { display: grid; gap: 2px; }
  .n { color: var(--text-muted); min-width: 34px; }
  .text { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
