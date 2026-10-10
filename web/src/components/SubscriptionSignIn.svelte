<script lang="ts">
  import { onDestroy, onMount } from "svelte";
  import { fail } from "../lib/app.svelte";
  import { SignIn } from "../lib/signin.svelte";

  let { profile, signedIn }: { profile: string; signedIn: () => void } = $props();
  const login = new SignIn(() => profile);

  onMount(() => { login.load().catch(fail); });
  $effect(() => { if (login.connected) signedIn(); });
  onDestroy(() => login.stop());
</script>

{#if !login.loaded}
  <p class="soft">Checking sign-in…</p>
{:else if login.pending}
  <p class="soft">Waiting for browser sign-in…</p>
  <button class="btn ghost small" onclick={() => login.cancel()}>Cancel</button>
{:else if !login.local}
  <p class="soft">Sign in on the computer running Jin.</p>
{:else}
  <button class="btn primary" disabled={login.busy} onclick={() => login.login()}>Sign in</button>
{/if}
