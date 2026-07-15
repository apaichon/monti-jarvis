<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { page } from '$app/stores';
  import Portrait from '$lib/components/Portrait.svelte';
  import Waveform from '$lib/components/Waveform.svelte';
  import {
    addTurn,
    createCall,
    endCall,
    subscribeTurns,
    type CallSession
  } from '$lib/api/calls';
  import { sendChat, type ChatMessage as ChatHistoryEntry } from '$lib/api/chat';
  import { classifyTone } from '$lib/tone';
  import { loadWorkforce, type Agent } from '$lib/api/workforce';
  import { GeminiVoice, micAvailabilityError } from '$lib/voice/gemini';
  import {
    assistantConfirmedFarewell,
    customerConfirmedEnd,
    CUSTOMER_END_COUNTDOWN_SECONDS
  } from '$lib/voice/end-call';

  type EmbedConfig = {
    tenant_id: string;
    slug: string;
    name: string;
    embed_key: string;
    default_agent_id?: string;
    agents?: Array<{ id: string; name: string; role?: string; image?: string }>;
  };

  type UiMsg = {
    id: string;
    role: 'user' | 'assistant';
    content: string;
    initial: string;
    voiceRole?: string;
  };

  let key = $state('');
  let tenantId = $state('');
  let workspace = $state('');
  let agents = $state<Agent[]>([]);
  let selected = $state<Agent | null>(null);
  let messages = $state<UiMsg[]>([]);
  let history = $state<ChatHistoryEntry[]>([]);
  let sessionId = $state('');
  let input = $state('');
  let busy = $state(false);
  let error = $state('');
  let loading = $state(true);

  let session = $state<CallSession | null>(null);
  let voice = $state<GeminiVoice | null>(null);
  let live = $state(false);
  let timer = $state('00:00:00');
  let voiceState = $state('Select an agent, then start a voice call or type a message.');
  let micBlocked = $state<string | null>(null);
  let tone = $state('');
  let chatEl: HTMLElement | undefined = $state();

  let startedAt = 0;
  let timerId: ReturnType<typeof setInterval> | undefined;
  let toneTimer: ReturnType<typeof setTimeout> | undefined;
  let customerCloseFallbackTimerId: ReturnType<typeof setTimeout> | undefined;
  let autoCloseTimerId: ReturnType<typeof setInterval> | undefined;
  let customerEndRequested = $state(false);
  let autoClosePending = $state(false);
  let unsubscribe: (() => void) | undefined;
  const transcriptKeys = new Set<string>();

  function resolveParentOrigin(): string {
    const fromQuery = $page.url.searchParams.get('parent_origin') || '';
    if (fromQuery) {
      try {
        const u = new URL(fromQuery);
        if (u.protocol === 'http:' || u.protocol === 'https:') return u.origin;
      } catch {
        /* ignore */
      }
    }
    if (document.referrer) {
      try {
        return new URL(document.referrer).origin;
      } catch {
        /* ignore */
      }
    }
    return '';
  }

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

  async function scrollChat() {
    await tick();
    if (chatEl) chatEl.scrollTop = chatEl.scrollHeight;
  }

  function showTone(text: string) {
    const detected = classifyTone(text);
    if (!detected) return;
    tone = detected;
    clearTimeout(toneTimer);
    toneTimer = setTimeout(() => (tone = ''), 4200);
  }

  function addMessage(role: 'assistant' | 'user', content: string, initial: string) {
    messages = [
      ...messages,
      { id: `${Date.now()}-${Math.random()}`, role, content, initial }
    ];
    void scrollChat();
  }

  function appendOrMergeTranscript(role: 'assistant' | 'user', text: string, initial: string) {
    const last = messages[messages.length - 1];
    if (last?.voiceRole === role) {
      messages = [...messages.slice(0, -1), { ...last, content: text }];
    } else {
      messages = [
        ...messages,
        { id: `${Date.now()}-${Math.random()}`, role, content: text, initial, voiceRole: role }
      ];
    }
    void scrollChat();
  }

  function upsertVoiceTurn(role: string, content: string) {
    const k = `${role}:${content}`;
    if (transcriptKeys.has(k)) return;
    transcriptKeys.add(k);
    const uiRole = role === 'caller' ? 'user' : 'assistant';
    const initial = uiRole === 'assistant' ? agentInitial(selected?.name) : 'C';
    appendOrMergeTranscript(uiRole, content, initial);
    if (uiRole === 'assistant') showTone(content);
    if (uiRole === 'assistant' && customerEndRequested && assistantConfirmedFarewell(content)) {
      startCustomerFinishedCountdown();
    }
  }

  async function persistTurn(callId: string, role: string, content: string) {
    try {
      await addTurn(callId, role, content, { tenantId });
    } catch {
      /* local transcript still shown */
    }
  }

  function agentTopic(agentId: string): string {
    const topicByAgent: Record<string, string> = {
      ava: 'general',
      max: 'billing',
      luna: 'technical',
      neo: 'general'
    };
    return topicByAgent[agentId] || 'general';
  }

  async function startCall() {
    if (!selected || !tenantId) {
      error = 'Select an AI agent first.';
      return;
    }
    error = '';
    busy = true;
    transcriptKeys.clear();
    customerEndRequested = false;
    autoClosePending = false;
    voiceState = 'Connecting…';
    let gemini: GeminiVoice | undefined;
    try {
      const topic = agentTopic(selected.id);
      const created = await createCall({ tenantId });
      session = created;
      sessionId = created.id;

      gemini = new GeminiVoice();
      voice = gemini;
      if (selected.greeting) {
        upsertVoiceTurn('agent', selected.greeting);
      }
      await gemini.start(
        selected.id,
        topic,
        {
          onLive: (v) => {
            live = v;
            if (v) {
              voiceState = `On call with ${selected?.name} — listen for the greeting…`;
            } else {
              voiceState = `Ready to call ${selected?.name}.`;
            }
          },
          onStatus: (message) => {
            voiceState = message;
          },
          onTranscript: (role, text, meta) => {
            upsertVoiceTurn(role, text);
            if (role === 'caller' && customerConfirmedEnd(text)) {
              requestCustomerFinishedClose();
            }
            if (meta?.final && session) void persistTurn(session.id, role, text);
          },
          onCustomerEndRequested: () => requestCustomerFinishedClose(),
          onError: (message) => {
            error = message;
            voiceState = message;
          }
        },
        { tenantId, lang: 'auto' }
      );

      voice = gemini;
      unsubscribe = subscribeTurns(
        created.id,
        (turn) => {
          upsertVoiceTurn(turn.role, turn.content);
          if (turn.role === 'caller' && customerConfirmedEnd(turn.content)) {
            requestCustomerFinishedClose();
          }
        },
        { tenantId }
      );
      live = true;
      startTimer();
      if (!voiceState.includes('greeting') && !voiceState.includes('Connected')) {
        voiceState = `On call with ${selected.name} — agent greets first.`;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Call failed';
      await gemini?.stop().catch(() => {});
      if (session) await endCall(session.id, { tenantId }).catch(() => {});
      await cleanup(true);
    } finally {
      busy = false;
    }
  }

  function requestCustomerFinishedClose() {
    if (!live || !session || customerEndRequested || autoClosePending) return;
    customerEndRequested = true;
    voiceState = 'The agent is saying goodbye. The call will end in 5 seconds.';
    const sent = voice?.sendText(
      'The caller said there is nothing else and thanked you. Respond in Thai: "ขออนุญาตวางสายก่อนนะครับ ขอบคุณครับ". Do not ask another question. The call will close in five seconds.'
    );
    if (!sent) {
      startCustomerFinishedCountdown();
      return;
    }
    customerCloseFallbackTimerId = setTimeout(startCustomerFinishedCountdown, 2000);
  }

  function startCustomerFinishedCountdown() {
    if (!live || !session || !customerEndRequested || autoClosePending) return;
    clearTimeout(customerCloseFallbackTimerId);
    customerCloseFallbackTimerId = undefined;
    autoClosePending = true;
    let seconds = CUSTOMER_END_COUNTDOWN_SECONDS;
    voiceState = `The call will end in ${seconds} seconds.`;
    autoCloseTimerId = setInterval(() => {
      seconds -= 1;
      if (seconds <= 0) {
        clearInterval(autoCloseTimerId);
        autoCloseTimerId = undefined;
        void endActiveCall('customer_finished');
        return;
      }
      voiceState = `The call will end in ${seconds} seconds.`;
    }, 1000);
  }

  async function endActiveCall(_reason: 'manual' | 'customer_finished' = 'manual') {
    if (!session) return;
    busy = true;
    try {
      await voice?.stop();
      await endCall(session.id, { tenantId });
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to end call';
    } finally {
      await cleanup(true);
      busy = false;
    }
  }

  async function cleanup(resetSession: boolean) {
    live = false;
    stopTimer();
    clearTimeout(customerCloseFallbackTimerId);
    clearInterval(autoCloseTimerId);
    customerCloseFallbackTimerId = undefined;
    autoCloseTimerId = undefined;
    customerEndRequested = false;
    autoClosePending = false;
    unsubscribe?.();
    unsubscribe = undefined;
    voice = null;
    if (resetSession) session = null;
    voiceState = selected
      ? `Ready to call ${selected.name}.`
      : 'Select an agent, then start a voice call or type a message.';
  }

  async function toggleCall() {
    if (live) await endActiveCall();
    else await startCall();
  }

  async function selectAgent(agent: Agent) {
    if (live) await endActiveCall();
    selected = agent;
    voiceState = `Ready to call ${agent.name}.`;
    if (agent.greeting) {
      addMessage('assistant', agent.greeting, agentInitial(agent.name));
      showTone(agent.greeting);
    }
  }

  onMount(async () => {
    micBlocked = micAvailabilityError();
    key = $page.url.searchParams.get('key') || '';
    if (!key) {
      error = 'Missing embed key';
      loading = false;
      return;
    }
    try {
      const parentOrigin = resolveParentOrigin();
      const qs = new URLSearchParams();
      if (parentOrigin) qs.set('parent_origin', parentOrigin);
      const q = qs.toString();
      const res = await fetch(
        `/api/public/embed/${encodeURIComponent(key)}${q ? `?${q}` : ''}`,
        parentOrigin ? { headers: { 'X-Embed-Parent-Origin': parentOrigin } } : undefined
      );
      const data = (await res.json()) as EmbedConfig & { error?: string; code?: string };
      if (!res.ok) {
        error = data.error || data.code || 'Embed unavailable';
        loading = false;
        return;
      }
      tenantId = data.tenant_id;
      workspace = data.name || data.slug || data.tenant_id;

      // Full agent portraits / expressions from workforce API
      try {
        agents = await loadWorkforce({ tenantId });
      } catch {
        agents = (data.agents || []).map((a) => ({
          id: a.id,
          name: a.name,
          role: a.role || '',
          trait: '',
          color: '#00b7ff',
          image: a.image || `/images/${a.id}.jpg`
        }));
      }

      const def = data.default_agent_id;
      selected =
        (def && agents.find((a) => a.id === def)) ||
        agents.find((a) => a.popular) ||
        agents[0] ||
        null;

      const welcome =
        selected?.greeting ||
        `Hi! You're chatting with ${workspace}. Ask a question or start a voice call.`;
      messages = [
        {
          id: 'welcome',
          role: 'assistant',
          content: welcome,
          initial: agentInitial(selected?.name)
        }
      ];
      if (selected) {
        voiceState = micBlocked
          ? 'Voice unavailable on this host — text chat still works.'
          : `Ready to call ${selected.name}.`;
        showTone(welcome);
      }
    } catch (e) {
      error = e instanceof Error ? e.message : 'Failed to load embed';
    } finally {
      loading = false;
    }
  });

  async function send() {
    const text = input.trim();
    if (!text || !selected || !tenantId || busy || live) return;
    busy = true;
    error = '';
    input = '';
    addMessage('user', text, 'C');
    const payloadHistory = history.slice();
    history = [...history, { role: 'user', content: text }];
    const thinkingId = `${Date.now()}-think`;
    messages = [
      ...messages,
      {
        id: thinkingId,
        role: 'assistant',
        content: 'One moment…',
        initial: agentInitial(selected.name)
      }
    ];
    void scrollChat();
    try {
      const topic = agentTopic(selected.id);
      const res = await sendChat(
        {
          session_id: sessionId,
          agent_id: selected.id,
          topic,
          message: text,
          history: payloadHistory
        },
        { tenantId }
      );
      sessionId = res.session_id || sessionId;
      history = [...history, { role: 'assistant', content: res.reply }];
      messages = messages.map((m) =>
        m.id === thinkingId ? { ...m, content: res.reply } : m
      );
      showTone(res.reply);
    } catch (e) {
      messages = messages.filter((m) => m.id !== thinkingId);
      history = history.slice(0, -1);
      error = e instanceof Error ? e.message : 'Chat failed';
    } finally {
      busy = false;
      void scrollChat();
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      void send();
    }
  }
</script>

<svelte:head>
  <title>Monti Embed</title>
  <meta name="robots" content="noindex" />
</svelte:head>

<div class="embed-shell">
  {#if loading}
    <p class="muted center">Loading…</p>
  {:else if error && !tenantId}
    <p class="err center">{error}</p>
  {:else}
    <header class="embed-hdr">
      <div class="brand">
        <img class="brand-mark" src="/images/monti-logo.png" width="28" height="28" alt="" />
        <div>
          <strong>{workspace || 'Monti'}</strong>
          <span class="sub">AI · text & voice</span>
        </div>
      </div>
      {#if agents.length > 1}
        <select
          class="agent-sel"
          value={selected?.id || ''}
          disabled={live}
          onchange={(e) => {
            const id = (e.currentTarget as HTMLSelectElement).value;
            const a = agents.find((x) => x.id === id);
            if (a) void selectAgent(a);
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
          <Portrait agent={selected} speaking={live} {tone} />
        </div>
        <div class="agent-meta">
          <h2>{selected.name}</h2>
          <p>{selected.role}{selected.trait ? ` · ${selected.trait}` : ''}</p>
        </div>
        <Waveform color={selected.color || '#00b7ff'} count={28} />
      </section>
    {/if}

    <section class="voice-card">
      {#if micBlocked}
        <p class="mic-warn" title={micBlocked}>
          Mic blocked: open Monti via <strong>localhost</strong> or <strong>HTTPS</strong> (not
          http://custom-host). Text chat still works.
        </p>
      {/if}
      <div class="voice-row">
        <div class="status-pill">{timer}</div>
        <button
          class="voice-button"
          class:live
          type="button"
          disabled={busy || !selected || (!!micBlocked && !live)}
          onclick={toggleCall}
          title={micBlocked || undefined}
        >
          {live ? 'End call' : busy ? 'Connecting…' : 'Start call'}
        </button>
      </div>
      {#if busy && !live}
        <div class="voice-state loading" aria-live="polite">⏳ {voiceState}</div>
      {:else}
        <div class="voice-state">{voiceState}</div>
      {/if}
    </section>

    <div class="msgs" bind:this={chatEl} aria-live="polite">
      {#each messages as m (m.id)}
        <div class="msg" class:user={m.role === 'user'}>
          <div class="dot">{m.initial}</div>
          <div class="bubble" class:user={m.role === 'user'}>{m.content}</div>
        </div>
      {/each}
    </div>

    {#if error}
      <p class="err small">{error}</p>
    {/if}

    <footer class="composer">
      <input
        type="text"
        placeholder={live ? 'On a voice call…' : 'Type a message…'}
        bind:value={input}
        onkeydown={onKey}
        disabled={busy || !selected || live}
      />
      <button
        type="button"
        class="send"
        onclick={send}
        disabled={busy || live || !input.trim() || !selected}
      >
        {busy ? '…' : 'Send'}
      </button>
    </footer>
  {/if}
</div>

<style>
  :global(html),
  :global(body) {
    margin: 0;
    height: 100%;
    overflow: hidden;
  }
  .embed-shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
    height: 100dvh;
    background: #05101f;
    color: #f7fbff;
    font-family: Inter, system-ui, sans-serif;
  }
  .embed-hdr {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    padding: 12px 48px 10px 14px; /* room for host-page close button */
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
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .brand .sub {
    display: block;
    font-size: 10px;
    color: #8fa5bf;
  }
  .brand-mark {
    border-radius: 50%;
    flex-shrink: 0;
  }
  .agent-sel {
    background: #0c1a2e;
    color: #f7fbff;
    border: 1px solid rgb(0 183 255 / 30%);
    border-radius: 8px;
    padding: 6px 8px;
    font-size: 12px;
    max-width: 110px;
  }
  .orb {
    display: grid;
    justify-items: center;
    gap: 6px;
    padding: 10px 12px 4px;
    flex-shrink: 0;
  }
  .halo {
    width: 132px;
    height: 132px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    border: 2px solid var(--assistant-color, #00b7ff);
    box-shadow: 0 0 28px color-mix(in srgb, var(--assistant-color, #00b7ff) 65%, transparent);
    animation: breathe 2.4s ease-in-out infinite;
  }
  .halo :global(.portrait) {
    width: 108px;
    height: 108px;
  }
  @keyframes breathe {
    50% {
      transform: scale(1.03);
    }
  }
  .agent-meta {
    text-align: center;
  }
  .agent-meta h2 {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
  }
  .agent-meta p {
    margin: 2px 0 0;
    font-size: 11px;
    color: #8fa5bf;
  }
  .voice-card {
    margin: 4px 12px 8px;
    padding: 10px;
    border-radius: 12px;
    border: 1px solid rgb(0 183 255 / 22%);
    background: rgb(8 20 36 / 80%);
    flex-shrink: 0;
  }
  .voice-row {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .status-pill {
    flex: 0 0 auto;
    min-width: 72px;
    text-align: center;
    padding: 8px 10px;
    border-radius: 10px;
    border: 1px solid rgb(0 183 255 / 28%);
    font-variant-numeric: tabular-nums;
    font-size: 13px;
    color: #9ec9ff;
  }
  .voice-button {
    flex: 1;
    border: 0;
    border-radius: 10px;
    padding: 10px 12px;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
    color: #fff;
    background: linear-gradient(135deg, #0084ff, #00b7ff);
  }
  .voice-button.live {
    background: linear-gradient(135deg, #c0392b, #e74c3c);
  }
  .voice-button:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .voice-state {
    margin-top: 8px;
    font-size: 11px;
    color: #8fa5bf;
    line-height: 1.35;
  }
  .mic-warn {
    margin: 0 0 8px;
    padding: 8px 10px;
    border-radius: 8px;
    font-size: 11px;
    line-height: 1.35;
    color: #ffc9a8;
    background: rgb(180 80 20 / 18%);
    border: 1px solid rgb(255 140 60 / 35%);
  }
  .mic-warn strong {
    color: #ffe0c2;
  }
  .msgs {
    flex: 1;
    overflow: auto;
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
    min-height: 0;
  }
  .msg {
    display: flex;
    gap: 8px;
    align-items: flex-end;
  }
  .msg.user {
    flex-direction: row-reverse;
  }
  .dot {
    width: 24px;
    height: 24px;
    border-radius: 50%;
    background: rgb(0 183 255 / 25%);
    border: 1px solid rgb(0 183 255 / 35%);
    display: grid;
    place-items: center;
    font-size: 11px;
    font-weight: 600;
    flex-shrink: 0;
  }
  .bubble {
    max-width: 78%;
    padding: 8px 11px;
    border-radius: 14px;
    font-size: 13px;
    line-height: 1.4;
    white-space: pre-wrap;
    background: rgb(0 100 200 / 18%);
    border: 1px solid rgb(0 183 255 / 22%);
  }
  .bubble.user {
    background: rgb(0 140 255 / 35%);
  }
  .composer {
    display: flex;
    gap: 8px;
    padding: 10px 12px 14px;
    border-top: 1px solid rgb(0 183 255 / 18%);
    flex-shrink: 0;
    background: rgb(5 16 31 / 98%);
  }
  .composer input {
    flex: 1;
    border-radius: 10px;
    border: 1px solid rgb(0 183 255 / 30%);
    background: #0a1628;
    color: #f7fbff;
    padding: 10px 12px;
    font-size: 14px;
  }
  .send {
    border: 0;
    border-radius: 10px;
    padding: 0 14px;
    background: linear-gradient(135deg, #0084ff, #00b7ff);
    color: #fff;
    font-weight: 600;
    cursor: pointer;
    min-width: 64px;
  }
  .send:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .muted {
    color: #8fa5bf;
  }
  .err {
    color: #ff7a90;
    padding: 16px;
  }
  .err.small {
    padding: 0 12px 6px;
    font-size: 12px;
    flex-shrink: 0;
  }
  .center {
    text-align: center;
    margin-top: 40%;
  }
</style>
