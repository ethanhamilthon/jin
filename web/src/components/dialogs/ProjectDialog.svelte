<script lang="ts">
  import { focus } from "../../lib/focus";
  import { app, fail } from "../../lib/app.svelte";
  import { post } from "../../lib/api";
  import { newSession } from "../../lib/actions";
  import type { Project } from "../../lib/types";
  import Dialog from "../Dialog.svelte";

  let path = $state("");

  async function add(event: Event) {
    event.preventDefault();
    try {
      const project = await post<Project>("/api/projects", { path });
      app.dialog = null;
      app.project = project.path;
      await newSession(project.path);
    } catch (err) {
      fail(err);
    }
  }
</script>

<Dialog title="Add a project" label="a directory on this machine">
  <form onsubmit={add}>
    <input class="field mono" placeholder="~/code/project or /abs/path" bind:value={path} use:focus />
    <p class="soft">A relative path starts at {app.dir}.</p>
    <div class="actions"><button class="btn primary" disabled={!path.trim()}>Add and open</button></div>
  </form>
</Dialog>

<style>
  .actions { display: flex; justify-content: flex-end; }
  p { font-size: 12px; }
</style>
