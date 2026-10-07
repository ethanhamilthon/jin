<script lang="ts">
  import { focus } from "../../lib/focus";
  import { app } from "../../lib/app.svelte";
  import { commands } from "../../lib/commands";
  import { rank } from "../../lib/composer";
  import Dialog from "../Dialog.svelte";

  let { id }: { id: string } = $props();
  let search = $state("");
  let selected = $state(0);
  const found = $derived(rank(commands.map((c) => ({ ...c, label: c.name })), search));

  function run(index: number) {
    const command = found[index];
    if (!command) return;
    const session = id;
    app.dialog = null;
    command.run(session);
  }

  function onKey(event: KeyboardEvent) {
    if (event.key === "ArrowDown") selected = Math.min(found.length - 1, selected + 1);
    else if (event.key === "ArrowUp") selected = Math.max(0, selected - 1);
    else if (event.key === "Enter") run(selected);
    else return;
    event.preventDefault();
  }
</script>

<Dialog title="Commands" label="ctrl+k">
  <input class="field" placeholder="Type a command" bind:value={search} oninput={() => (selected = 0)} onkeydown={onKey} use:focus />
  <div class="rows">
    {#each found as command, i (command.name)}
      <button class="list-row" class:active={i === selected} onclick={() => run(i)} onmouseenter={() => (selected = i)}>
        <span class="mono name">/{command.name}</span><span class="soft">{command.description}</span>
      </button>
    {/each}
  </div>
</Dialog>

<style>
  .rows { margin-top: 10px; display: grid; gap: 2px; }
  .name { color: var(--text-strong); min-width: 90px; }
</style>
