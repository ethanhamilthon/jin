import type { MarkedExtension, Tokens } from "marked";
import { escape } from "./escape";

const alerts: Record<string, string> = {
  NOTE: "Note", TIP: "Tip", IMPORTANT: "Important", WARNING: "Warning", CAUTION: "Caution",
};

// alertQuote turns "> [!NOTE]" blockquotes into labeled callouts.
export function alertQuote(html: string): string | null {
  const match = /^<p>\[!(NOTE|TIP|IMPORTANT|WARNING|CAUTION)\]\s*(?:<br>\s*)?/i.exec(html.trim());
  if (!match) return null;
  const kind = match[1].toUpperCase();
  const body = html.trim().slice(match[0].length);
  const rest = body.startsWith("</p>") ? body.slice(4) : "<p>" + body;
  return `<div class="alert alert-${kind.toLowerCase()}"><div class="alert-title">${alerts[kind]}</div>${rest}</div>`;
}

export const highlight: MarkedExtension = {
  extensions: [{
    name: "mark",
    level: "inline",
    start: (src) => src.indexOf("=="),
    tokenizer(src) {
      const match = /^==(?=\S)([^=]+?)(?<=\S)==/.exec(src);
      return match ? { type: "mark", raw: match[0], text: match[1] } : undefined;
    },
    renderer: (token) => `<mark>${escape((token as Tokens.Generic).text)}</mark>`,
  }],
};

export const localImage = (href: string, dir: string): string =>
  /^(https?:|data:|blob:|\/api\/)/i.test(href)
    ? href
    : "/api/local-image?" + new URLSearchParams({ path: href.replace(/^file:\/\//, ""), dir });
