// Game sounds, synthesized with the Web Audio API so no audio assets are
// needed. Every sound is built from the same "soft tock" voice — a triangle
// wave with a fast attack and gentle exponential decay — so the whole game
// shares one muted, clock-like character. Browsers keep an AudioContext
// suspended until a user gesture, so callers should mount installAudioUnlock()
// once to resume it on the first interaction.

let audioContext: AudioContext | null = null;

const getAudioContext = (): AudioContext | null => {
  if (typeof window === "undefined" || !("AudioContext" in window)) {
    return null;
  }

  if (!audioContext) {
    try {
      audioContext = new AudioContext();
    } catch {
      return null;
    }
  }

  if (audioContext.state === "suspended") {
    void audioContext.resume();
  }

  return audioContext;
};

// Resumes the audio context on the first pointer or key interaction after
// mount. Returns a cleanup function for use in a useEffect.
export const installAudioUnlock = (): (() => void) => {
  const unlock = () => {
    getAudioContext();
  };

  window.addEventListener("pointerdown", unlock, { once: true });
  window.addEventListener("keydown", unlock, { once: true });

  return () => {
    window.removeEventListener("pointerdown", unlock);
    window.removeEventListener("keydown", unlock);
  };
};

type Tock = {
  frequency: number;
  atMs?: number;
  peak?: number;
  decayMs?: number;
};

// Schedules one or more soft tocks on the audio clock. atMs offsets let a
// pattern play as a small phrase without setTimeout drift.
const playTocks = (tocks: Tock[]) => {
  const ctx = getAudioContext();
  if (!ctx || ctx.state !== "running") {
    return;
  }

  for (const { frequency, atMs = 0, peak = 0.1, decayMs = 80 } of tocks) {
    const start = ctx.currentTime + atMs / 1000;

    const oscillator = ctx.createOscillator();
    oscillator.type = "triangle";
    oscillator.frequency.value = frequency;

    const gain = ctx.createGain();
    gain.gain.setValueAtTime(0.0001, start);
    gain.gain.exponentialRampToValueAtTime(peak, start + 0.012);
    gain.gain.exponentialRampToValueAtTime(0.0001, start + decayMs / 1000);

    oscillator.connect(gain);
    gain.connect(ctx.destination);
    oscillator.start(start);
    oscillator.stop(start + decayMs / 1000 + 0.02);
  }
};

// A short, muted clock tock. urgency in [0, 1] raises the pitch as the
// deadline approaches so the countdown feels increasingly pressing.
export const playTimerTick = (urgency: number) => {
  const clamped = Math.min(1, Math.max(0, urgency));
  playTocks([{ frequency: 660 + 220 * clamped, peak: 0.12, decayMs: 90 }]);
};

// A single low tock for a card landing on the table (any player).
export const playCardTock = () => {
  playTocks([{ frequency: 440, peak: 0.09, decayMs: 80 }]);
};

// Two rising tocks: your turn has started.
export const playYourTurnTocks = () => {
  playTocks([
    { frequency: 523, peak: 0.1, decayMs: 110 },
    { frequency: 784, atMs: 150, peak: 0.1, decayMs: 130 },
  ]);
};

// A low double knock: someone is demanding payment or property from you.
export const playDemandKnock = () => {
  playTocks([
    { frequency: 392, peak: 0.12, decayMs: 130 },
    { frequency: 294, atMs: 170, peak: 0.12, decayMs: 160 },
  ]);
};

// A tiny high tock for an incoming chat message.
export const playChatTock = () => {
  playTocks([{ frequency: 988, peak: 0.05, decayMs: 70 }]);
};

// A rising four-note phrase, last note left ringing: you won.
export const playWinTocks = () => {
  playTocks([
    { frequency: 523, peak: 0.1, decayMs: 120 },
    { frequency: 659, atMs: 140, peak: 0.1, decayMs: 120 },
    { frequency: 784, atMs: 280, peak: 0.1, decayMs: 120 },
    { frequency: 1047, atMs: 440, peak: 0.11, decayMs: 420 },
  ]);
};

// A falling three-note phrase: someone else won.
export const playLossTocks = () => {
  playTocks([
    { frequency: 392, peak: 0.1, decayMs: 140 },
    { frequency: 330, atMs: 170, peak: 0.1, decayMs: 140 },
    { frequency: 262, atMs: 340, peak: 0.1, decayMs: 300 },
  ]);
};
