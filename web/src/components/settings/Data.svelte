<script lang="ts">
  import { fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";

  let current = $state("");
  let kind = $state<"reset" | "swap" | "">("");
  let path = $state("");

  $effect(() => {
    get<{ current: string }>("/api/data").then((d) => ((current = d.current), (path = d.current + "-backup"))).catch(fail);
  });

  function choose(next: "reset" | "swap") {
    kind = kind === next ? "" : next;
    path = next === "reset" ? current + "-backup" : "";
  }

  async function apply(event: Event) {
    event.preventDefault();
    await post("/api/data", { kind, path }).catch(fail);
  }
</script>

<div class="set-hd"><span class="label">Data folder</span><span class="hint mono">{current}</span></div>
<p class="set-lead">Settings, providers, prompts, hooks and sessions live here, shared with the TUI.</p>
<div class="set-row">
  <div class="set-text"><span class="set-name">Reset</span><span class="set-desc">Move all jin data aside and start from scratch</span></div>
  <div class="set-ctl"><button class="btn small" class:on={kind === "reset"} onclick={() => choose("reset")}>Reset…</button></div>
</div>
<div class="set-row">
  <div class="set-text"><span class="set-name">Swap</span><span class="set-desc">Use another data folder; the current data moves to its old place</span></div>
  <div class="set-ctl"><button class="btn small" class:on={kind === "swap"} onclick={() => choose("swap")}>Swap…</button></div>
</div>
{#if kind}
  <form class="form" onsubmit={apply}>
    <label class="label" for="data-path">{kind === "reset" ? "Move the current data to" : "Folder to use instead"}</label>
    <input id="data-path" class="field mono" bind:value={path} required />
    <p class="warn">jin web stops after the move. Run it again to continue.</p>
    <button class="btn danger">{kind === "reset" ? "Reset jin" : "Swap data folders"}</button>
  </form>
{/if}

<style>
  .form { display: grid; gap: 8px; padding: 14px 16px; }
  .warn { color: var(--warn); font-size: 12px; margin: 0; }
  .on { border-color: var(--accent); color: var(--accent); }
</style>
