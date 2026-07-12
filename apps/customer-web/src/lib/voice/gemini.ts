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

  /** Accumulators for streamed partial transcripts (one active turn each). */
  private callerBuf = '';
  private agentBuf = '';
  /** Prefer output transcription over modelTurn.text to avoid short dual captions. */
  private agentFromTranscript = false;
  private preferredLang: PreferredLang = '';

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
      this.playbackCtx.audioWorklet.addModule('/player.js')
    ]);
    this.source = this.captureCtx.createMediaStreamSource(this.micStream);
    this.recorder = new AudioWorkletNode(this.captureCtx, 'recorder-processor');
    this.player = new AudioWorkletNode(this.playbackCtx, 'player-processor');
    this.source.connect(this.recorder);
    this.player.connect(this.playbackCtx.destination);

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
      this.ws.send(JSON.stringify({ type: 'audio', data: floatToBase64PCM16(event.data) }));
    };
  }

  async stop() {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'end' }));
      this.ws.close();
    }
    await this.cleanup();
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
    this.callerBuf = '';
    this.agentBuf = '';
    this.agentFromTranscript = false;
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
      const samples = base64PCM16ToFloat(msg.data);
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

function floatToBase64PCM16(float32: Float32Array) {
  const bytes = new Uint8Array(float32.length * 2);
  const view = new DataView(bytes.buffer);
  for (let i = 0; i < float32.length; i++) {
    const sample = Math.max(-1, Math.min(1, float32[i]));
    view.setInt16(i * 2, sample < 0 ? sample * 0x8000 : sample * 0x7fff, true);
  }
  return bytesToBase64(bytes);
}

function base64PCM16ToFloat(base64: string) {
  const bytes = base64ToBytes(base64);
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
