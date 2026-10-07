import { Marked } from "marked";
import DOMPurify from "dompurify";
import hljs from "highlight.js/lib/common";

const marked = new Marked({
  gfm: true,
  breaks: false,
  renderer: {
    code({ text, lang }) {
      const language = lang && hljs.getLanguage(lang) ? lang : "";
      const html = language ? hljs.highlight(text, { language }).value : escape(text);
      const label = language ? `<span class="code-lang">${escape(language)}</span>` : "";
      return `<pre class="code">${label}<code class="hljs">${html}</code></pre>`;
    },
  },
});

export function escape(text: string): string {
  return text.replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]!);
}

// render turns model Markdown into sanitized HTML; links open in a new tab.
export function render(text: string): string {
  const html = marked.parse(text, { async: false }) as string;
  const clean = DOMPurify.sanitize(html, { ADD_ATTR: ["target"] });
  return clean.replace(/<a /g, '<a target="_blank" rel="noreferrer noopener" ');
}
