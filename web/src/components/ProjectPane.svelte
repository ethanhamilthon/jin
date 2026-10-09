<script lang="ts">
  import { app, fail } from "../lib/app.svelte";
  import { get, post, query } from "../lib/api";
  import { shortPath } from "../lib/format";
  import type { Project } from "../lib/types";
  import PanelFrame from "./PanelFrame.svelte";
  import HookList from "./settings/HookList.svelte";

  let { index }: { index: number } = $props();
  const pane = $derived(app.panes[index]);
  const path = $derived(pane.project || app.projectPath);
  let project = $state<Project | null>(null);
  let trust = $state(0);
  let name = $state("");
  let confirm = $state(false);

  $effect(() => {
    void app.projectsRev, app.configRev;
    const dir = path;
    get<Project[]>("/api/projects").then((list) => {
      project = list.find((p) => p.path === dir) ?? null;
      name = project?.name ?? "";
    }).catch(fail);
    get<{ trust: number }>("/api/hooks" + query({ dir })).then((r) => (trust = r.trust)).catch(fail);
  });

  const rename = () => project && name.trim() !== project.name && post("/api/projects/rename", { path, name }).catch(fail);
  const setTrust = (trusted: boolean) => post("/api/hooks/trust", { dir: path, trusted }).catch(fail);

  async function archive() {
    if (!confirm) {
      confirm = true;
      return;
    }
    confirm = false;
    await post("/api/projects/archive", { path, archived: true }).catch(fail);
    app.toast("Project archived. Restore it in Settings, Archived projects");
  }
</script>

<PanelFrame {index} title="Project" subtitle={shortPath(path, app.home)}>
  <div class="set-hd"><span class="label">Project</span><span class="hint mono" title={path}>{shortPath(path, app.home)}</span></div>
  {#if !project}
    <p class="empty">This folder is not a registered project yet. Start a session in it first.</p>
  {:else}
    <div class="set-row">
      <div class="set-text"><span class="set-name">Name</span><span class="set-desc">shown in the TUI</span></div>
      <div class="set-ctl"><input class="field" bind:value={name} onblur={rename} onkeydown={(e) => e.key === "Enter" && e.currentTarget.blur()} style="width: 160px" /></div>
    </div>
    <div class="set-row">
      <div class="set-text"><span class="set-name">Hooks trust</span><span class="set-desc">project hooks run commands on your machine</span></div>
      <div class="set-ctl">
        <span class="pill" class:warn={trust === 0} class:dim={trust === 2}>{trust === 1 ? "trusted" : trust === 2 ? "not trusted" : "not decided"}</span>
        {#if trust !== 1}<button class="btn small" onclick={() => setTrust(true)}>Trust</button>{/if}
        {#if trust !== 2}<button class="btn ghost small" onclick={() => setTrust(false)}>Distrust</button>{/if}
      </div>
    </div>
    <HookList project dir={path} />
    <div class="set-hd"><span class="label">Danger</span></div>
    <div class="set-row">
      <div class="set-text"><span class="set-name">Archive project</span><span class="set-desc">Hide it from the sidebar. Sessions stay.</span></div>
      <div class="set-ctl">
        <button class="btn danger small" onclick={archive} onblur={() => (confirm = false)}>{confirm ? "Click again to archive" : "Archive"}</button>
      </div>
    </div>
  {/if}
</PanelFrame>
