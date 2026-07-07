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

export class GeminiVoice {
  private ws: WebSocket | null = null;
  private micStream: MediaStream | null = null;
  private captureCtx: AudioContext | null = null;
  private playbackCtx: AudioContext | null = null;
  private source: MediaStreamAudioSourceNode | null = null;
  private recorder: AudioWorkletNode | null = null;
  private player: AudioWorkletNode | null = null;

  async start(agentId: string, callbacks: VoiceCallbacks) {
    this.micStream = await navigator.mediaDevices.getUserMedia({
      audio: { channelCount: 1, echoCancellation: true, noiseSuppression: true, autoGainControl: true }
    });
    this.captureCtx = new AudioContext({ sampleRate: 16000 });
    this.playbackCtx = new AudioContext({ sampleRate: 24000 });
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
    this.ws = new WebSocket(`${scheme}://${location.host}/ws/voice?agent=${encodeURIComponent(agentId)}`);

    await new Promise<void>((resolve, reject) => {
      if (!this.ws) return reject(new Error('WebSocket failed'));
      this.ws.addEventListener('open', () => {
        callbacks.onLive?.(true);
        resolve();
      });
      this.ws.addEventListener('error', () => reject(new Error('Voice connection failed')));
    });

    this.ws.addEventListener('message', (event) => this.handleMessage(event.data, callbacks));
    this.ws.addEventListener('close', () => callbacks.onLive?.(false));
    this.recorder.port.onmessage = (event) => {
      if (!(event.data instanceof Float32Array)) return;
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return;
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
    if (msg.type === 'audio' && msg.data && this.player) {
      this.player.port.postMessage(base64PCM16ToFloat(msg.data));
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