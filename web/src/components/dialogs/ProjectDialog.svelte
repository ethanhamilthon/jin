<script lang="ts">
  import { onMount } from "svelte";
  import { app, fail } from "../../lib/app.svelte";
  import { get, post, query } from "../../lib/api";
  import { newSession } from "../../lib/actions";
  import { shortPath } from "../../lib/format";
  import type { Folders, Project } from "../../lib/types";
  import Dialog from "../Dialog.svelte";
  import FolderList from "./FolderList.svelte";

  const storageKey = "jin.lastFolder";
  let folders = $state<Folders | null>(null);
  let error = $state("");
  let hidden = $state(false);
  let adding = $state(false);

  async function load(path: string) {
    try {
      folders = await get<Folders>("/api/dirs" + query({ path, hidden: hidden ? "1" : undefined }));
      error = "";
    } catch (err) {
      error = err instanceof Error ? err.message : String(err);
    }
  }

  async function start(saved: string) {
    await load(saved);
    if (error && saved) await load("");
  }

  onMount(() => {
    void start(localStorage.getItem(storageKey) ?? "");
  });

  async function add() {
    if (!folders) return;
    adding = true;
    try {
      const project = await post<Project>("/api/projects", { path: folders.path });
      localStorage.setItem(storageKey, project.path);
      app.dialog = null;
      app.project = project.path;
      await newSession(project.path);
    } catch (err) {
      fail(err);
    } finally {
      adding = false;
    }
  }
</script>

<Dialog title="Add a project" label="a folder on this machine">
  {#if folders}
    <p class="current mono" title={folders.path}><bdi>{shortPath(folders.path, app.home)}</bdi></p>
    <FolderList {folders} open={load} />
    <label class="hidden"><input type="checkbox" checked={hidden} onchange={(e) => { hidden = e.currentTarget.checked; load(folders?.path ?? ""); }} /> Show hidden folders</label>
  {:else if !error}<p class="empty">Loading folders…</p>{/if}
  {#if error}<p class="error">{error}</p>{/if}
  <div class="actions"><button class="btn primary" disabled={!folders || adding} onclick={add}>Use this folder</button></div>
</Dialog>

<style>
  .current { margin: 0 0 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; color: var(--text-strong); }
  .hidden { display: flex; align-items: center; gap: 8px; min-height: 44px; font-size: 13px; color: var(--text-dim); cursor: pointer; }
  .actions { display: flex; justify-content: flex-end; margin-top: 8px; }
  .error { color: var(--error); font-size: 13px; }
  @media (max-width: 700px) {
    .actions .btn { min-height: 44px; }
  }
</style>
