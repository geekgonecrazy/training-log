// Audio cues for the routine runner.
// Web Audio API tones — no shipped sound files. Works offline, no asset bundle weight.
// (The plan flagged "bundled sound files" as a preference; v1 ships tones, and
// real .ogg/.mp3 files can drop into /sounds later without code changes.)

let ctx: AudioContext | null = null;
let unlocked = false;

function getCtx(): AudioContext | null {
  if (ctx) return ctx;
  try {
    const Ctor = window.AudioContext || (window as any).webkitAudioContext;
    if (!Ctor) return null;
    ctx = new Ctor();
    return ctx;
  } catch {
    return null;
  }
}

/**
 * On iOS, audio is muted until the user has interacted with the page. Call
 * this once from a tap/click handler to unlock the AudioContext.
 */
export function unlockAudio() {
  if (unlocked) return;
  const c = getCtx();
  if (!c) return;
  if (c.state === 'suspended') {
    void c.resume();
  }
  // Play a near-silent click so the OS knows audio is "in use".
  const osc = c.createOscillator();
  const gain = c.createGain();
  gain.gain.value = 0.001;
  osc.connect(gain);
  gain.connect(c.destination);
  osc.start();
  osc.stop(c.currentTime + 0.01);
  unlocked = true;
}

function tone(freq: number, durationMs: number, volume = 0.2) {
  const c = getCtx();
  if (!c) return;
  const osc = c.createOscillator();
  const gain = c.createGain();
  osc.frequency.value = freq;
  osc.type = 'sine';
  gain.gain.setValueAtTime(volume, c.currentTime);
  gain.gain.exponentialRampToValueAtTime(0.0001, c.currentTime + durationMs / 1000);
  osc.connect(gain);
  gain.connect(c.destination);
  osc.start();
  osc.stop(c.currentTime + durationMs / 1000);
}

async function chord(steps: { freq: number; durationMs: number; delayMs: number }[]) {
  for (const s of steps) {
    setTimeout(() => tone(s.freq, s.durationMs), s.delayMs);
  }
}

export const cues = {
  start: () =>
    chord([
      { freq: 659, durationMs: 120, delayMs: 0 },
      { freq: 988, durationMs: 180, delayMs: 100 }
    ]),
  end: () =>
    chord([
      { freq: 988, durationMs: 200, delayMs: 0 },
      { freq: 1318, durationMs: 350, delayMs: 180 }
    ]),
  tick: () => tone(440, 50, 0.15)
};
