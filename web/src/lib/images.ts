import { post } from "./api";
import { imageLabel } from "./composer";
import type { AttachedFile, Image } from "./draft";

export interface Uploaded { images: Image[]; files: AttachedFile[] }

// upload saves what a paste, a drop or the file picker brings: pictures
// get labels, other files keep their names.
export async function upload(files: File[], known: Image[]): Promise<Uploaded> {
  const out: Uploaded = { images: [], files: [] };
  for (const file of files) {
    if (file.type.startsWith("image/")) {
      const { path } = await post<{ path: string }>("/api/images", file);
      out.images.push({ label: imageLabel(known.length + out.images.length + 1), path, url: URL.createObjectURL(file) });
    } else {
      const { path, name } = await post<AttachedFile>("/api/files/upload?name=" + encodeURIComponent(file.name), file);
      out.files.push({ name, path });
    }
  }
  return out;
}
