import { Marked } from "marked";
import DOMPurify from "dompurify";
import hljs from "highlight.js/lib/common";
import { escape } from "./escape";
import { alertQuote, highlight, localImage } from "./markdown_ext";
import { footnotes, notesHtml, resetNotes } from "./footnotes";

export { escape };

let imageDir = "";

const marked = new Marked(highlight, footnotes, {
  gfm: true,
  breaks: false,
  renderer: {
    code({ text, lang }) {
      const language = lang && hljs.getLanguage(lang) ? lang : "";
      const html = language ? hljs.highlight(text, { language }).value : escape(text);
      const label = language ? `<span class="code-lang">${escape(language)}</span>` : "";
      const copy = `<button type="button" class="code-copy" title="Copy">copy</button>`;
      return `<pre class="code">${label}${copy}<code class="hljs">${html}</code></pre>`;
    },
    blockquote({ tokens }) {
      const html = this.parser.parse(tokens);
      return alertQuote(html) ?? `<blockquote>\n${html}</blockquote>\n`;
    },
    image({ href, title, text }) {
      const caption = title ? ` title="${escape(title)}"` : "";
      return `<img src="${escape(localImage(href, imageDir))}" alt="${escape(text)}"${caption} loading="lazy">`;
    },
  },
});

// render turns model Markdown into sanitized HTML; links open in a new tab.
// dir is the project that relative image paths belong to.
export function render(text: string, dir = ""): string {
  imageDir = dir;
  resetNotes();
  const body = marked.parse(text, { async: false }) as string;
  const notes = notesHtml((line) => marked.parseInline(line, { async: false }) as string);
  const clean = DOMPurify.sanitize(body + notes, { ADD_ATTR: ["target"] });
  return clean.replace(/<a (?![^>]*href="#)/g, '<a target="_blank" rel="noreferrer noopener" ');
}
