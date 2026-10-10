import { get, post } from "./api";
import { fail } from "./app.svelte";

export class SignIn {
  loaded = $state(false);
  local = $state(false);
  connected = $state(false);
  pending = $state("");
  busy = $state(false);
  #profile: () => string;
  #timer: ReturnType<typeof setTimeout> | undefined;

  constructor(profile: () => string) {
    this.#profile = profile;
  }

  async load() {
    const status = await get<{ local_login: boolean }>("/api/cliproxy");
    const accounts = await get<{ profile: string }[]>("/api/cliproxy/accounts");
    this.local = status.local_login;
    this.connected = accounts.some((a) => a.profile === this.#profile());
    this.loaded = true;
  }

  async login() {
    const popup = window.open("about:blank", "_blank");
    if (popup) popup.opener = null;
    this.busy = true;
    try {
      const result = await post<{ state: string; url: string }>("/api/cliproxy/login", { profile: this.#profile() });
      this.pending = result.state;
      if (popup) popup.location.href = result.url;
      else { await this.cancel(); throw new Error("Allow popups to sign in."); }
      this.#timer = setTimeout(() => this.#poll(), 1000);
    } catch (error) { popup?.close(); fail(error); }
    finally { this.busy = false; }
  }

  async cancel() {
    clearTimeout(this.#timer);
    const state = this.pending;
    this.pending = "";
    if (state) await post("/api/cliproxy/cancel", { state }).catch(fail);
  }

  async logout() {
    this.busy = true;
    try { await post("/api/cliproxy/logout", { profile: this.#profile() }); await this.load(); }
    catch (error) { fail(error); }
    finally { this.busy = false; }
  }

  stop() {
    clearTimeout(this.#timer);
    if (this.pending) void this.cancel();
  }

  async #poll() {
    if (!this.pending) return;
    try {
      const result = await post<{ status: string }>("/api/cliproxy/poll", { profile: this.#profile(), state: this.pending });
      if (result.status === "ok") { this.pending = ""; this.connected = true; return; }
      this.#timer = setTimeout(() => this.#poll(), 1000);
    } catch (error) { this.pending = ""; fail(error); }
  }
}
