<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { closeSection } from "../lib/panels";
  import PanelFrame from "./PanelFrame.svelte";
  import General from "./settings/General.svelte";
  import Tools from "./settings/Tools.svelte";
  import Scope from "./settings/Scope.svelte";
  import Prompts from "./settings/Prompts.svelte";
  import Hooks from "./settings/Hooks.svelte";
  import SystemPrompt from "./settings/SystemPrompt.svelte";
  import Title from "./settings/Title.svelte";
  import Data from "./settings/Data.svelte";
  import Archived from "./settings/Archived.svelte";
  import RemoteAccess from "./settings/RemoteAccess.svelte";
  import Providers from "./settings/Providers.svelte";
  import CLIProxy from "./settings/CLIProxy.svelte";

  let { index }: { index: number } = $props();
  const pane = $derived(app.panes[index]);
  const sections = [
    { id: "general", name: "General", hint: "Accent color, notification sound" },
    { id: "providers", name: "Providers", hint: "API connections, subscriptions and switches" },
    { id: "cliproxy", name: "CLIProxyAPI", hint: "Install or update the subscription proxy" },
    { id: "tools", name: "Tools", hint: "Switch tools on and off" },
    { id: "scope", name: "Models", hint: "Scope of the model picker" },
    { id: "prompts", name: "Prompts", hint: "#prompts you can call" },
    { id: "hooks", name: "Hooks", hint: "Text added to the system prompt" },
    { id: "system", name: "System prompt", hint: "System, compact and handoff" },
    { id: "titles", name: "Session titles", hint: "Name a session after its first exchange" },
    { id: "archived", name: "Archived projects", hint: "Restore a hidden project" },
    { id: "devices", name: "Remote access", hint: "Tailscale, pairing and connected devices" },
    { id: "data", name: "Data folder", hint: "Reset or swap" },
  ];
  const section = $derived(sections.find((s) => s.id === pane.section));
</script>

<PanelFrame {index} title={section?.name ?? "Settings"} subtitle={section ? "" : "global · shared with the TUI"} back={section ? closeSection : undefined}>
  {#if !section}
    {#each sections as s (s.id)}
      <button class="row" onclick={() => (pane.section = s.id)}>
        <span class="text">{s.name}<small>{s.hint}</small></span>
        <span class="go">›</span>
      </button>
    {/each}
  {:else}
    <div class="section">
      {#if section.id === "general"}<General />
      {:else if section.id === "providers"}<Providers />
      {:else if section.id === "cliproxy"}<CLIProxy />
      {:else if section.id === "tools"}<Tools />
      {:else if section.id === "scope"}<Scope />
      {:else if section.id === "prompts"}<Prompts />
      {:else if section.id === "hooks"}<Hooks />
      {:else if section.id === "system"}<SystemPrompt />
      {:else if section.id === "titles"}<Title />
      {:else if section.id === "archived"}<Archived />
      {:else if section.id === "devices"}<RemoteAccess />
      {:else}<Data />{/if}
    </div>
  {/if}
</PanelFrame>

<style>
  .row {
    display: flex; align-items: center; gap: 10px; width: 100%; padding: 11px 16px; text-align: left;
    background: transparent; border: 0; border-bottom: 1px solid var(--raised); cursor: pointer; color: var(--text);
  }
  .row:hover { background: var(--raised); }
  .text { flex: 1; min-width: 0; display: flex; flex-direction: column; color: var(--text-strong); }
  small { color: var(--text-muted); font-size: 12px; }
  .go { color: var(--text-muted); }
  .section { padding-bottom: 20px; }
</style>
