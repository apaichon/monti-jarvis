/** Per-avatar mouth seam (0–1 from top of portrait). Tuned for bundled photos. */
export type LipSyncPreset = {
  mouthLine: number;
  jawDrop: number;
  jawScale: number;
  mouthWidth: number;
};

const defaults: LipSyncPreset = {
  mouthLine: 0.64,
  jawDrop: 0.09,
  jawScale: 0.28,
  mouthWidth: 0.24
};

export const lipSyncPresets: Record<string, LipSyncPreset> = {
  ava: { mouthLine: 0.66, jawDrop: 0.1, jawScale: 0.32, mouthWidth: 0.26 },
  max: { mouthLine: 0.65, jawDrop: 0.095, jawScale: 0.3, mouthWidth: 0.27 },
  luna: { mouthLine: 0.67, jawDrop: 0.1, jawScale: 0.31, mouthWidth: 0.25 },
  neo: { mouthLine: 0.56, jawDrop: 0.055, jawScale: 0.18, mouthWidth: 0.2 }
};

export function lipPresetFor(agentId: string, robot?: boolean): LipSyncPreset {
  const preset = lipSyncPresets[agentId];
  if (preset) return preset;
  if (robot) return lipSyncPresets.neo;
  return defaults;
}

export function mouthOpenFromLevel(level: number, preset: LipSyncPreset): number {
  const raw = Math.min(1, Math.max(0, level * 16));
  return Math.pow(raw, 0.82);
}