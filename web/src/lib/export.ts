import { get } from "./api";

// exportMarkdown downloads a session as a Markdown file.
export async function exportMarkdown(id: string) {
  const { markdown } = await get<{ markdown: string }>(`/api/sessions/${id}/export`);
  const link = document.createElement("a");
  link.href = URL.createObjectURL(new Blob([markdown], { type: "text/markdown" }));
  link.download = `jin-${id.slice(0, 8)}.md`;
  link.click();
  URL.revokeObjectURL(link.href);
}
