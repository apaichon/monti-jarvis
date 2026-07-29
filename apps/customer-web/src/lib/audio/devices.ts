/** Browser mic/speaker selection for customer voice (Sprint 56). */

export type AudioDevice = {
  deviceId: string;
  label: string;
  kind: 'audioinput' | 'audiooutput';
};

const INPUT_KEY = 'monti_jarvis:audio_input_id';
const OUTPUT_KEY = 'monti_jarvis:audio_output_id';

export function getStoredInputDeviceId(): string {
  if (typeof window === 'undefined') return '';
  try {
    return window.localStorage.getItem(INPUT_KEY)?.trim() || '';
  } catch {
    return '';
  }
}

export function getStoredOutputDeviceId(): string {
  if (typeof window === 'undefined') return '';
  try {
    return window.localStorage.getItem(OUTPUT_KEY)?.trim() || '';
  } catch {
    return '';
  }
}

export function storeInputDeviceId(id: string) {
  if (typeof window === 'undefined') return;
  try {
    if (id) window.localStorage.setItem(INPUT_KEY, id);
    else window.localStorage.removeItem(INPUT_KEY);
  } catch {
    /* ignore */
  }
}

export function storeOutputDeviceId(id: string) {
  if (typeof window === 'undefined') return;
  try {
    if (id) window.localStorage.setItem(OUTPUT_KEY, id);
    else window.localStorage.removeItem(OUTPUT_KEY);
  } catch {
    /* ignore */
  }
}

export function supportsSpeakerSelection(): boolean {
  if (typeof HTMLMediaElement === 'undefined') return false;
  return typeof (HTMLMediaElement.prototype as { setSinkId?: unknown }).setSinkId === 'function';
}

/** Prompt for mic once so enumerateDevices returns readable labels. */
export async function ensureAudioPermission(): Promise<void> {
  if (!navigator.mediaDevices?.getUserMedia) {
    throw new Error('Audio devices are not available in this browser.');
  }
  const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
  for (const track of stream.getTracks()) track.stop();
}

export async function listAudioDevices(): Promise<{
  inputs: AudioDevice[];
  outputs: AudioDevice[];
}> {
  if (!navigator.mediaDevices?.enumerateDevices) {
    return { inputs: [], outputs: [] };
  }
  const all = await navigator.mediaDevices.enumerateDevices();
  const inputs: AudioDevice[] = [];
  const outputs: AudioDevice[] = [];
  let inIdx = 0;
  let outIdx = 0;
  for (const d of all) {
    if (d.kind === 'audioinput') {
      inIdx += 1;
      inputs.push({
        deviceId: d.deviceId,
        label: d.label?.trim() || `Microphone ${inIdx}`,
        kind: 'audioinput'
      });
    } else if (d.kind === 'audiooutput') {
      outIdx += 1;
      outputs.push({
        deviceId: d.deviceId,
        label: d.label?.trim() || `Speaker ${outIdx}`,
        kind: 'audiooutput'
      });
    }
  }
  return { inputs, outputs };
}

export function friendlyMediaError(err: unknown): string {
  const name = err instanceof DOMException ? err.name : '';
  if (name === 'NotAllowedError' || name === 'PermissionDeniedError') {
    return 'Microphone access is needed for voice calls. You can still chat.';
  }
  if (name === 'NotFoundError' || name === 'DevicesNotFoundError') {
    return 'No microphone found. Check your device connections.';
  }
  if (err instanceof Error && err.message) return err.message;
  return 'Could not access audio devices.';
}

export function buildAudioConstraints(deviceId?: string): MediaTrackConstraints {
  const base: MediaTrackConstraints = {
    channelCount: 1,
    echoCancellation: true,
    noiseSuppression: true,
    autoGainControl: true
  };
  const id = deviceId?.trim();
  if (id) {
    return { ...base, deviceId: { ideal: id } };
  }
  return base;
}

/** Apply output device when browser supports setSinkId. */
export async function applyOutputDevice(
  el: { setSinkId?: (id: string) => Promise<void> } | null | undefined,
  deviceId?: string
): Promise<{ applied: boolean; note?: string }> {
  const id = deviceId?.trim();
  if (!id) return { applied: false };
  if (!el || typeof el.setSinkId !== 'function') {
    return {
      applied: false,
      note: 'This browser uses the system default speaker.'
    };
  }
  try {
    await el.setSinkId(id);
    return { applied: true };
  } catch {
    return {
      applied: false,
      note: 'Could not switch speaker; using the system default.'
    };
  }
}
