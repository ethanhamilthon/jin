export const defaultAccent = "#52a8ff";

export const accents = [
  { name: "Blue", value: "#52a8ff" },
  { name: "Lime", value: "#c5ff4a" },
  { name: "Teal", value: "#0ac7b4" },
  { name: "Violet", value: "#bf7af0" },
  { name: "Amber", value: "#ffb224" },
  { name: "Pink", value: "#f75f8f" },
];

// applyAccent sets the one chromatic color everything else derives from.
export function applyAccent(color: string) {
  document.documentElement.style.setProperty("--accent", color || defaultAccent);
}
