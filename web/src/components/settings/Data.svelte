<script lang="ts">
  import { fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";

  let current = $state("");
  let kind = $state<"reset" | "swap" | "">("");
  let path = $state("");

  $effect(() => {
    get<{ current: string }>("/api/data").then((d) => ((current = d.current), (path = d.current + "-backup"))).catch(fail);
  });

  async function apply(event: Event) {
    event.preventDefault();
    await post("/api/data", { kind, path }).catch(fail);
  }
</script>

<p class="soft">jin keeps its settings, providers, prompts, hooks and sessions in <span class="mono">{current}</span>.</p>
<div class="choices">
  <button class="list-row" class:active={kind === "reset"} onclick={() => ((kind = "reset"), (path = current + "-backup"))}>
    <span><strong>Reset</strong><br /><span class="soft">Move all jin data aside and start from scratch</span></span>
  </button>
  <button class="list-row" class:active={kind === "swap"} onclick={() => ((kind = "swap"), (path = ""))}>
    <span><strong>Swap</strong><br /><span class="soft">Use another data folder; the current data moves to its old place</span></span>
  </button>
</div>
{#if kind}
  <form onsubmit={apply}>
    <label class="label" for="data-path">{kind === "reset" ? "Move the current data to" : "Folder to use instead"}</label>
    <input id="data-path" class="field mono" bind:value={path} required />
    <p class="warn">jin web stops after the move. Run it again to continue.</p>
    <button class="btn danger">{kind === "reset" ? "Reset jin" : "Swap data folders"}</button>
  </form>
{/if}

<style>
  .choices { display: grid; gap: 6px; margin: 12px 0; }
  .choices .list-row { border: 1px solid var(--raised); }
  form { display: grid; gap: 8px; }
  .warn { color: var(--warn); font-size: 12px; margin: 0; }
</style>
