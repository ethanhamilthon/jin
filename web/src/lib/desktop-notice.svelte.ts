export interface DesktopNotice {
  kind: "path" | "link";
  dir?: string;
  line?: string;
  message?: string;
}

// desktopNotice is a problem the desktop app reports about its jin command.
// Only the desktop app answers the route; in a browser it is not found.
export const desktopNotice = $state<{ value: DesktopNotice | null }>({ value: null });

export async function fetchDesktopNotice(): Promise<DesktopNotice | null> {
  try {
    const response = await fetch("/desktop/notice");
    if (response.status !== 200) return null;
    const body = (await response.json()) as DesktopNotice;
    return body.kind === "path" || body.kind === "link" ? body : null;
  } catch {
    return null;
  }
}

export async function loadDesktopNotice() {
  desktopNotice.value = await fetchDesktopNotice();
}

// closeDesktopNotice hides the banner until the next start.
export function closeDesktopNotice() {
  desktopNotice.value = null;
}
