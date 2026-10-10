<script lang="ts">
  import { app } from "../lib/app.svelte";
  import { loadState } from "../lib/actions";
  import ProviderForm from "./ProviderForm.svelte";
</script>

<div class="onboarding">
  <div class="hero">
    <p class="label">[ jin {app.version} · first run ]</p>
    <h1 class="serif">A coding agent, <span class="accent-word">in your browser.</span></h1>
    <p class="lead">Connect a model API or a subscription to start. Jin keeps the provider with your sessions, prompts and hooks; the terminal UI uses the same ones.</p>
    <ol class="soft">
      <li>Choose API or subscription, then one of three options</li>
      <li>For an API: a name, the base URL and the key; for a subscription: install and sign in</li>
      <li>Pick a model and a reasoning effort</li>
    </ol>
  </div>
  <div class="card">
    <p class="label">[ connect a model ]</p>
    <ProviderForm done={() => loadState()} />
  </div>
</div>
<div class="divider"></div>

<style>
  .onboarding {
    min-height: calc(100% - 6px); display: grid; grid-template-columns: minmax(0, 1.1fr) minmax(0, 1fr); gap: 48px;
    align-items: center; max-width: 1180px; margin: 0 auto; padding: 48px 32px;
  }
  h1 { font-size: 64px; line-height: 0.95; letter-spacing: -0.025em; margin: 12px 0 20px; }
  .lead { max-width: 480px; color: var(--text); }
  ol { padding-left: 20px; }
  .card { background: var(--card); border: 1px solid var(--raised); padding: 28px; display: grid; gap: 14px; }
  .divider { height: 6px; background: var(--accent); box-shadow: var(--glow); }
  @media (max-width: 860px) {
    .onboarding { grid-template-columns: 1fr; gap: 24px; }
    h1 { font-size: 44px; }
  }
</style>
