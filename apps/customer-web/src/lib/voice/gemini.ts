import {
  mergeTranscriptChunk,
  preferMainLanguage,
  type PreferredLang
} from './transcript';

type VoiceMsg = {
  type: string;
  data?: string;
  text?: string;
  role?: string;
  message?: string;
};

export type TranscriptMeta = {
  /** True when Gemini finished the model turn (or input side should settle). */
  final?: boolean;
};

export type VoiceCallbacks = {
  onLive?: (live: boolean) => void;
  /** Progress while connecting (mic, Gemini setup) — show loading UI. */
  onStatus?: (message: string) => void;
  /** Live caption updates — `text` is the full turn so far (not a short fragment). */
  onTranscript?: (role: 'caller' | 'agent', text: string, meta?: TranscriptMeta) => void;
  onError?: (message: string) => void;
};

export type VoiceRecording = {
  name: string;
  content_type: 'audio/wav';
  data_base64: string;
};

const MAX_RECORDING_BYTES = 32 * 1024 * 1024;
const MIX_SAMPLE_RATE = 24000;

type TimedPCMChunk = {
  pcm: Uint8Array;
  sampleRate: number;
  offsetSamples: number;
  source: 'caller' | 'agent';
};

/** Why getUserMedia is missing (common for embed iframes on http://custom-host). */
export function micAvailabilityError(): string | null {
  if (typeof window === 'undefined') return 'Microphone unavailable';
  if (!window.isSecureContext) {
    const host = location.hostname || 'this host';
    return (
      `Microphone needs a secure context (HTTPS or localhost). ` +
      `This page is ${location.protocol}//${host} — browsers block mic/audio here. ` +
      `Use https://… or http://localhost:… for the Monti embed host (not a custom http hostname).`
    );
  }
  if (!navigator.mediaDevices || typeof navigator.mediaDevices.getUserMedia !== 'function') {
    return (
      'Microphone API unavailable (navigator.mediaDevices missing). ' +
      'Allow microphone for this site, or open Monti on HTTPS/localhost. ' +
      'Cross-origin embeds need iframe allow="microphone *".'
    );
  }
  return null;
}

export class GeminiVoice {
  private ws: WebSocket | null = null;
  private micStream: MediaStream | null = null;
  private captureCtx: AudioContext | null = null;
  private playbackCtx: AudioContext | null = null;
  private source: MediaStreamAudioSourceNode | null = null;
  private recorder: AudioWorkletNode | null = null;
  private player: AudioWorkletNode | null = null;
  private playbackRecorder: AudioWorkletNode | null = null;

  /** Accumulators for streamed partial transcripts (one active turn each). */
  private callerBuf = '';
  private agentBuf = '';
  /** Prefer output transcription over modelTurn.text to avoid short dual captions. */
  private agentFromTranscript = false;
  private preferredLang: PreferredLang = '';
  private recordingStartedAt = 0;
  private recordingChunks: TimedPCMChunk[] = [];
  private recordingBytes = 0;
  private recordingTrackStarts: Record<'caller' | 'agent', number | null> = { caller: null, agent: null };
  private recordingTrackCursors: Record<'caller' | 'agent', number> = { caller: 0, agent: 0 };

  async start(
    agentId: string,
    topic: string,
    callbacks: VoiceCallbacks,
    opts?: { tenantId?: string; preferredLang?: PreferredLang; lang?: PreferredLang | 'auto' }
  ) {
    const blocked = micAvailabilityError();
    if (blocked) throw new Error(blocked);

    this.preferredLang = opts?.preferredLang || (opts?.lang === 'th' || opts?.lang === 'en' ? opts.lang : '');
    this.callerBuf = '';
    this.agentBuf = '';
    this.agentFromTranscript = false;
    this.recordingStartedAt = performance.now();
    this.recordingChunks = [];
    this.recordingBytes = 0;
    this.recordingTrackStarts = { caller: null, agent: null };
    this.recordingTrackCursors = { caller: 0, agent: 0 };

    callbacks.onStatus?.('Requesting microphone…');
    // Create and resume audio while the Start call click still carries user activation.
    // Waiting for getUserMedia first can leave the playback context suspended in embeds.
    this.captureCtx = new AudioContext({ sampleRate: 16000 });
    this.playbackCtx = new AudioContext({ sampleRate: 24000 });
    await Promise.all([this.captureCtx.resume(), this.playbackCtx.resume()]);

    try {
      this.micStream = await navigator.mediaDevices.getUserMedia({
        audio: {
          channelCount: 1,
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      });
    } catch (err) {
      const name = err instanceof DOMException ? err.name : '';
      const msg = err instanceof Error ? err.message : 'Microphone permission denied';
      if (name === 'NotAllowedError' || name === 'PermissionDeniedError') {
        await this.cleanup();
        throw new Error(
          'Microphone permission denied. Click the lock icon in the address bar and allow mic for this site, then try Start call again.'
        );
      }
      if (name === 'NotFoundError' || name === 'DevicesNotFoundError') {
        await this.cleanup();
        throw new Error('No microphone found. Plug in a mic or check system sound settings.');
      }
      await this.cleanup();
      throw new Error(msg);
    }
    callbacks.onStatus?.('Loading audio…');
    await Promise.all([
      this.captureCtx.audioWorklet.addModule('/recorder.js'),
      this.playbackCtx.audioWorklet.addModule('/player.js'),
      this.playbackCtx.audioWorklet.addModule('/recorder.js')
    ]);
    this.source = this.captureCtx.createMediaStreamSource(this.micStream);
    this.recorder = new AudioWorkletNode(this.captureCtx, 'recorder-processor');
    this.player = new AudioWorkletNode(this.playbackCtx, 'player-processor');
    this.playbackRecorder = new AudioWorkletNode(this.playbackCtx, 'recorder-processor');
    this.source.connect(this.recorder);
    this.player.connect(this.playbackRecorder);
    this.playbackRecorder.connect(this.playbackCtx.destination);

    callbacks.onStatus?.('Connecting to agent (may take a few seconds)…');
    const scheme = location.protocol === 'https:' ? 'wss' : 'ws';
    const params = new URLSearchParams({ agent: agentId, topic: topic || 'general' });
    if (opts?.tenantId) params.set('tenant_id', opts.tenantId);
    const lang = opts?.lang || opts?.preferredLang;
    if (lang) params.set('lang', lang);
    this.ws = new WebSocket(`${scheme}://${location.host}/ws/voice?${params}`);

    let ready = false;
    await new Promise<void>((resolve, reject) => {
      if (!this.ws) return reject(new Error('WebSocket failed'));
      // Gemini Live setup + RAG can exceed 10s; keep UI status updates alive.
      const timeout = window.setTimeout(
        () => reject(new Error('Voice connection timed out (AI setup took too long). Try again.')),
        45000
      );
      const fail = (message: string) => {
        window.clearTimeout(timeout);
        reject(new Error(message));
      };
      this.ws.addEventListener('error', () => fail('Voice connection failed'));
      this.ws.addEventListener('close', () => {
        if (!ready) fail('Voice connection closed');
      });
      this.ws.addEventListener('message', (event) => {
        const msg: VoiceMsg = JSON.parse(event.data as string);
        if (msg.type === 'status' && msg.message) {
          callbacks.onStatus?.(msg.message);
          return;
        }
        if (msg.type === 'ready') {
          ready = true;
          window.clearTimeout(timeout);
          callbacks.onStatus?.(msg.message || 'Connected — agent is greeting you…');
          callbacks.onLive?.(true);
          resolve();
          return;
        }
        if (msg.type === 'error') {
          fail(msg.message || 'Voice error');
        }
      });
    });

    this.ws.addEventListener('message', (event) => this.handleMessage(event.data, callbacks));
    this.ws.addEventListener('close', () => callbacks.onLive?.(false));
    this.recorder.port.onmessage = (event) => {
      if (!(event.data instanceof Float32Array)) return;
      if (!ready || !this.ws || this.ws.readyState !== WebSocket.OPEN) return;
      const pcm = floatToPCM16Bytes(event.data);
      this.appendRecordingChunk(pcm, 16000, 'caller');
      this.ws.send(JSON.stringify({ type: 'audio', data: bytesToBase64(pcm) }));
    };
    this.playbackRecorder.port.onmessage = (event) => {
      if (!(event.data instanceof Float32Array)) return;
      this.appendRecordingChunk(floatToPCM16Bytes(event.data), 24000, 'agent', 0);
    };
  }

  async stop(): Promise<VoiceRecording[]> {
    const recordings = this.recordings();
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'end' }));
      this.ws.close();
    }
    await this.cleanup();
    return recordings;
  }

  sendText(text: string) {
    const content = text.trim();
    if (!content || this.ws?.readyState !== WebSocket.OPEN) return false;
    this.ws.send(JSON.stringify({ type: 'text', text: content }));
    return true;
  }

  recordings(): VoiceRecording[] {
    if (this.recordingBytes === 0 || this.recordingChunks.length === 0) return [];
    return [
      {
        name: 'call-recording',
        content_type: 'audio/wav',
        data_base64: bytesToBase64(stereoPCM16ToWav(renderStereoRecording(this.recordingChunks), MIX_SAMPLE_RATE))
      }
    ];
  }

  private async cleanup() {
    this.micStream?.getTracks().forEach((t) => t.stop());
    await this.captureCtx?.close().catch(() => {});
    await this.playbackCtx?.close().catch(() => {});
    this.ws = null;
    this.micStream = null;
    this.captureCtx = null;
    this.playbackCtx = null;
    this.source = null;
    this.recorder = null;
    this.player = null;
    this.playbackRecorder = null;
    this.callerBuf = '';
    this.agentBuf = '';
    this.agentFromTranscript = false;
    this.recordingStartedAt = 0;
    this.recordingChunks = [];
    this.recordingBytes = 0;
    this.recordingTrackStarts = { caller: null, agent: null };
    this.recordingTrackCursors = { caller: 0, agent: 0 };
  }

  private appendRecordingChunk(
    pcm: Uint8Array,
    sampleRate: number,
    source: 'caller' | 'agent',
    firstChunkOffsetMs?: number
  ) {
    if (pcm.byteLength === 0 || this.recordingBytes >= MAX_RECORDING_BYTES) return;
    const available = MAX_RECORDING_BYTES - this.recordingBytes;
    const next = pcm.byteLength > available ? pcm.slice(0, available) : pcm;
    if (this.recordingTrackStarts[source] === null) {
      const offsetMs = firstChunkOffsetMs ?? Math.max(0, performance.now() - this.recordingStartedAt);
      this.recordingTrackStarts[source] = Math.round((offsetMs / 1000) * MIX_SAMPLE_RATE);
    }
    const offsetSamples = this.recordingTrackStarts[source]! + this.recordingTrackCursors[source];
    this.recordingChunks.push({
      pcm: next,
      sampleRate,
      offsetSamples,
      source
    });
    this.recordingTrackCursors[source] += Math.ceil((next.byteLength / 2) * (MIX_SAMPLE_RATE / sampleRate));
    this.recordingBytes += next.byteLength;
  }

  private emitTranscript(
    callbacks: VoiceCallbacks,
    role: 'caller' | 'agent',
    text: string,
    meta?: TranscriptMeta
  ) {
    const cleaned = preferMainLanguage(text, this.preferredLang);
    if (!cleaned) return;
    callbacks.onTranscript?.(role, cleaned, meta);
  }

  private handleMessage(raw: string, callbacks: VoiceCallbacks) {
    const msg: VoiceMsg = JSON.parse(raw);
    if (msg.type === 'status' && msg.message) {
      callbacks.onStatus?.(msg.message);
      return;
    }
    if (msg.type === 'audio' && msg.data && this.player && this.playbackCtx) {
      const pcm = base64ToBytes(msg.data);
      const samples = pcm16BytesToFloat(pcm);
      if (this.playbackCtx.state === 'suspended') {
        void this.playbackCtx
          .resume()
          .then(() => this.player?.port.postMessage(samples))
          .catch(() => callbacks.onError?.('Audio playback is blocked. Click Start call again.'));
      } else {
        this.player.port.postMessage(samples);
      }
      return;
    }
    if (msg.type === 'interrupted' && this.player) {
      this.player.port.postMessage('flush');
      // Keep partial agent caption; start fresh on next agent speech.
      if (this.agentBuf) {
        this.emitTranscript(callbacks, 'agent', this.agentBuf, { final: true });
      }
      this.agentBuf = '';
      this.agentFromTranscript = false;
      return;
    }
    if (msg.type === 'transcript' && msg.text) {
      const role = msg.role === 'assistant' ? 'agent' : 'caller';
      if (role === 'caller') {
        // New user speech after agent finished → clear agent buffer for next reply.
        if (this.agentBuf && this.callerBuf === '') {
          // already finalized on turn_complete usually
        }
        // If agent was speaking and user starts, finalize agent caption first.
        if (this.agentBuf && !this.callerBuf) {
          this.emitTranscript(callbacks, 'agent', this.agentBuf, { final: true });
          this.agentBuf = '';
          this.agentFromTranscript = false;
        }
        this.callerBuf = mergeTranscriptChunk(this.callerBuf, msg.text);
        this.emitTranscript(callbacks, 'caller', this.callerBuf);
      } else {
        // Agent speaking → finalize caller turn once.
        if (this.callerBuf) {
          this.emitTranscript(callbacks, 'caller', this.callerBuf, { final: true });
          this.callerBuf = '';
        }
        this.agentFromTranscript = true;
        this.agentBuf = mergeTranscriptChunk(this.agentBuf, msg.text);
        this.emitTranscript(callbacks, 'agent', this.agentBuf);
      }
      return;
    }
    // modelTurn text: only use if we never got outputAudioTranscription for this turn.
    if (msg.type === 'text' && msg.text) {
      if (this.agentFromTranscript) return;
      if (this.callerBuf) {
        this.emitTranscript(callbacks, 'caller', this.callerBuf, { final: true });
        this.callerBuf = '';
      }
      this.agentBuf = mergeTranscriptChunk(this.agentBuf, msg.text);
      this.emitTranscript(callbacks, 'agent', this.agentBuf);
      return;
    }
    if (msg.type === 'turn_complete') {
      if (this.callerBuf) {
        this.emitTranscript(callbacks, 'caller', this.callerBuf, { final: true });
        this.callerBuf = '';
      }
      if (this.agentBuf) {
        this.emitTranscript(callbacks, 'agent', this.agentBuf, { final: true });
        this.agentBuf = '';
      }
      this.agentFromTranscript = false;
      return;
    }
    if (msg.type === 'error') {
      callbacks.onError?.(msg.message || 'Voice error');
    }
  }
}

function floatToPCM16Bytes(float32: Float32Array) {
  const bytes = new Uint8Array(float32.length * 2);
  const view = new DataView(bytes.buffer);
  for (let i = 0; i < float32.length; i++) {
    const sample = Math.max(-1, Math.min(1, float32[i]));
    view.setInt16(i * 2, sample < 0 ? sample * 0x8000 : sample * 0x7fff, true);
  }
  return bytes;
}

function pcm16BytesToFloat(bytes: Uint8Array) {
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  const out = new Float32Array(bytes.byteLength / 2);
  for (let i = 0; i < out.length; i++) out[i] = view.getInt16(i * 2, true) / 0x8000;
  return out;
}

function bytesToBase64(bytes: Uint8Array) {
  let binary = '';
  const chunk = 0x8000;
  for (let i = 0; i < bytes.length; i += chunk) {
    binary += String.fromCharCode(...bytes.subarray(i, i + chunk));
  }
  return btoa(binary);
}

function base64ToBytes(base64: string) {
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
  return bytes;
}

function renderStereoRecording(chunks: TimedPCMChunk[]) {
  let totalSamples = 0;
  const prepared = chunks.map((chunk) => {
    const samples = pcm16BytesToFloat(chunk.pcm);
    const mixed = chunk.sampleRate === MIX_SAMPLE_RATE ? samples : resampleLinear(samples, chunk.sampleRate, MIX_SAMPLE_RATE);
    const offset = chunk.offsetSamples;
    totalSamples = Math.max(totalSamples, offset + mixed.length);
    return { samples: mixed, offset, source: chunk.source };
  });
  const left = new Float32Array(totalSamples);
  const right = new Float32Array(totalSamples);
  for (const item of prepared) {
    const target = item.source === 'caller' ? left : right;
    for (let i = 0; i < item.samples.length; i++) {
      const idx = item.offset + i;
      target[idx] = item.samples[i];
    }
  }
  return interleaveStereoPCM16(left, right);
}

function interleaveStereoPCM16(left: Float32Array, right: Float32Array) {
  const length = Math.max(left.length, right.length);
  const bytes = new Uint8Array(length * 4);
  const view = new DataView(bytes.buffer);
  for (let i = 0; i < length; i++) {
    view.setInt16(i * 4, floatToInt16(left[i] || 0), true);
    view.setInt16(i * 4 + 2, floatToInt16(right[i] || 0), true);
  }
  return bytes;
}

function floatToInt16(value: number) {
  const sample = Math.max(-1, Math.min(1, value));
  return sample < 0 ? sample * 0x8000 : sample * 0x7fff;
}

function resampleLinear(input: Float32Array, fromRate: number, toRate: number) {
  if (fromRate === toRate || input.length === 0) return input;
  const ratio = fromRate / toRate;
  const out = new Float32Array(Math.ceil(input.length / ratio));
  for (let i = 0; i < out.length; i++) {
    const pos = i * ratio;
    const left = Math.floor(pos);
    const right = Math.min(left + 1, input.length - 1);
    const frac = pos - left;
    out[i] = input[left] + (input[right] - input[left]) * frac;
  }
  return out;
}

function stereoPCM16ToWav(pcm: Uint8Array, sampleRate: number) {
  const header = new ArrayBuffer(44);
  const view = new DataView(header);
  const channels = 2;
  const bytesPerSample = 2;
  const blockAlign = channels * bytesPerSample;
  writeAscii(view, 0, 'RIFF');
  view.setUint32(4, 36 + pcm.byteLength, true);
  writeAscii(view, 8, 'WAVE');
  writeAscii(view, 12, 'fmt ');
  view.setUint32(16, 16, true);
  view.setUint16(20, 1, true);
  view.setUint16(22, channels, true);
  view.setUint32(24, sampleRate, true);
  view.setUint32(28, sampleRate * blockAlign, true);
  view.setUint16(32, blockAlign, true);
  view.setUint16(34, 16, true);
  writeAscii(view, 36, 'data');
  view.setUint32(40, pcm.byteLength, true);
  const wav = new Uint8Array(44 + pcm.byteLength);
  wav.set(new Uint8Array(header), 0);
  wav.set(pcm, 44);
  return wav;
}

function writeAscii(view: DataView, offset: number, value: string) {
  for (let i = 0; i < value.length; i++) view.setUint8(offset + i, value.charCodeAt(i));
}
