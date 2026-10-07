<script lang="ts">
  import Dialog from "../Dialog.svelte";
  import General from "../settings/General.svelte";
  import Tools from "../settings/Tools.svelte";
  import Scope from "../settings/Scope.svelte";
  import Prompts from "../settings/Prompts.svelte";
  import Hooks from "../settings/Hooks.svelte";
  import SystemPrompt from "../settings/SystemPrompt.svelte";
  import Data from "../settings/Data.svelte";

  let { tab = "general" }: { tab?: string } = $props();
  const tabs = [
    ["general", "General"], ["tools", "Tools"], ["scope", "Models"], ["prompts", "Prompts"],
    ["hooks", "Hooks"], ["system", "System prompt"], ["data", "Data folder"],
  ];
  let current = $state("general");
  $effect(() => {
    current = tab;
  });
</script>

<Dialog title="Settings" label="global · shared with the TUI" wide>
  <div class="layout">
    <nav>
      {#each tabs as [id, name] (id)}
        <button class="list-row" class:active={current === id} onclick={() => (current = id)}>{name}</button>
      {/each}
    </nav>
    <div class="pane">
      {#if current === "general"}<General />
      {:else if current === "tools"}<Tools />
      {:else if current === "scope"}<Scope />
      {:else if current === "prompts"}<Prompts />
      {:else if current === "hooks"}<Hooks />
      {:else if current === "system"}<SystemPrompt />
      {:else}<Data />{/if}
    </div>
  </div>
</Dialog>

<style>
  .layout { display: grid; grid-template-columns: 170px 1fr; gap: 20px; min-height: 380px; }
  nav { display: grid; align-content: start; gap: 2px; border-right: 1px solid var(--raised); padding-right: 12px; }
  .pane { min-width: 0; }
</style>
