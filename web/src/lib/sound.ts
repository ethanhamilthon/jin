import type { Sound } from "./types";

export type SoundEvent = "turn" | "ask" | "error";
export interface Tone { id: string; name: string }

let active: HTMLAudioElement | undefined;

export function shouldRing(sound: Sound, focused: boolean) {
  return sound.enabled && sound.volume > 0 && (!sound.only_blur || !focused);
}

export function ring(sound: Sound, focused: boolean, _event: SoundEvent = "turn") {
  if (shouldRing(sound, focused)) play(sound);
}

export function preview(sound: Sound) {
  if (sound.volume > 0) play(sound);
}

function play(sound: Sound) {
  active?.pause();
  const audio = new Audio(`/api/sounds/${encodeURIComponent(sound.tone || "bell")}.wav`);
  active = audio;
  audio.volume = Math.min(1, Math.max(0, sound.volume / 100));
  void audio.play().catch(() => {});
}
