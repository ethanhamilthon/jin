<script lang="ts">
  import { app, fail, type SessionView } from "../lib/app.svelte";
  import { act } from "../lib/actions";
  import type { Item } from "../lib/composer";
  import { Completer } from "../lib/completer.svelte";
  import { submit, type AttachedFile, type Image } from "../lib/draft";
  import { upload } from "../lib/images";
  import { syncDraft } from "../lib/draft-sync.svelte";
  import { toggleTools, toolsShown } from "../lib/fold";
  import Completion from "./Completion.svelte";
  import Icon from "./Icon.svelte";
  import Thumbnails from "./Thumbnails.svelte";
  import AttachButton from "./AttachButton.svelte";
  import FileChips from "./FileChips.svelte";
  import MoreMenu from "./MoreMenu.svelte";
  import { mobile } from "../lib/mobile.svelte";

  let { view, focused }: { view: SessionView; focused: boolean } = $props();
  const info = $derived(view.state);
  let text = $state("");
  let images = $state<Image[]>([]);
  let files = $state<AttachedFile[]>([]);
  let area: HTMLTextAreaElement;
  const menu = new Completer();

  syncDraft(() => info, () => text, (value) => {
    text = value;
    images = [];
    files = [];
  });
  $effect(() => {
    if (focused && area) area.focus();
  });

  let inserted = 0;
  $effect(() => {
    const next = app.insertion;
    if (!next || next.n === inserted || next.session !== info.id) return;
    inserted = next.n;
    text = text && !/\s$/.test(text) ? `${text} ${next.text} ` : `${text}${next.text} `;
  });

  const shell = $derived(text.startsWith("$"));
  const disabled = $derived(!info.ready || !!info.read_only);
  const shown = $derived(toolsShown(app.config.fold));

  const refresh = () => menu.refresh(text, area.selectionStart, info.path);

  function choose(item: Item) {
    const next = menu.apply(text, area.selectionStart, item);
    text = next.text;
    requestAnimationFrame(() => {
      area.setSelectionRange(next.cursor, next.cursor);
      refresh();
    });
  }

  function focusEnd(event: MouseEvent) {
    if (disabled || (event.target as HTMLElement).closest("button, textarea, a, input")) return;
    area.focus();
    area.setSelectionRange(text.length, text.length);
  }

  async function send() {
    const draft = text;
    text = "";
    if (!(await submit(info.id, draft, images, files))) text = draft;
    else [images, files] = [[], []];
  }

  function onKey(event: KeyboardEvent) {
    const enter = event.key === "Enter" && !event.shiftKey && !event.altKey && !event.isComposing;
    if (menu.shown && (event.key === "ArrowDown" || event.key === "ArrowUp")) {
      menu.move(event.key === "ArrowDown" ? 1 : -1);
    } else if (menu.shown && (event.key === "Tab" || enter)) {
      choose(menu.list[menu.selected]);
    } else if (event.key === "Escape" && menu.open) {
      menu.close();
    } else if (enter && !mobile.on) {
      send();
    } else {
      if (event.key === "c" && event.ctrlKey && info.busy && area.selectionStart === area.selectionEnd) act(info.id, "stop");
      return;
    }
    event.preventDefault();
  }

  async function attach(picked: File[]) {
    if (!picked.length) return;
    try {
      const added = await upload(picked, images);
      images = [...images, ...added.images];
      files = [...files, ...added.files];
    } catch (err) {
      fail(err);
    }
  }
</script>

{#if images.length}<Thumbnails {images} remove={(label) => (images = images.filter((i) => i.label !== label))} />{/if}
{#if files.length}<FileChips {files} remove={(path) => (files = files.filter((f) => f.path !== path))} />{/if}
<!-- svelte-ignore a11y_click_events_have_key_events, a11y_no_static_element_interactions -->
<div class="composer" class:shell class:busy={info.busy} onclick={focusEnd}>
  {#if menu.shown}<Completion items={menu.list} selected={menu.selected} {choose} />{/if}
  <div class="input">
  <textarea
    bind:this={area} bind:value={text} rows="2" {disabled}
    placeholder={info.read_only ? "Read-only: another jin process uses this session" : !info.ready ? "Starting: running prompt commands…" : "Message jin · # prompts · @ files · $ shell"}
    oninput={refresh} onkeydown={onKey} onclick={refresh} onblur={() => menu.close()}
    onpaste={(e) => attach([...(e.clipboardData?.files ?? [])])}
    ondrop={(e) => { e.preventDefault(); attach([...(e.dataTransfer?.files ?? [])]); }}
  ></textarea>
  </div>
  <div class="bar">
    {#if shell}<span class="label">[ shell ]</span>{/if}
    <div class="tools">
      <AttachButton pick={attach} disabled={disabled} />
      <button class="btn ghost small" onclick={() => app.open("context", info.id)} title="Context" aria-label="Context"><Icon name="context" size={15} /></button>
      <button class="btn ghost small" class:on={shown} onclick={toggleTools} title="Chat details (Ctrl+O)" aria-label="Chat details" aria-pressed={shown}><Icon name="eye" size={15} /></button>
      <MoreMenu id={info.id} />
    </div>
    {#if info.queued}<span class="queued mono">{info.queued} queued</span>{/if}
    <span class="spacer"></span>
    <button class="btn ghost small model" onclick={() => app.open("model", info.id)} title="Model and effort">
      {info.model || "no model"}{#if info.effort}<span class="muted"> · {info.effort}</span>{/if}
    </button>
    {#if info.busy}
      <button class="btn danger small" onclick={() => act(info.id, "stop")} title="Stop (Ctrl+C)" aria-label="Stop"><Icon name="stop" size={13} /></button>
    {/if}
    <button class="btn primary small" onclick={send} disabled={disabled || (!text.trim() && !images.length && !files.length)} title={mobile.on ? "Send" : "Send (Enter)"}><Icon name="send" size={13} /></button>
  </div>
</div>

<style>
  .composer {
    position: relative; display: flex; flex-direction: column; gap: 6px; padding: 10px 8px 8px 12px;
    background: var(--card); border: 0; box-shadow: none; outline: none; border-radius: var(--radius);
  }
  .input { position: relative; }
  textarea {
    display: block; width: 100%; box-sizing: border-box; resize: none; border: 0; outline: none; background: transparent; min-height: calc(2lh + 4px); max-height: 40vh;
    field-sizing: content; line-height: 1.55; padding: 2px 0;
  }
  @media (max-width: 700px) { .composer { border-radius: 0; } }
  .shell textarea { font-family: var(--mono); font-size: 13px; }
  .spacer { flex: 1; }
  .queued { font-size: 11px; color: var(--accent); }
  .bar { display: flex; align-items: center; gap: 8px; }
  .tools { display: flex; align-items: center; gap: 4px; }
  .on { color: var(--accent); }
  .model { font-family: var(--mono); font-size: 12px; max-width: 260px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
</style>
