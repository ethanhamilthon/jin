<script lang="ts">
  import { app, fail, type SessionView } from "../lib/app.svelte";
  import { act } from "../lib/actions";
  import type { Item } from "../lib/composer";
  import { Completer } from "../lib/completer.svelte";
  import { submit, type Image } from "../lib/draft";
  import { upload } from "../lib/images";
  import { syncDraft } from "../lib/draft-sync.svelte";
  import Completion from "./Completion.svelte";
  import Suggestion from "./Suggestion.svelte";
  import Icon from "./Icon.svelte";

  let { view, focused }: { view: SessionView; focused: boolean } = $props();
  const info = $derived(view.state);
  let text = $state("");
  let images = $state<Image[]>([]);
  let area: HTMLTextAreaElement;
  const menu = new Completer();

  syncDraft(() => info, () => text, (value) => {
    text = value;
    images = [];
  });
  $effect(() => {
    if (focused && area) area.focus();
  });

  const shell = $derived(text.startsWith("$"));
  const disabled = $derived(!info.ready || !!info.read_only);

  const refresh = () => menu.refresh(text, area.selectionStart, info.path);

  function choose(item: Item) {
    const next = menu.apply(text, area.selectionStart, item);
    text = next.text;
    requestAnimationFrame(() => {
      area.setSelectionRange(next.cursor, next.cursor);
      refresh();
    });
  }

  async function send() {
    const draft = text;
    text = "";
    if (!(await submit(info.id, draft, images))) text = draft;
    else images = [];
  }

  function onKey(event: KeyboardEvent) {
    const enter = event.key === "Enter" && !event.shiftKey && !event.altKey && !event.isComposing;
    if (menu.shown && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
      menu.move(event.key === "ArrowDown" ? 1 : -1);
    } else if (menu.shown && enter && menu.open?.kind === "/") {
      text = menu.list[menu.selected].insert;
      menu.close();
      send();
    } else if (menu.shown && (event.key === "Tab" || enter)) {
      choose(menu.list[menu.selected]);
    } else if (event.key === "Escape" && menu.open) {
      menu.close();
    } else if (enter) {
      send();
    } else {
      if (event.key === "c" && event.ctrlKey && info.busy && area.selectionStart === area.selectionEnd) act(info.id, "stop");
      return;
    }
    event.preventDefault();
  }

  async function attach(files: File[]) {
    if (!files.length) return;
    try {
      const added = await upload(files, images);
      images = [...images, ...added];
      text += (text && !text.endsWith(" ") ? " " : "") + added.map((i) => i.label).join(" ") + " ";
    } catch (err) {
      fail(err);
    }
  }
</script>

{#if info.suggestion && !text && !info.busy}
  <Suggestion text={info.suggestion} send={() => submit(info.id, info.suggestion ?? "", [])} edit={() => (text = info.suggestion ?? "")} />
{/if}
<div class="composer" class:shell class:busy={info.busy}>
  {#if menu.shown}<Completion items={menu.list} selected={menu.selected} {choose} />{/if}
  {#if shell}<span class="label mode">[ shell ]</span>{/if}
  <textarea
    bind:this={area} bind:value={text} rows="1" {disabled}
    placeholder={info.read_only ? "Read-only: another jin process uses this session" : !info.ready ? "Starting: running prompt commands…" : "Message jin · / commands · # prompts · @ files · $ shell"}
    oninput={refresh} onkeydown={onKey} onclick={refresh} onblur={() => menu.close()}
    onpaste={(e) => attach([...(e.clipboardData?.files ?? [])])}
    ondrop={(e) => { e.preventDefault(); attach([...(e.dataTransfer?.files ?? [])]); }}
  ></textarea>
  {#if info.busy}
    <button class="btn danger small" onclick={() => act(info.id, "stop")} title="Stop (Ctrl+C)"><Icon name="stop" size={13} />Stop</button>
  {/if}
  <button class="btn primary small" onclick={send} disabled={disabled || !text.trim()} title="Send (Enter)"><Icon name="send" size={13} /></button>
</div>

<style>
  .composer {
    position: relative; display: flex; align-items: flex-end; gap: 8px; padding: 8px 8px 8px 12px;
    background: var(--card); border: 1px solid var(--raised); border-radius: var(--radius);
  }
  .composer:focus-within { border-color: var(--accent); box-shadow: var(--glow); }
  .composer.shell:focus-within { border-color: var(--text-dim); box-shadow: none; }
  textarea {
    flex: 1; resize: none; border: 0; outline: none; background: transparent; min-height: 24px; max-height: 40vh;
    field-sizing: content; line-height: 1.55; padding: 2px 0;
  }
  .shell textarea { font-family: var(--mono); font-size: 13px; }
  .mode { align-self: center; white-space: nowrap; }
</style>
