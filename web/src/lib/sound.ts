import type { Sound } from "./types";

let context: AudioContext | undefined;

// ring plays the notification tone when the settings allow it.
export function ring(sound: Sound, focused: boolean) {
  if (!sound.enabled || (sound.only_blur && focused)) return;
  context ??= new AudioContext();
  const now = context.currentTime;
  const gain = context.createGain();
  gain.connect(context.destination);
  gain.gain.setValueAtTime(0.0001, now);
  gain.gain.exponentialRampToValueAtTime((sound.volume / 100) * 0.3, now + 0.02);
  gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.45);
  for (const [freq, start] of [[880, 0], [1320, 0.12]]) {
    const osc = context.createOscillator();
    osc.type = "sine";
    osc.frequency.value = freq;
    osc.connect(gain);
    osc.start(now + start);
    osc.stop(now + 0.45);
  }
}
