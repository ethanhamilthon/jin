export const sectionNames = ["system", "compact", "handoff"] as const;
export type SectionName = (typeof sectionNames)[number];
export type Sections = Record<SectionName, string>;

// parseSections splits the system-prompt.md file at its "# system", "# compact"
// and "# handoff" lines, as the server does. Text before the first line is
// ignored; a name that appears twice keeps its last text.
export function parseSections(file: string): Sections {
  const out: Sections = { system: "", compact: "", handoff: "" };
  let current: SectionName | "" = "";
  let body: string[] = [];
  const flush = () => {
    if (current) out[current] = body.join("\n").trim();
    body = [];
  };
  for (const line of file.replaceAll("\r\n", "\n").split("\n")) {
    const name = sectionNames.find((n) => line.trimEnd() === "# " + n);
    if (name) {
      flush();
      current = name;
    } else {
      body.push(line);
    }
  }
  flush();
  return out;
}

// renderSections writes the three sections in file form.
export function renderSections(s: Sections): string {
  return sectionNames.map((n) => `# ${n}\n\n${s[n].trim()}\n`).join("\n");
}
