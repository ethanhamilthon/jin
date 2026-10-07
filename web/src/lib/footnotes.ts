import type { MarkedExtension, Tokens } from "marked";
import { escape } from "./escape";

type Note = { id: string; text: string };

let defined = new Map<string, string>();
let order: string[] = [];

export function resetNotes() {
  defined = new Map();
  order = [];
}

// notesHtml lists the notes that were referenced, in order of first use.
export function notesHtml(inline: (text: string) => string): string {
  const notes: Note[] = order.filter((id) => defined.has(id)).map((id) => ({ id, text: defined.get(id)! }));
  if (!notes.length) return "";
  const items = notes.map((n) => `<li id="fn-${escape(n.id)}">${inline(n.text)} <a href="#fnref-${escape(n.id)}" class="fn-back">↩</a></li>`);
  return `<section class="footnotes"><ol>${items.join("")}</ol></section>`;
}

export const footnotes: MarkedExtension = {
  extensions: [
    {
      name: "footnoteDef",
      level: "block",
      start: (src) => src.search(/^\[\^[^\]\s]+\]:/m),
      tokenizer(src) {
        const match = /^\[\^([^\]\s]+)\]:[ \t]*([^\n]*(?:\n(?: {2,}|\t)[^\n]*)*)(?:\n|$)/.exec(src);
        if (!match) return undefined;
        const text = match[2].split("\n").map((l) => l.trim()).join(" ");
        return { type: "footnoteDef", raw: match[0], id: match[1], text };
      },
      renderer(token) {
        const { id, text } = token as Tokens.Generic;
        defined.set(id, text);
        return "";
      },
    },
    {
      name: "footnoteRef",
      level: "inline",
      start: (src) => src.indexOf("[^"),
      tokenizer(src) {
        const match = /^\[\^([^\]\s]+)\]/.exec(src);
        return match ? { type: "footnoteRef", raw: match[0], id: match[1] } : undefined;
      },
      renderer(token) {
        const { id } = token as Tokens.Generic;
        if (!order.includes(id)) order.push(id);
        const n = order.indexOf(id) + 1;
        return `<sup class="fn-ref"><a id="fnref-${escape(id)}" href="#fn-${escape(id)}">${n}</a></sup>`;
      },
    },
  ],
};
