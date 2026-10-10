<script lang="ts">
  import { onDestroy } from "svelte";
  import { fail } from "../../lib/app.svelte";
  import { get, post } from "../../lib/api";

  let { profile }: { profile: string } = $props();
  let connected = $state(false);
  let local = $state(false);
  let pending = $state("");
  let busy = $state(false);
  let timer: ReturnType<typeof setTimeout> | undefined;

  async function load() {
    const status = await get<{ local_login: boolean }>("/api/cliproxy");
    local = status.local_login;
    const accounts = await get<{ profile: string }[]>("/api/cliproxy/accounts");
    connected = accounts.some((a) => a.profile === profile);
  }
  $effect(() => { void profile; load().catch(fail); });
  async function poll() {
    if (!pending) return;
    try {
      const result = await post<{ status: string }>("/api/cliproxy/poll", { profile, state: pending });
      if (result.status === "ok") { pending = ""; await load(); return; }
      timer = setTimeout(poll, 1000);
    } catch (error) { pending = ""; fail(error); }
  }
  async function login() {
    const popup = window.open("about:blank", "_blank");
    if (popup) popup.opener = null;
    busy = true;
    try {
      const result = await post<{ state: string; url: string }>("/api/cliproxy/login", { profile });
      pending = result.state;
      if (popup) popup.location.href = result.url;
      else { await cancel(); throw new Error("Allow popups to sign in."); }
      timer = setTimeout(poll, 1000);
    } catch (error) { popup?.close(); fail(error); }
    finally { busy = false; }
  }
  async function cancel() {
    clearTimeout(timer);
    const state = pending; pending = "";
    if (state) await post("/api/cliproxy/cancel", { state }).catch(fail);
  }
  async function logout() {
    busy = true;
    try { await post("/api/cliproxy/logout", { profile }); await load(); }
    catch (error) { fail(error); }
    finally { busy = false; }
  }
  onDestroy(() => { clearTimeout(timer); if (pending) void cancel(); });
</script>

<div class="set-row">
  <div class="set-text"><span class="set-desc">{pending ? "Waiting for browser sign-in…" : connected ? "Account connected" : "Not signed in"}</span></div>
  <div class="set-ctl">
    {#if pending}<button class="btn small" onclick={cancel}>Cancel</button>
    {:else if connected}<button class="btn small" disabled={busy} onclick={logout}>Sign out</button>
    {:else}<button class="btn small" disabled={busy || !local} onclick={login}>Sign in</button>{/if}
  </div>
</div>
{#if !local && !connected}<p class="set-note">Sign in on the computer running Jin.</p>{/if}
