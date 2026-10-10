import type { Sound } from "./types";

export type SoundEvent = "turn" | "ask" | "task" | "error";
export interface EventSound { on: boolean; tone: ToneName }
export type SoundChoices = Record<SoundEvent, EventSound>;

interface Note { freq: number; start: number; length: number }
interface Tone { wave: OscillatorType; level: number; notes: Note[] }

const note = (freq: number, start: number, length: number): Note => ({ freq, start, length });

export const tones = {
  bell: { wave: "sine", level: 0.3, notes: [note(880, 0, 0.45), note(1320, 0.12, 0.33)] },
  pluck: { wave: "triangle", level: 0.3, notes: [note(660, 0, 0.25), note(990, 0.12, 0.3)] },
  arpeggio: { wave: "square", level: 0.12, notes: [note(523, 0, 0.2), note(659, 0.1, 0.2), note(784, 0.2, 0.3)] },
  buzz: { wave: "sawtooth", level: 0.12, notes: [note(220, 0, 0.3), note(165, 0.2, 0.35)] },
} satisfies Record<string, Tone>;

export type ToneName = keyof typeof tones;
const toneNames = Object.keys(tones) as ToneName[];

const storageKey = "jin.sounds";
// The turn keeps the bell jin always rang; task and error stay silent until enabled.
const defaults: SoundChoices = {
  turn: { on: true, tone: "bell" },
  ask: { on: true, tone: "pluck" },
  task: { on: false, tone: "arpeggio" },
  error: { on: false, tone: "buzz" },
};
const eventNames = Object.keys(defaults) as SoundEvent[];

type Saved = Partial<Record<SoundEvent, Partial<EventSound>>>;

function readSaved(): Saved {
  try {
    const value: unknown = JSON.parse(localStorage.getItem(storageKey) ?? "{}");
    return typeof value === "object" && value !== null ? (value as Saved) : {};
  } catch {
    return {};
  }
}

// loadChoices reads the sound of each event; a missing or invalid value gets its default.
export function loadChoices(): SoundChoices {
  const saved = readSaved();
  const choices = {} as SoundChoices;
  for (const name of eventNames) {
    const pick = saved[name];
    choices[name] = {
      on: typeof pick?.on === "boolean" ? pick.on : defaults[name].on,
      tone: toneNames.find((tone) => tone === pick?.tone) ?? defaults[name].tone,
    };
  }
  return choices;
}

export function saveChoices(choices: SoundChoices) {
  try {
    localStorage.setItem(storageKey, JSON.stringify(choices));
  } catch {
    // Storage is blocked or full: the choice lasts until the page closes.
  }
}

let context: AudioContext | undefined;

// ring plays the sound of an event when the settings allow it.
export function ring(sound: Sound, focused: boolean, event: SoundEvent = "turn") {
  const choice = loadChoices()[event];
  if (!sound.enabled || !choice.on || (sound.only_blur && focused)) return;
  if (matchMedia("(prefers-reduced-motion: reduce)").matches) return;
  play(tones[choice.tone], (sound.volume / 100) * tones[choice.tone].level);
}

function play(tone: Tone, level: number) {
  context ??= new AudioContext();
  const now = context.currentTime;
  for (const { freq, start, length } of tone.notes) {
    const osc = context.createOscillator();
    const gain = context.createGain();
    const from = now + start;
    const to = from + length;
    osc.type = tone.wave;
    osc.frequency.value = freq;
    gain.gain.setValueAtTime(0.0001, from);
    gain.gain.exponentialRampToValueAtTime(level, from + 0.02);
    gain.gain.exponentialRampToValueAtTime(0.0001, to);
    osc.connect(gain);
    gain.connect(context.destination);
    osc.start(from);
    osc.stop(to);
  }
}
