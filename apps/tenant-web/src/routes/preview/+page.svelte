<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import {
    hasRegistrationSession,
    getAccessToken,
    setAccessToken,
    getStoredUser
  } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import { getEmbedConfig, type EmbedConfig } from '$lib/api/embed';
  import Portrait from '$lib/components/Portrait.svelte';
  import {
    listScenarios,
    listPreviewAgents,
    previewChat,
    previewVoiceURL,
    type PreviewScenario,
    type PreviewAgent,
    type PreviewChatMessage
  } from '$lib/api/preview';
  import { listTiers, type CustomerTier } from '$lib/api/tiers';
  import {
    micAvailabilityError,
    requestMicrophone,
    localhostPreviewHref,
    isInsecureCustomHost
  } from '$lib/voice/mic';

  type UiMsg = {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    initial: string;
    missingKm?: boolean;
  };

  let agents = $state<PreviewAgent[]>([]);
  let selected = $state<PreviewAgent | null>(null);
  let scenarios = $state<PreviewScenario[]>([]);
  let topic = $state('general');
  /** Session language: auto | en | th */
  let lang = $state<'auto' | 'en' | 'th'>('auto');
  let tiers = $state<CustomerTier[]>([]);
  let tierId = $state('');
  let input = $state('');
  let messages = $state<UiMsg[]>([]);
  let sessionId = $state('');
  let history = $state<PreviewChatMessage[]>([]);
  let loading = $state(true);
  let sending = $state(false);
  let voiceLive = $state(false);
  let voiceBusy = $state(false);
  let voiceState = $state('Select an agent, then start a voice call or type a message.');
  let connectStep = $state(''); // loading progress while starting call
  let embed = $state<EmbedConfig | null>(null);
  let chatEl: HTMLElement | undefined = $state();
  let micBlock = $state<string | null>(null);
  let tone = $state('');
  let timer = $state('00:00:00');
  let startedAt = 0;
  let timerId: ReturnType<typeof setInterval> | undefined;

  let voiceWs: WebSocket | null = null;
  let micStream: MediaStream | null = null;
  let captureCtx: AudioContext | null = null;
  let playbackCtx: AudioContext | null = null;
  let recorder: AudioWorkletNode | null = null;
  let player: AudioWorkletNode | null = null;
  let agentCaption = $state('');

  const workspace = $derived(getStoredUser()?.display_name || 'Tenant preview');

  function agentInitial(name?: string) {
    return (name || 'A').slice(0, 1).toUpperCase();
  }

  function formatTimer(seconds: number) {
    return new Date(seconds * 1000).toISOString().slice(11, 19);
  }

  function startTimer() {
    startedAt = Date.now();
    clearInterval(timerId);
    timerId = setInterval(() => {
      timer = formatTimer(Math.floor((Date.now() - startedAt) / 1000));
    }, 1000);
  }

  function stopTimer() {
    clearInterval(timerId);
    timer = '00:00:00';
  }

  function greetingFor(a: PreviewAgent): string {
    const baseGreet =
      a.greeting ||
      `Hi, I'm ${a.name}. Ask me anything to preview your knowledge and package limits.`;
    if (lang === 'th') {
      return `${baseGreet}\n\n(สวัสดีค่ะ/ครับ ฉันคือ ${a.name} พร้อมช่วยเหลือคุณ)`;
    }
    return baseGreet;
  }

  function selectAgent(a: PreviewAgent) {
    selected = a;
    const greet = greetingFor(a);
    messages = [
      {
        id: `welcome-${a.id}`,
        role: 'assistant',
        content: greet,
        initial: agentInitial(a.name)
      }
    ];
    history = [];
    sessionId = '';
    agentCaption = '';
    voiceState = `Ready to call ${a.name} (package minutes apply · agent greets first).`;
  }

  onMount(async () => {
    const sp = $page.url.searchParams;
    const bootToken = sp.get('access_token') || sp.get('token');
    if (bootToken) {
      setAccessToken(bootToken);
      const agentQ = sp.get('agent');
      const topicQ = sp.get('topic');
      const langQ = sp.get('lang');
      if (topicQ) topic = topicQ;
      if (langQ === 'en' || langQ === 'th' || langQ === 'auto') lang = langQ;
      const autoVoice = sp.get('auto_voice') === '1' || sp.get('auto_voice') === 'true';
      const clean = new URL(location.href);
      clean.searchParams.delete('access_token');
      clean.searchParams.delete('token');
      clean.searchParams.delete('auto_voice');
      window.history.replaceState({}, '', clean.pathname + (clean.search || '') + clean.hash);

      micBlock = micAvailabilityError();
      await load(agentQ || undefined);
      if (autoVoice && !micBlock) {
        setTimeout(() => void startVoice(), 300);
      }
      return;
    }

    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/preview`)}`);
      return;
    }
    micBlock = micAvailabilityError();
    await load();
  });

  function openVoiceOnLocalhost() {
    const token = getAccessToken();
    if (!token) {
      feedback.error('Not signed in — log in again, then retry voice');
      return;
    }
    const href = localhostPreviewHref({
      access_token: token,
      agent: selected?.id || 'ava',
      topic,
      lang,
      auto_voice: '1'
    });
    const w = window.open(href, 'monti-preview-voice', 'noopener,noreferrer');
    if (!w) {
      window.location.href = href;
      return;
    }
    feedback.success('Opened Preview on localhost for voice');
  }

  function onStartVoiceClick() {
    if (micBlock || isInsecureCustomHost()) {
      openVoiceOnLocalhost();
      return;
    }
    void startVoice();
  }

  async function load(preferAgent?: string) {
    loading = true;
    try {
      const [ag, sc, emb, tr] = await Promise.all([
        listPreviewAgents(),
        listScenarios(),
        getEmbedConfig().catch(() => null),
        listTiers().catch(() => ({ tiers: [] as CustomerTier[] }))
      ]);
      tiers = (tr.tiers || []).filter((t) => t.active);
      agents = (ag.agents || []).filter((a) => a.id);
      if (!agents.length) {
        agents = [
          {
            id: 'ava',
            name: 'Ava',
            role: 'Reception',
            color: '#00b7ff',
            image: '/images/ava.jpg'
          },
          {
            id: 'max',
            name: 'Max',
            role: 'Billing',
            color: '#7c5cff',
            image: '/images/max.jpg'
          },
          {
            id: 'luna',
            name: 'Luna',
            role: 'Technical',
            color: '#3dd68c',
            image: '/images/luna.jpg'
          },
          {
            id: 'neo',
            name: 'Neo',
            role: 'Triage',
            color: '#ffb86c',
            image: '/images/neo.jpg'
          }
        ];
      }
      const pick =
        agents.find((a) => a.id === preferAgent) ||
        agents.find((a) => a.id === emb?.default_agent_id) ||
        agents[0];
      selectAgent(pick);
      scenarios = sc.scenarios || [];
      embed = emb;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load preview');
    } finally {
      loading = false;
    }
  }

  async function scrollChat() {
    await tick();
    if (chatEl) chatEl.scrollTop = chatEl.scrollHeight;
  }

  function applyScenario(s: PreviewScenario) {
    topic = s.topic || topic;
    input = s.question;
  }

  async function send() {
    const text = input.trim();
    if (!text || sending || !selected || voiceLive) return;
    sending = true;
    input = '';
    messages = [
      ...messages,
      { id: `${Date.now()}-u`, role: 'user', content: text, initial: 'Y' }
    ];
    void scrollChat();
    try {
      const res = await previewChat({
        agent_id: selected.id,
        topic,
        message: text,
        session_id: sessionId || undefined,
        history,
        lang,
        tier_id: tierId || undefined
      });
      sessionId = res.session_id;
      history = [
        ...history,
        { role: 'user' as const, content: text },
        { role: 'assistant' as const, content: res.reply }
      ].slice(-16);
      messages = [
        ...messages,
        {
          id: `${Date.now()}-a`,
          role: 'assistant',
          content: res.reply,
          initial: agentInitial(selected.name),
          missingKm: res.missing_km
        }
      ];
      void scrollChat();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Preview chat failed');
      input = text;
    } finally {
      sending = false;
    }
  }

  function openEmbed() {
    if (!embed?.enabled || !embed.embed_key) {
      feedback.error('Embed is disabled — enable it under Embed settings');
      return;
    }
    const origin = encodeURIComponent(location.origin);
    const url = `${location.origin}/embed?key=${encodeURIComponent(embed.embed_key)}&parent_origin=${origin}`;
    window.open(url, '_blank', 'noopener,noreferrer');
  }

  async function startVoice() {
    if (voiceBusy || voiceLive || !selected) return;
    voiceBusy = true;
    connectStep = 'Requesting microphone…';
    voiceState = 'Connecting…';
    agentCaption = '';
    // Show text greeting immediately while audio path connects.
    if (selected.greeting || selected.name) {
      messages = [
        ...messages,
        {
          id: `${Date.now()}-greet`,
          role: 'assistant',
          content: greetingFor(selected),
          initial: agentInitial(selected.name)
        }
      ];
      void scrollChat();
    }
    try {
      micBlock = micAvailabilityError();
      if (micBlock) throw new Error(micBlock);

      micStream = await requestMicrophone();
      micBlock = null;
      connectStep = 'Loading audio…';

      const AC =
        window.AudioContext ||
        (window as unknown as { webkitAudioContext?: typeof AudioContext }).webkitAudioContext;
      if (!AC) throw new Error('Web Audio API is not supported in this browser.');
      captureCtx = new AC({ sampleRate: 16000 });
      playbackCtx = new AC({ sampleRate: 24000 });
      await Promise.all([captureCtx.resume(), playbackCtx.resume()]);

      const workletBase = location.origin;
      await Promise.all([
        captureCtx.audioWorklet.addModule(`${workletBase}/recorder.js`),
        playbackCtx.audioWorklet.addModule(`${workletBase}/player.js`)
      ]);
      const source = captureCtx.createMediaStreamSource(micStream);
      recorder = new AudioWorkletNode(captureCtx, 'recorder-processor');
      player = new AudioWorkletNode(playbackCtx, 'player-processor');
      source.connect(recorder);
      player.connect(playbackCtx.destination);

      connectStep = 'Connecting to AI (may take several seconds)…';
      const url = previewVoiceURL(selected.id, topic, lang, tierId || undefined);
      voiceWs = new WebSocket(url);
      await new Promise<void>((resolve, reject) => {
        // Gemini Live dial + RAG can take >10s
        const t = window.setTimeout(
          () => reject(new Error('Voice connection timed out. Try again.')),
          45000
        );
        voiceWs!.onerror = () => {
          window.clearTimeout(t);
          reject(new Error('Voice connection failed'));
        };
        voiceWs!.onclose = () => {
          window.clearTimeout(t);
          reject(new Error('Voice connection closed before ready'));
        };
        voiceWs!.onmessage = (ev) => {
          const msg = JSON.parse(ev.data as string);
          if (msg.type === 'status' && msg.message) {
            connectStep = msg.message;
            voiceState = msg.message;
            return;
          }
          if (msg.type === 'ready') {
            window.clearTimeout(t);
            connectStep = msg.message || 'Agent is greeting you…';
            voiceState = connectStep;
            resolve();
          }
          if (msg.type === 'error') {
            window.clearTimeout(t);
            reject(new Error(msg.message || 'Voice error'));
          }
        };
      });

      voiceWs.onmessage = (ev) => {
        const msg = JSON.parse(ev.data as string);
        if (msg.type === 'status' && msg.message) {
          voiceState = msg.message;
          return;
        }
        if (msg.type === 'audio' && msg.data && player && playbackCtx) {
          const samples = base64PCM16ToFloat(msg.data);
          if (playbackCtx.state === 'suspended') {
            void playbackCtx.resume().then(() => player?.port.postMessage(samples));
          } else {
            player.port.postMessage(samples);
          }
        }
        if (msg.type === 'interrupted' && player) {
          player.port.postMessage('flush');
        }
        if (msg.type === 'transcript' && msg.text) {
          if (msg.role === 'assistant') {
            agentCaption = (agentCaption + msg.text).slice(-400);
          }
        }
        if (msg.type === 'text' && msg.text) {
          agentCaption = (agentCaption + ' ' + msg.text).trim().slice(-400);
        }
        if (msg.type === 'turn_complete' && agentCaption) {
          messages = [
            ...messages,
            {
              id: `${Date.now()}-v`,
              role: 'assistant',
              content: agentCaption,
              initial: agentInitial(selected?.name)
            }
          ];
          agentCaption = '';
          void scrollChat();
        }
      };
      voiceWs.onclose = () => {
        voiceLive = false;
        connectStep = '';
        stopTimer();
        voiceState = 'Voice ended (package minutes charged)';
      };

      recorder.port.onmessage = (event) => {
        if (!(event.data instanceof Float32Array)) return;
        if (!voiceWs || voiceWs.readyState !== WebSocket.OPEN) return;
        voiceWs.send(JSON.stringify({ type: 'audio', data: floatToBase64PCM16(event.data) }));
      };

      voiceLive = true;
      startTimer();
      voiceState = `On call with ${selected.name} — listen for the greeting…`;
      connectStep = '';
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Voice failed';
      voiceState = msg;
      connectStep = '';
      feedback.error(msg);
      await stopVoice();
    } finally {
      voiceBusy = false;
    }
  }

  async function stopVoice() {
    try {
      if (voiceWs?.readyState === WebSocket.OPEN) {
        voiceWs.send(JSON.stringify({ type: 'end' }));
        voiceWs.close();
      }
    } catch {
      /* ignore */
    }
    voiceWs = null;
    micStream?.getTracks().forEach((t) => t.stop());
    micStream = null;
    await captureCtx?.close().catch(() => {});
    await playbackCtx?.close().catch(() => {});
    captureCtx = null;
    playbackCtx = null;
    recorder = null;
    player = null;
    voiceLive = false;
    stopTimer();
  }

  function floatToBase64PCM16(float32: Float32Array) {
    const bytes = new Uint8Array(float32.length * 2);
    const view = new DataView(bytes.buffer);
    for (let i = 0; i < float32.length; i++) {
      const sample = Math.max(-1, Math.min(1, float32[i]));
      view.setInt16(i * 2, sample < 0 ? sample * 0x8000 : sample * 0x7fff, true);
    }
    let binary = '';
    for (let i = 0; i < bytes.length; i++) binary += String.fromCharCode(bytes[i]);
    return btoa(binary);
  }

  function base64PCM16ToFloat(base64: string) {
    const binary = atob(base64);
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
    const view = new DataView(bytes.buffer);
    const out = new Float32Array(bytes.byteLength / 2);
    for (let i = 0; i < out.length; i++) out[i] = view.getInt16(i * 2, true) / 0x8000;
    return out;
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void send();
    }
  }
</script>

<div class="page-wrap">
  <div class="banner" role="status">
    <strong>Preview mode</strong> — same experience as customer embed.
    <span class="charge">Uses package rate limits &amp; call minutes</span>
    (logged as <code>preview</code>).
  </div>

  {#if micBlock}
    <div class="banner warn" role="alert">
      <strong>Voice needs localhost</strong> — browsers block mic on <code>http://*.local</code>.
      <button class="btn" type="button" onclick={openVoiceOnLocalhost} style="margin-left:8px">
        Start voice on localhost
      </button>
    </div>
  {/if}

  {#if loading}
    <p class="dim">Loading…</p>
  {:else}
    <div class="embed-panel">
      <header class="embed-hdr">
        <div class="brand">
          <img class="brand-mark" src="{base}/images/monti-logo.png" width="28" height="28" alt="" />
          <div>
            <strong>{workspace}</strong>
            <span class="sub">Preview · text &amp; voice</span>
          </div>
        </div>
        {#if agents.length > 1}
          <select
            class="agent-sel"
            value={selected?.id || ''}
            disabled={voiceLive}
            onchange={(e) => {
              const id = (e.currentTarget as HTMLSelectElement).value;
              const a = agents.find((x) => x.id === id);
              if (a) selectAgent(a);
            }}
          >
            {#each agents as a}
              <option value={a.id}>{a.name}</option>
            {/each}
          </select>
        {/if}
      </header>

      {#if selected}
        <section class="orb">
          <div class="halo" style="--assistant-color:{selected.color || '#00b7ff'}">
            <Portrait agent={selected} speaking={voiceLive} {tone} />
          </div>
          <div class="agent-meta">
            <h2>{selected.name}</h2>
            <p>
              {selected.role || 'AI agent'}{selected.trait ? ` · ${selected.trait}` : ''}
            </p>
          </div>
        </section>
      {/if}

      <section class="voice-card">
        <div class="voice-row">
          <div class="status-pill">{timer}</div>
          <label class="topic-lab">
            Topic
            <select bind:value={topic} disabled={voiceLive || voiceBusy}>
              <option value="general">General</option>
              <option value="billing">Billing</option>
              <option value="technical">Technical</option>
            </select>
          </label>
          <label class="topic-lab">
            Language
            <select bind:value={lang} disabled={voiceLive || voiceBusy}>
              <option value="auto">Auto</option>
              <option value="en">English</option>
              <option value="th">ไทย</option>
            </select>
          </label>
          {#if tiers.length}
            <label class="topic-lab">
              Tier
              <select bind:value={tierId} disabled={voiceLive || voiceBusy}>
                <option value="">None</option>
                {#each tiers as t}
                  <option value={t.id}>{t.name}</option>
                {/each}
              </select>
            </label>
          {/if}
          {#if !voiceLive}
            <button
              class="voice-button"
              type="button"
              disabled={voiceBusy || !selected}
              onclick={onStartVoiceClick}
            >
              {voiceBusy ? 'Connecting…' : micBlock ? 'Voice on localhost' : 'Start call'}
            </button>
          {:else}
            <button class="voice-button live" type="button" onclick={stopVoice}>End call</button>
          {/if}
        </div>
        {#if voiceBusy || connectStep}
          <div class="loading-bar" aria-live="polite">
            <span class="spinner" aria-hidden="true"></span>
            <span>{connectStep || voiceState || 'Connecting…'}</span>
          </div>
        {:else}
          <div class="voice-state">{voiceState}</div>
        {/if}
        {#if agentCaption}
          <div class="caption">Live: {agentCaption}</div>
        {/if}
      </section>

      <div class="chips">
        {#each scenarios as s}
          <button class="chip" type="button" onclick={() => applyScenario(s)}>{s.label}</button>
        {/each}
        {#if embed?.enabled}
          <button class="chip" type="button" onclick={openEmbed}>Open live embed</button>
        {:else}
          <a class="chip linkish" href="{base}/embed">Enable embed →</a>
        {/if}
      </div>

      <div class="msgs" bind:this={chatEl} aria-live="polite">
        {#each messages as m (m.id)}
          <div class="msg" class:user={m.role === 'user'}>
            <div class="dot">{m.initial}</div>
            <div class="bubble" class:user={m.role === 'user'}>
              {m.content}
              {#if m.missingKm}
                <div class="km-miss">Missing KM — add documents under Knowledge</div>
              {/if}
            </div>
          </div>
        {/each}
      </div>

      <footer class="composer">
        <input
          type="text"
          placeholder={voiceLive ? 'On a voice call…' : 'Type a message…'}
          bind:value={input}
          onkeydown={onKey}
          disabled={sending || !selected || voiceLive}
        />
        <button
          type="button"
          class="send"
          onclick={send}
          disabled={sending || voiceLive || !input.trim() || !selected}
        >
          {sending ? '…' : 'Send'}
        </button>
      </footer>
    </div>
  {/if}
</div>

<style>
  .page-wrap {
    max-width: 440px;
    margin: 0 auto;
    padding: 16px 16px 32px;
  }
  .banner {
    margin: 0 0 12px;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid rgb(0 183 255 / 35%);
    background: rgb(0 183 255 / 10%);
    font-size: 12px;
    line-height: 1.4;
  }
  .banner .charge {
    color: #ffb86c;
    font-weight: 600;
  }
  .banner.warn {
    border-color: rgb(255 184 108 / 50%);
    background: rgb(255 140 0 / 12%);
    color: #ffd9a8;
  }
  .dim {
    color: var(--muted);
  }
  .embed-panel {
    display: flex;
    flex-direction: column;
    min-height: 640px;
    max-height: calc(100vh - 120px);
    border-radius: 16px;
    overflow: hidden;
    border: 1px solid rgb(0 183 255 / 22%);
    background: #05101f;
    color: #f7fbff;
    box-shadow: 0 16px 48px rgb(0 0 0 / 45%);
  }
  .embed-hdr {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 12px 14px;
    border-bottom: 1px solid rgb(0 183 255 / 18%);
    background: rgb(8 20 36 / 95%);
    flex-shrink: 0;
  }
  .brand {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .brand strong {
    display: block;
    font-size: 13px;
  }
  .sub {
    font-size: 11px;
    color: #8fa5bf;
  }
  .brand-mark {
    border-radius: 8px;
  }
  .agent-sel {
    max-width: 140px;
    padding: 6px 8px;
    border-radius: 8px;
    border: 1px solid rgb(0 183 255 / 25%);
    background: #0a1528;
    color: inherit;
    font-size: 12px;
  }
  .orb {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 20px 16px 8px;
    gap: 10px;
    flex-shrink: 0;
  }
  .halo {
    --assistant-color: #00b7ff;
    padding: 4px;
    border-radius: 50%;
    background: radial-gradient(
      circle,
      color-mix(in srgb, var(--assistant-color) 25%, transparent),
      transparent 70%
    );
  }
  .agent-meta {
    text-align: center;
  }
  .agent-meta h2 {
    margin: 0;
    font-size: 18px;
  }
  .agent-meta p {
    margin: 4px 0 0;
    font-size: 12px;
    color: #8fa5bf;
  }
  .voice-card {
    padding: 8px 14px 10px;
    flex-shrink: 0;
  }
  .voice-row {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .status-pill {
    font-variant-numeric: tabular-nums;
    font-size: 12px;
    padding: 6px 10px;
    border-radius: 999px;
    border: 1px solid rgb(0 183 255 / 25%);
    color: #8fa5bf;
  }
  .topic-lab {
    display: flex;
    align-items: center;
    gap: 6px;
    font-size: 11px;
    color: #8fa5bf;
  }
  .topic-lab select {
    padding: 4px 6px;
    border-radius: 6px;
    border: 1px solid rgb(0 183 255 / 25%);
    background: #0a1528;
    color: inherit;
    font-size: 12px;
  }
  .voice-button {
    margin-left: auto;
    padding: 8px 14px;
    border-radius: 999px;
    border: 1px solid #00b7ff;
    background: rgb(0 183 255 / 15%);
    color: #f7fbff;
    font-size: 13px;
    font-weight: 600;
    cursor: pointer;
  }
  .voice-button.live {
    border-color: #ff5c7a;
    background: rgb(255 92 122 / 18%);
  }
  .voice-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .voice-state,
  .caption {
    margin-top: 8px;
    font-size: 12px;
    color: #8fa5bf;
  }
  .caption {
    color: #00b7ff;
  }
  .loading-bar {
    margin-top: 10px;
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 12px;
    color: #7dd3fc;
    padding: 8px 10px;
    border-radius: 10px;
    background: rgb(0 183 255 / 10%);
    border: 1px solid rgb(0 183 255 / 25%);
  }
  .spinner {
    width: 14px;
    height: 14px;
    border: 2px solid rgb(0 183 255 / 30%);
    border-top-color: #00b7ff;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
    flex-shrink: 0;
  }
  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    padding: 0 14px 8px;
    flex-shrink: 0;
  }
  .chip {
    border: 1px solid rgb(0 183 255 / 22%);
    background: rgb(12 24 40);
    color: inherit;
    border-radius: 999px;
    padding: 5px 10px;
    font-size: 11px;
    cursor: pointer;
  }
  .chip.linkish {
    text-decoration: none;
    display: inline-flex;
    align-items: center;
  }
  .msgs {
    flex: 1;
    overflow-y: auto;
    padding: 8px 14px 12px;
    min-height: 160px;
  }
  .msg {
    display: flex;
    gap: 8px;
    margin-bottom: 10px;
    align-items: flex-end;
  }
  .msg.user {
    flex-direction: row-reverse;
  }
  .dot {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    background: #122038;
    border: 1px solid rgb(0 183 255 / 25%);
    display: grid;
    place-items: center;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }
  .bubble {
    max-width: 78%;
    padding: 10px 12px;
    border-radius: 14px;
    font-size: 13px;
    line-height: 1.45;
    background: #122038;
    border: 1px solid rgb(0 183 255 / 18%);
    white-space: pre-wrap;
  }
  .bubble.user {
    background: rgb(0 100 140 / 35%);
    border-color: rgb(0 183 255 / 30%);
  }
  .km-miss {
    margin-top: 6px;
    font-size: 11px;
    color: #ffb86c;
  }
  .composer {
    display: flex;
    gap: 8px;
    padding: 10px 12px 12px;
    border-top: 1px solid rgb(0 183 255 / 18%);
    background: rgb(8 16 28);
    flex-shrink: 0;
  }
  .composer input {
    flex: 1;
    min-width: 0;
    padding: 10px 12px;
    border-radius: 12px;
    border: 1px solid rgb(0 183 255 / 22%);
    background: #0a1528;
    color: inherit;
    font-size: 14px;
  }
  .send {
    padding: 10px 14px;
    border-radius: 12px;
    border: none;
    background: #00b7ff;
    color: #041018;
    font-weight: 600;
    cursor: pointer;
  }
  .send:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .btn {
    padding: 6px 12px;
    border-radius: 8px;
    border: 1px solid #00b7ff;
    background: rgb(0 183 255 / 18%);
    color: inherit;
    font-size: 12px;
    cursor: pointer;
  }
</style>
