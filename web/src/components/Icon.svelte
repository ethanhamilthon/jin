<script lang="ts">
  const paths: Record<string, string> = {
    plus: "M12 5v14M5 12h14",
    close: "M6 6l12 12M18 6L6 18",
    split: "M4 4h16v16H4zM12 4v16",
    gear: "M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6zM19.4 15a1.7 1.7 0 0 0 .3 1.9l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.9-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.9.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.9 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.9l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.9.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.9-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.9V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z",
    search: "M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16zM21 21l-4.3-4.3",
    stop: "M7 7h10v10H7z",
    send: "M5 12h14M13 6l6 6-6 6",
    folder: "M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z",
    tasks: "M12 3a9 9 0 1 0 9 9M12 7v5l3 2",
    chevron: "M9 6l6 6-6 6",
    image: "M4 5h16v14H4zM4 15l4-4 4 4 3-3 5 5M15 9h.01",
    command: "M9 6a3 3 0 1 0-3 3h12a3 3 0 1 0-3-3v12a3 3 0 1 0 3-3H6a3 3 0 1 0 3 3z",
    menu: "M4 6h16M4 12h16M4 18h16",
    check: "M5 12l5 5 9-10",
    trash: "M4 7h16M10 11v6M14 11v6M6 7l1 13h10l1-13M9 7V4h6v3",
    context: "M12 3v9h9M12 3a9 9 0 1 0 9 9",
    terminal: "M4 5h16v14H4zM8 10l3 2-3 2M13 14h3",
    clip: "m21.44 11.05-9.19 9.19a6 6 0 0 1-8.49-8.49l8.57-8.57A4 4 0 1 1 18 8.84l-8.59 8.57a2 2 0 0 1-2.83-2.83l8.49-8.48",
    dots: "M4 12a1 1 0 1 0 2 0 1 1 0 1 0-2 0zM11 12a1 1 0 1 0 2 0 1 1 0 1 0-2 0zM18 12a1 1 0 1 0 2 0 1 1 0 1 0-2 0z",
    edit: "M4 20h4L19 9l-4-4L4 16zM14 6l4 4",
    eye: "M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12zM12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z",
    compact: "M4 14h6v6M20 10h-6V4M14 10l7-7M3 21l7-7",
    handoff: "M4 4v6a4 4 0 0 0 4 4h12M16 10l4 4-4 4",
    rewind: "M11 19l-8-7 8-7zM21 19l-8-7 8-7z",
    undo: "M9 14L4 9l5-5M4 9h10a6 6 0 0 1 0 12h-3",
    reload: "M21 12a9 9 0 1 1-3-6.7L21 8M21 3v5h-5",
    export: "M12 3v12M7 10l5 5 5-5M5 21h14",
    play: "M7 5l12 7-12 7z",
    back: "M19 12H5M11 6l-6 6 6 6",
  };
  let { name, size = 16 }: { name: string; size?: number } = $props();
</script>

<svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
  <path d={paths[name]} />
</svg>
