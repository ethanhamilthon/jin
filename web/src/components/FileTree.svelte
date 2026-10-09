<script lang="ts">
  import { fail } from "../lib/app.svelte";
  import { get, query } from "../lib/api";
  import { size, type TreeEntry } from "../lib/files";
  import { every } from "../lib/poll";
  import Icon from "./Icon.svelte";

  let { dir, hidden, filter, selected, open }: {
    dir: string; hidden: boolean; filter: string; selected: string; open: (path: string) => void;
  } = $props();

  let children = $state<Record<string, TreeEntry[]>>({});
  let expanded = $state<Record<string, boolean>>({});
  let loaded = "";

  // load reads one folder. A refresh keeps the list as it is when nothing changed, and
  // forgets a folder that is gone.
  async function load(path: string, refresh = false) {
    try {
      const list = await get<TreeEntry[]>("/api/tree" + query({ dir, path, hidden: hidden ? "1" : "" }));
      if (!refresh || JSON.stringify(list) !== JSON.stringify(children[path])) children[path] = list;
    } catch (err) {
      if (!refresh) return fail(err);
      if (path) {
        expanded[path] = false;
        delete children[path];
      }
    }
  }

  // The folders in view are read again every second.
  $effect(() => every(() => Promise.all(Object.keys(expanded).filter((p) => expanded[p] && children[p]).map((p) => load(p, true)))));

  $effect(() => {
    const key = dir + "|" + hidden;
    if (key === loaded) return;
    loaded = key;
    const folders = Object.keys(expanded).filter((p) => expanded[p]);
    children = {};
    if (!folders.includes("")) folders.unshift("");
    expanded = { ...expanded, "": true };
    folders.forEach((p) => load(p));
  });

  async function toggle(path: string) {
    expanded[path] = !expanded[path];
    if (expanded[path] && !children[path]) await load(path);
  }

  const join = (base: string, name: string) => (base ? base + "/" + name : name);

  // rows lists the opened folders in order; with a filter, every loaded entry whose name matches.
  const rows = $derived.by(() => {
    const out: { path: string; name: string; dir: boolean; size: number; depth: number }[] = [];
    const needle = filter.trim().toLowerCase();
    const walk = (base: string, depth: number) => {
      for (const entry of children[base] ?? []) {
        const path = join(base, entry.name);
        if (!needle || entry.name.toLowerCase().includes(needle)) out.push({ path, ...entry, depth: needle ? 0 : depth });
        if (entry.dir && (needle || expanded[path])) walk(path, depth + 1);
      }
    };
    walk("", 0);
    return out;
  });
</script>

<div class="tree">
  {#each rows as row (row.path)}
    <button class="node" class:dir={row.dir} class:sel={row.path === selected} style:padding-left="{12 + row.depth * 16}px" onclick={() => (row.dir ? toggle(row.path) : open(row.path))} title={row.path}>
      <span class="ic">{#if row.dir}<Icon name="chevron" size={10} />{/if}</span>
      <span class="name">{row.name}</span>
      {#if filter.trim() && row.path !== row.name}<span class="where">{row.path.slice(0, -row.name.length - 1)}</span>{/if}
      {#if !row.dir}<span class="meta">{size(row.size)}</span>{/if}
    </button>
  {:else}<p class="empty">{filter.trim() ? "Nothing matches in the opened folders" : "Empty folder"}</p>{/each}
</div>

<style>
  .tree { font: 13px var(--mono); }
  .node {
    display: flex; align-items: center; gap: 6px; width: 100%; height: 28px; padding-right: 12px; text-align: left;
    background: none; border: 0; color: var(--text-dim); cursor: pointer;
  }
  .node:hover { background: var(--raised); }
  .node.dir { color: var(--text-strong); }
  .node.sel { background: var(--raised); box-shadow: inset 2px 0 0 var(--accent); color: var(--text-strong); }
  .ic { width: 12px; flex: none; color: var(--text-muted); display: grid; place-items: center; }
  .name { flex: 0 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .where { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 11px; color: var(--text-muted); }
  .meta { margin-left: auto; font-size: 11px; color: var(--text-muted); flex: none; }
</style>
