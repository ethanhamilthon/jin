import { post } from "./api";
import { imageLabel } from "./composer";
import type { AttachedFile, Image } from "./draft";

export interface Uploaded { images: Image[]; files: AttachedFile[] }

// fileBody wraps a file for the upload routes. The bytes go as base64 in JSON:
// the desktop app can lose a raw binary body.
export async function fileBody(file: Blob): Promise<{ data: string }> {
  const bytes = new Uint8Array(await file.arrayBuffer());
  let text = "";
  for (let i = 0; i < bytes.length; i += 0x8000) text += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
  return { data: btoa(text) };
}

// upload saves what a paste, a drop or the file picker brings: pictures
// get labels, other files keep their names.
export async function upload(files: File[], known: Image[]): Promise<Uploaded> {
  const out: Uploaded = { images: [], files: [] };
  for (const file of files) {
    const body = await fileBody(file);
    if (file.type.startsWith("image/")) {
      const { path } = await post<{ path: string }>("/api/images", body);
      out.images.push({ label: imageLabel(known.length + out.images.length + 1), path, url: URL.createObjectURL(file) });
    } else {
      const { path, name } = await post<AttachedFile>("/api/files/upload?name=" + encodeURIComponent(file.name), body);
      out.files.push({ name, path });
    }
  }
  return out;
}
