<script lang="ts">
  import { onDestroy } from "svelte";
  import { fail } from "../../lib/app.svelte";
  import { SignIn } from "../../lib/signin.svelte";

  let { profile }: { profile: string } = $props();
  const login = new SignIn(() => profile);
  $effect(() => { void profile; login.load().catch(fail); });
  onDestroy(() => login.stop());
</script>

<div class="set-row">
  <div class="set-text"><span class="set-desc">{login.pending ? "Waiting for browser sign-in…" : login.connected ? "Account connected" : "Not signed in"}</span></div>
  <div class="set-ctl">
    {#if login.pending}<button class="btn small" onclick={() => login.cancel()}>Cancel</button>
    {:else if login.connected}<button class="btn small" disabled={login.busy} onclick={() => login.logout()}>Sign out</button>
    {:else}<button class="btn small" disabled={login.busy || !login.local} onclick={() => login.login()}>Sign in</button>{/if}
  </div>
</div>
{#if !login.local && !login.connected}<p class="set-note">Sign in on the computer running Jin.</p>{/if}
