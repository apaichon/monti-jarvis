type VoiceMsg = {
  type: string;
  data?: string;
  text?: string;
  role?: string;
  message?: string;
};

export type VoiceCallbacks = {
  onLive?: (live: boolean) => void;
  onTranscript?: (role: 'caller' | 'agent', text: string) => void;
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

  async start(
    agentId: string,
    topic: string,
    callbacks: VoiceCallbacks,
    opts?: { tenantId?: string }
  ) {
    const blocked = micAvailabilityError();
    if (blocked) throw new Error(blocked);

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
    await Promise.all([
      this.captureCtx.audioWorklet.addModule('/recorder.js'),
      this.playbackCtx.audioWorklet.addModule('/player.js')
    ]);
    this.source = this.captureCtx.createMediaStreamSource(this.micStream);
    this.recorder = new AudioWorkletNode(this.captureCtx, 'recorder-processor');
    this.player = new AudioWorkletNode(this.playbackCtx, 'player-processor');
    this.source.connect(this.recorder);
    this.player.connect(this.playbackCtx.destination);

    const scheme = location.protocol === 'https:' ? 'wss' : 'ws';
    const params = new URLSearchParams({ agent: agentId, topic: topic || 'general' });
    if (opts?.tenantId) params.set('tenant_id', opts.tenantId);
    this.ws = new WebSocket(`${scheme}://${location.host}/ws/voice?${params}`);

    let ready = false;
    await new Promise<void>((resolve, reject) => {
      if (!this.ws) return reject(new Error('WebSocket failed'));
      const timeout = window.setTimeout(
        () => reject(new Error('Voice connection timed out')),
        15000
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
        if (msg.type === 'ready') {
          ready = true;
          window.clearTimeout(timeout);
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
  }

  private handleMessage(raw: string, callbacks: VoiceCallbacks) {
    const msg: VoiceMsg = JSON.parse(raw);
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
      return;
    }
    if (msg.type === 'transcript' && msg.text) {
      const role = msg.role === 'assistant' ? 'agent' : 'caller';
      callbacks.onTranscript?.(role, msg.text);
      return;
    }
    if (msg.type === 'text' && msg.text) {
      callbacks.onTranscript?.('agent', msg.text);
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
