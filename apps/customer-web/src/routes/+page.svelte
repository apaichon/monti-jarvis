<script lang="ts">
  import { onMount, tick } from 'svelte';
  import Portrait from '$lib/components/Portrait.svelte';
  import Waveform from '$lib/components/Waveform.svelte';
  import {
    addTurn,
    createCall,
    endCall,
    subscribeTurns,
    type CallSession
  } from '$lib/api/calls';
  import { sendChat, type ChatMessage as ChatHistoryEntry, type ChatSource } from '$lib/api/chat';
  import { formatInfra, loadInfra } from '$lib/api/infra';
  import { classifyTone } from '$lib/tone';
  import { loadWorkforce, type Agent } from '$lib/api/workforce';
  import { GeminiVoice } from '$lib/voice/gemini';
  import {
    getStoredCustomer,
    loadCustomerMe,
    logoutCustomer,
    requestCustomerOTP,
    verifyCustomerOTP,
    type CustomerProfile
  } from '$lib/api/customerAuth';

  type UiMessage = {
    id: string;
    role: 'assistant' | 'user';
    content: string;
    initial: string;
    voiceRole?: string;
    sources?: ChatSource[];
    missingKm?: boolean;
  };

  const topics = [
    { id: 'general', label: 'General' },
    { id: 'billing', label: 'Billing' },
    { id: 'technical', label: 'Technical' }
  ] as const;

  let agents = $state<Agent[]>([]);
  let selectedAgent = $state<Agent | null>(null);
  let session = $state<CallSession | null>(null);
  let voice = $state<GeminiVoice | null>(null);
  let live = $state(false);
  let busy = $state(false);
  let error = $state('');
  let timer = $state('00:00:00');
  let voiceState = $state('Select an agent, then start an inbound voice call.');
  let topic = $state<(typeof topics)[number]['id']>('general');
  let chatSessionId = $state('');
  let chatHistory = $state<ChatHistoryEntry[]>([]);
  let messages = $state<UiMessage[]>([
    {
      id: 'welcome',
      role: 'assistant',
      content:
        'Welcome to Monti Inbound Call Center. Choose an AI agent on the left, then type a question or start a voice call.',
      initial: 'A'
    }
  ]);
  let input = $state('');
  let infraStatus = $state('checking infra');
  let chatEl: HTMLElement | undefined = $state();
  let tenantId = $state('');
  let customer = $state<CustomerProfile | null>(null);
  let customerEmail = $state('');
  let customerName = $state('');
  let otp = $state('');
  let challengeId = $state('');
  let authStatus = $state('');
  let authBusy = $state(false);
  let pickerOpen = $state(false);

  let tone = $state('');
  let toneTimer: ReturnType<typeof setTimeout> | undefined;

  let startedAt = 0;
  let timerId: ReturnType<typeof setInterval> | undefined;
  let unsubscribe: (() => void) | undefined;
  const transcriptKeys = new Set<string>();

  onMount(async () => {
    tenantId = new URLSearchParams(window.location.search).get('tenant_id')?.trim() || '';
    customer = getStoredCustomer();
    void loadCustomerMe().then((profile) => {
      if (profile) customer = profile;
    });
    try {
      agents = await loadWorkforce(tenantId ? { tenantId } : undefined);
      selectedAgent = agents.find((a) => a.popular) || agents[0] || null;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load agents';
    }
    const infra = await loadInfra();
    infraStatus = formatInfra(infra);
  });

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

  // Match the portrait's expression to the tone of the assistant's reply,
  // then fall back to the neutral talking loop / still image.
  function showTone(text: string) {
    const detected = classifyTone(text);
    if (!detected) return;
    tone = detected;
    clearTimeout(toneTimer);
    toneTimer = setTimeout(() => (tone = ''), 4200);
  }

  function addMessage(role: 'assistant' | 'user', content: string, initial: string) {
    const msg: UiMessage = {
      id: `${Date.now()}-${Math.random()}`,
      role,
      content,
      initial
    };
    messages = [...messages, msg];
    void scrollChat();
    return msg;
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

  async function selectAgent(agent: Agent) {
    if (live) await hangUp();
    selectedAgent = agent;
    pickerOpen = false;
    if (agent.greeting) {
      addMessage('assistant', agent.greeting, agentInitial(agent.name));
      showTone(agent.greeting);
    }
  }

  async function persistTurn(callId: string, role: string, content: string) {
    try {
      await addTurn(callId, role, content);
    } catch {
      // transcript still visible locally if persist fails
    }
  }

  function upsertVoiceTurn(role: string, content: string) {
    const key = `${role}:${content}`;
    if (transcriptKeys.has(key)) return;
    transcriptKeys.add(key);
    const uiRole = role === 'caller' ? 'user' : 'assistant';
    const initial = uiRole === 'assistant' ? agentInitial(selectedAgent?.name) : 'C';
    appendOrMergeTranscript(uiRole, content, initial);
    if (uiRole === 'assistant') showTone(content);
  }

  async function startCall() {
    if (!selectedAgent) {
      error = 'Select an AI agent first.';
      return;
    }
    error = '';
    busy = true;
    transcriptKeys.clear();
    voiceState = 'Connecting…';
    try {
      const gemini = new GeminiVoice();
      // Show greeting text immediately while audio path connects.
      if (selectedAgent.greeting) {
        upsertVoiceTurn('agent', selectedAgent.greeting);
      }
      const [created] = await Promise.all([
        createCall(tenantId ? { tenantId } : undefined),
        gemini.start(
          selectedAgent.id,
          topic,
          {
            onLive: (v) => {
              live = v;
              if (v) {
                voiceState = `On call with ${selectedAgent?.name} — listen for the greeting…`;
              } else {
                voiceState = `Ready to call ${selectedAgent?.name}.`;
              }
            },
            onStatus: (message) => {
              voiceState = message;
            },
            onTranscript: (role, text, meta) => {
              // Live caption grows as partial ASR chunks merge into full sentences.
              upsertVoiceTurn(role, text);
              // Persist only finalized turns (not every short partial fragment).
              if (meta?.final && session) void persistTurn(session.id, role, text);
            },
            onError: (message) => {
              error = message;
            }
          },
          { lang: 'auto', tenantId: tenantId || undefined }
        )
      ]);

      session = created;
      chatSessionId = created.id;
      unsubscribe = subscribeTurns(
        created.id,
        (turn) => {
          upsertVoiceTurn(turn.role, turn.content);
        },
        tenantId ? { tenantId } : undefined
      );

      voice = gemini;
      live = true;
      startTimer();
      if (!voiceState.startsWith('On call')) {
        voiceState = `On call with ${selectedAgent.name} — agent greets first.`;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Call failed';
      await cleanup(false);
    } finally {
      busy = false;
    }
  }

  async function hangUp() {
    if (!session) return;
    busy = true;
    try {
      await voice?.stop();
      await endCall(session.id);
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
    unsubscribe?.();
    unsubscribe = undefined;
    voice = null;
    if (resetSession) session = null;
    voiceState = selectedAgent
      ? `Ready to call ${selectedAgent.name}.`
      : 'Select an agent, then start an inbound voice call.';
  }

  async function toggleCall() {
    if (live) await hangUp();
    else await startCall();
  }

  async function sendOTP(event: Event) {
    event.preventDefault();
    authBusy = true;
    authStatus = '';
    try {
      const res = await requestCustomerOTP({
        tenant_id: tenantId || undefined,
        email: customerEmail.trim(),
        display_name: customerName.trim()
      }, tenantId ? { tenantId } : undefined);
      challengeId = res.challenge_id;
      authStatus = `OTP sent to ${res.delivery.to}`;
    } catch (err) {
      authStatus = err instanceof Error ? err.message : 'Failed to send OTP';
    } finally {
      authBusy = false;
    }
  }

  async function verifyOTP(event: Event) {
    event.preventDefault();
    authBusy = true;
    authStatus = '';
    try {
      const res = await verifyCustomerOTP({
        tenant_id: tenantId || undefined,
        challenge_id: challengeId,
        otp: otp.trim()
      }, tenantId ? { tenantId } : undefined);
      customer = res.customer;
      customerEmail = '';
      customerName = '';
      otp = '';
      challengeId = '';
      authStatus = `Signed in as ${res.customer.display_name || res.customer.email}`;
    } catch (err) {
      authStatus = err instanceof Error ? err.message : 'OTP verification failed';
    } finally {
      authBusy = false;
    }
  }

  async function signOutCustomer() {
    await logoutCustomer();
    customer = null;
    authStatus = 'Signed out';
  }

  async function submitChat(event: Event) {
    event.preventDefault();
    if (!selectedAgent) {
      error = 'Select an AI agent first.';
      return;
    }
    const text = input.trim();
    if (!text) return;

    input = '';
    error = '';
    addMessage('user', text, 'C');

    const payloadHistory = chatHistory.slice();
    chatHistory = [...chatHistory, { role: 'user', content: text }];
    busy = true;

    const thinking = addMessage('assistant', 'One moment...', agentInitial(selectedAgent.name));
    try {
      const data = await sendChat(
        {
          session_id: chatSessionId,
          agent_id: selectedAgent.id,
          topic,
          message: text,
          history: payloadHistory
        },
        tenantId ? { tenantId } : undefined
      );
      chatSessionId = data.session_id;
      messages = messages.map((m) =>
        m.id === thinking.id
          ? { ...m, content: data.reply, sources: data.sources, missingKm: data.missing_km }
          : m
      );
      chatHistory = [...chatHistory, { role: 'assistant', content: data.reply }];
      showTone(data.reply);
    } catch (err) {
      messages = messages.filter((m) => m.id !== thinking.id);
      chatHistory = chatHistory.slice(0, -1);
      error = err instanceof Error ? err.message : 'Chat failed';
    } finally {
      busy = false;
      void scrollChat();
    }
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      void submitChat(event);
    }
  }

  const sessionLabel = $derived(
    session
      ? `Call ${session.id.slice(0, 8)} · ${selectedAgent?.name ?? 'agent'}`
      : chatSessionId
        ? `Call ${chatSessionId.slice(0, 8)} · ${selectedAgent?.name ?? 'agent'}`
        : 'New call session'
  );
</script>

<main class="app">
  <aside class="panel control-panel">
    <header class="brand">
      <img class="brand-mark" src="/images/monti-logo.png" width="46" height="46" alt="Monti AI Ambassadors" />
      <div>
        <h1>MONTI</h1>
        <p>Inbound Call Center · AI Workforce</p>
        {#if tenantId}
          <p>Tenant · {tenantId}</p>
        {/if}
      </div>
    </header>

    <section class="voice-card" aria-label="Customer sign in">
      {#if customer}
        <div style="display:flex;justify-content:space-between;gap:12px;align-items:center">
          <div>
            <div style="font-weight:700">{customer.display_name || customer.email}</div>
            <div class="voice-state">Signed in · {customer.email}</div>
          </div>
          <button class="voice-button" type="button" onclick={signOutCustomer}>Sign out</button>
        </div>
      {:else}
        <form onsubmit={challengeId ? verifyOTP : sendOTP} style="display:grid;gap:10px">
          <div class="voice-state">Optional customer sign-in for account-aware support.</div>
          <input
            type="email"
            bind:value={customerEmail}
            placeholder="customer@example.com"
            autocomplete="email"
            disabled={authBusy || !!challengeId}
            style="width:100%;box-sizing:border-box;border:1px solid var(--line);border-radius:14px;background:#071120;color:var(--text);padding:12px"
          />
          {#if !challengeId}
            <input
              type="text"
              bind:value={customerName}
              placeholder="Name (optional)"
              autocomplete="name"
              disabled={authBusy}
              style="width:100%;box-sizing:border-box;border:1px solid var(--line);border-radius:14px;background:#071120;color:var(--text);padding:12px"
            />
          {:else}
            <input
              type="text"
              bind:value={otp}
              placeholder="6-digit OTP"
              inputmode="numeric"
              autocomplete="one-time-code"
              disabled={authBusy}
              style="width:100%;box-sizing:border-box;border:1px solid var(--line);border-radius:14px;background:#071120;color:var(--text);padding:12px"
            />
          {/if}
          <button class="voice-button" type="submit" disabled={authBusy || (!challengeId && !customerEmail.trim()) || (!!challengeId && !otp.trim())}>
            {authBusy ? '…' : challengeId ? 'Verify OTP' : 'Send OTP'}
          </button>
        </form>
      {/if}
      {#if authStatus}
        <div class="voice-state" style="margin-top:8px">{authStatus}</div>
      {/if}
    </section>

    {#if selectedAgent}
      <section class="assistant-orb">
        <div class="halo" style="--assistant-color:{selectedAgent.color}">
          <Portrait agent={selectedAgent} speaking={live} {tone} />
        </div>
        <div class="assistant-name">
          <h2>{selectedAgent.name}</h2>
          <p>{selectedAgent.role} · {selectedAgent.trait}</p>
        </div>
        <button class="picker-trigger" type="button" disabled={live} onclick={() => (pickerOpen = true)}>
          <span aria-hidden="true">◇</span>
          Change avatar
        </button>
        <Waveform color={selectedAgent.color} />
      </section>
    {/if}

    <section class="voice-card">
      <div class="voice-row">
        <div class="status-pill">{timer}</div>
        <button
          class="voice-button"
          class:live={live}
          type="button"
          disabled={busy}
          onclick={toggleCall}
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

    <section class="assistants" aria-label="Choose AI agent">
      {#each agents as agent (agent.id)}
        <button
          type="button"
          class="assistant-card"
          class:active={selectedAgent?.id === agent.id}
          style="--assistant-color:{agent.color}"
          disabled={live}
          onclick={() => selectAgent(agent)}
        >
          <Portrait {agent} mini />
          <div>
            <div>
              <strong>{agent.name}</strong>
              {#if agent.popular}
                <span class="tag">Popular</span>
              {/if}
            </div>
            <div class="assistant-meta">{agent.role}</div>
            <div class="assistant-meta">{agent.trait}</div>
            <Waveform color={agent.color} count={16} mini />
          </div>
          <span class="tag">Select</span>
        </button>
      {/each}
    </section>
  </aside>

  {#if pickerOpen}
    <div class="picker-backdrop">
      <button class="picker-scrim" type="button" aria-label="Close avatar picker" onclick={() => (pickerOpen = false)}></button>
      <div
        class="picker-dialog"
        role="dialog"
        aria-modal="true"
        aria-label="Select AI avatar"
      >
        <div class="picker-head">
          <div>
            <h2>Select avatar</h2>
            <p>Choose who will answer this customer session.</p>
          </div>
          <button class="picker-close" type="button" aria-label="Close avatar picker" onclick={() => (pickerOpen = false)}>
            ×
          </button>
        </div>
        <div class="picker-grid">
          {#each agents as agent (agent.id)}
            <button
              type="button"
              class="assistant-card picker-card"
              class:active={selectedAgent?.id === agent.id}
              style="--assistant-color:{agent.color}"
              onclick={() => selectAgent(agent)}
            >
              <Portrait {agent} mini />
              <div>
                <div>
                  <strong>{agent.name}</strong>
                  {#if agent.popular}
                    <span class="tag">Popular</span>
                  {/if}
                </div>
                <div class="assistant-meta">{agent.role}</div>
                <div class="assistant-meta">{agent.trait}</div>
              </div>
              <span class="tag">{selectedAgent?.id === agent.id ? 'Current' : 'Select'}</span>
            </button>
          {/each}
        </div>
      </div>
    </div>
  {/if}

  <section class="panel workspace">
    <header class="topbar">
      <div>
        <h2>Caller Desk</h2>
        <div class="tabs" role="tablist" aria-label="Call topic">
          {#each topics as tab (tab.id)}
            <button
              type="button"
              class="tab"
              class:active={topic === tab.id}
              role="tab"
              aria-selected={topic === tab.id}
              onclick={() => (topic = tab.id)}
            >
              {tab.label}
            </button>
          {/each}
        </div>
      </div>
      <div class="infra">{infraStatus}</div>
    </header>

    <section class="chat" aria-live="polite" bind:this={chatEl}>
      {#each messages as msg (msg.id)}
        <div class="msg" class:user={msg.role === 'user'}>
          <div class="dot">{msg.initial}</div>
          <div class="bubble" class:user={msg.role === 'user'}>
            {msg.content}
            {#if msg.sources && msg.sources.length > 0}
              <div class="citations">
                {#each msg.sources as src (src.chunk_id)}
                  <span class="citation" title={src.excerpt}>{src.scope} · KB</span>
                {/each}
              </div>
            {:else if msg.missingKm}
              <div class="citations"><span class="citation warn">No KB match</span></div>
            {/if}
          </div>
        </div>
      {/each}
    </section>

    <section class="composer-wrap">
      <form onsubmit={submitChat}>
        <div class="composer">
          <textarea
            bind:value={input}
            placeholder="Ask your question..."
            autocomplete="off"
            disabled={busy}
            onkeydown={handleKeydown}
          ></textarea>
          <button class="send" type="submit" disabled={busy}>Send</button>
        </div>
        <div class="error">{error}</div>
      </form>
      <div class="infra">{sessionLabel}</div>
    </section>
  </section>
</main>
