import { post } from "./api";
import { imageLabel } from "./composer";
import type { Image } from "./draft";

// upload saves pictures from a paste or drop and returns their labels.
export async function upload(files: File[], known: Image[]): Promise<Image[]> {
  const added: Image[] = [];
  for (const file of files.filter((f) => f.type.startsWith("image/"))) {
    const { path } = await post<{ path: string }>("/api/images", file);
    added.push({ label: imageLabel(known.length + added.length + 1), path, url: URL.createObjectURL(file) });
  }
  return added;
}
