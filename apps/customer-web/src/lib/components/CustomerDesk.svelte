<script lang="ts">
  import { onMount, tick } from 'svelte';
  import Portrait from '$lib/components/Portrait.svelte';
  import Waveform from '$lib/components/Waveform.svelte';
  import {
    addTurn,
    archiveCallAudio,
    createCall,
    endCall,
    submitCallRating,
    subscribeTurns,
    type CallSession
  } from '$lib/api/calls';
  import { sendChat, type ChatMessage as ChatHistoryEntry, type ChatSource, type TicketOffer } from '$lib/api/chat';
  import { createCustomerTicket } from '$lib/api/tickets';
  import { formatSystemLive, loadInfra, systemLiveState, type SystemLiveState } from '$lib/api/infra';
  import { classifyTone } from '$lib/tone';
  import { loadWorkforce, type Agent } from '$lib/api/workforce';
  import { GeminiVoice } from '$lib/voice/gemini';
  import {
    assistantConfirmedFarewell,
    customerConfirmedEnd,
    CUSTOMER_END_COUNTDOWN_SECONDS
  } from '$lib/voice/end-call';
  import {
    getStoredCustomer,
    loadCustomerPortalPolicy,
    loadCustomerQuota,
    loadCustomerMe,
    logoutCustomer,
    requestCustomerOTP,
    verifyCustomerOTP,
    type CustomerPortalPolicy,
    type CustomerQuotaSummary,
    type CustomerProfile
  } from '$lib/api/customerAuth';
  import {
    applyThemeTokens,
    fetchPublicTheme,
    resolveBranding,
    type ThemeBranding
  } from '$lib/theme/applyTheme';
  import { brandMonogram } from '$lib/brandMark';
  import {
    ensureAudioPermission,
    friendlyMediaError,
    getStoredInputDeviceId,
    getStoredOutputDeviceId,
    listAudioDevices,
    storeInputDeviceId,
    storeOutputDeviceId,
    supportsSpeakerSelection,
    type AudioDevice
  } from '$lib/audio/devices';

  type Props = {
    tenantId: string;
    tenantSlug?: string;
    tenantName?: string;
    onChangeTenant?: () => void;
  };

  let { tenantId, tenantSlug = '', tenantName = '', onChangeTenant }: Props = $props();

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
  let remainingTimer = $state('00:00:00');
  let remainingSeconds = $state(0);
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
  let systemLive = $state('Checking…');
  let systemLiveKind = $state<SystemLiveState>('checking');
  let chatEl: HTMLElement | undefined = $state();
  let brand = $state(resolveBranding(null));
  let tenantLabel = $derived(tenantName || tenantSlug || tenantId);
  let companyMonogram = $derived(brandMonogram(tenantName || brand.brand_name, tenantSlug));
  let audioInputs = $state<AudioDevice[]>([]);
  let audioOutputs = $state<AudioDevice[]>([]);
  let selectedMicId = $state('');
  let selectedSpeakerId = $state('');
  let audioBusy = $state(false);
  let audioNote = $state('');
  let speakerSelectable = $state(false);
  let audioOpen = $state(true);
  let audioTesting = $state(false);
  let micLevel = $state(0);
  let openPanel = $state<'notifications' | 'language' | 'security' | 'about' | ''>('');
  let appVersion = $state('');
  let audioTestStream: MediaStream | null = null;
  let audioTestRaf = 0;
  let customer = $state<CustomerProfile | null>(null);
  let customerEmail = $state('');
  let customerName = $state('');
  let otp = $state('');
  let challengeId = $state('');
  let authStatus = $state('');
  let authBusy = $state(false);
  let pickerOpen = $state(false);
  let ratingOpen = $state(false);
  let ratingCallId = $state('');
  let ratingScore = $state(0);
  let ratingBusy = $state(false);
  let ratingError = $state('');
  let portalPolicy = $state<CustomerPortalPolicy | null>(null);
  let quota = $state<CustomerQuotaSummary | null>(null);
  let callControlsExpanded = $state(false);
  let ticketOffer = $state<TicketOffer | null>(null);
  let ticketContactEmail = $state('');
  let ticketContactName = $state('');
  let ticketBusy = $state(false);
  let ticketError = $state('');

  let tone = $state('');
  let toneTimer: ReturnType<typeof setTimeout> | undefined;

  let startedAt = 0;
  let activeCallLimitSeconds = 0;
  let timerId: ReturnType<typeof setInterval> | undefined;
  let warningTimerId: ReturnType<typeof setTimeout> | undefined;
  let timeoutTimerId: ReturnType<typeof setTimeout> | undefined;
  let autoCloseTimerId: ReturnType<typeof setInterval> | undefined;
  let customerCloseFallbackTimerId: ReturnType<typeof setTimeout> | undefined;
  let customerEndRequested = $state(false);
  let autoClosePending = $state(false);
  let unsubscribe: (() => void) | undefined;
  const transcriptKeys = new Set<string>();

  onMount(async () => {
    customer = getStoredCustomer();
    if (tenantId) {
      const theme = await fetchPublicTheme(window.location.origin, tenantId);
      brand = resolveBranding(
        theme?.branding as ThemeBranding | undefined,
        tenantName || tenantSlug || tenantId
      );
      applyThemeTokens(document.documentElement, theme);
    }
    void loadCustomerMe(tenantId ? { tenantId } : undefined).then((profile) => {
      // Always apply result so a stale sessionStorage profile without a valid
      // cookie/token cannot unlock Start call.
      customer = profile;
      void refreshPortalState();
    });
    try {
      await refreshPortalState();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load agents';
    }
    const infra = await loadInfra();
    systemLiveKind = systemLiveState(infra);
    systemLive = formatSystemLive(infra);
    selectedMicId = getStoredInputDeviceId();
    selectedSpeakerId = getStoredOutputDeviceId();
    speakerSelectable = supportsSpeakerSelection();
    void refreshAudioDevices(false);
    void fetch('/api/version')
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data?.version) appVersion = String(data.version);
      })
      .catch(() => {});
  });

  async function refreshAudioDevices(requestPermission: boolean) {
    audioBusy = true;
    audioNote = '';
    try {
      if (requestPermission) {
        await ensureAudioPermission();
      }
      const { inputs, outputs } = await listAudioDevices();
      audioInputs = inputs;
      audioOutputs = outputs;
      if (selectedMicId && !inputs.some((d) => d.deviceId === selectedMicId)) {
        selectedMicId = inputs[0]?.deviceId || '';
      } else if (!selectedMicId && inputs[0]) {
        selectedMicId = inputs[0].deviceId;
      }
      if (selectedSpeakerId && !outputs.some((d) => d.deviceId === selectedSpeakerId)) {
        selectedSpeakerId = outputs[0]?.deviceId || '';
      } else if (!selectedSpeakerId && outputs[0]) {
        selectedSpeakerId = outputs[0].deviceId;
      }
      if (!speakerSelectable && outputs.length === 0) {
        audioNote = 'This browser uses the system default speaker.';
      }
    } catch (err) {
      audioNote = friendlyMediaError(err);
    } finally {
      audioBusy = false;
    }
  }

  function onMicChange(id: string) {
    selectedMicId = id;
    storeInputDeviceId(id);
  }

  function onSpeakerChange(id: string) {
    selectedSpeakerId = id;
    storeOutputDeviceId(id);
  }

  function togglePanel(id: typeof openPanel) {
    openPanel = openPanel === id ? '' : id;
  }

  function stopAudioTest() {
    if (audioTestRaf) cancelAnimationFrame(audioTestRaf);
    audioTestRaf = 0;
    if (audioTestStream) {
      for (const t of audioTestStream.getTracks()) t.stop();
      audioTestStream = null;
    }
    audioTesting = false;
    micLevel = 0;
  }

  async function startAudioTest() {
    if (audioTesting) {
      stopAudioTest();
      audioNote = 'Audio test stopped.';
      return;
    }
    audioNote = '';
    audioBusy = true;
    try {
      await refreshAudioDevices(true);
      const constraints: MediaStreamConstraints = {
        audio: selectedMicId
          ? { deviceId: { ideal: selectedMicId }, echoCancellation: true, noiseSuppression: true }
          : { echoCancellation: true, noiseSuppression: true }
      };
      audioTestStream = await navigator.mediaDevices.getUserMedia(constraints);
      const ctx = new AudioContext();
      const source = ctx.createMediaStreamSource(audioTestStream);
      const analyser = ctx.createAnalyser();
      analyser.fftSize = 256;
      source.connect(analyser);
      const data = new Uint8Array(analyser.frequencyBinCount);
      audioTesting = true;
      audioNote = 'Testing microphone… speak now.';
      const tick = () => {
        analyser.getByteFrequencyData(data);
        let sum = 0;
        for (const v of data) sum += v;
        micLevel = Math.min(100, Math.round((sum / data.length / 255) * 140));
        audioTestRaf = requestAnimationFrame(tick);
      };
      tick();
      window.setTimeout(() => {
        if (audioTesting) {
          stopAudioTest();
          void ctx.close().catch(() => {});
          audioNote = 'Audio test finished. Mic and speaker look ready.';
        }
      }, 6000);
    } catch (err) {
      stopAudioTest();
      audioNote = friendlyMediaError(err);
    } finally {
      audioBusy = false;
    }
  }

  function selectedMicLabel() {
    return audioInputs.find((d) => d.deviceId === selectedMicId)?.label || 'Default microphone';
  }

  function selectedSpeakerLabel() {
    if (!speakerSelectable) return 'System default speaker';
    return audioOutputs.find((d) => d.deviceId === selectedSpeakerId)?.label || 'Default speaker';
  }

  const onlineLabel = $derived(
    systemLiveKind === 'ok' ? 'Online' : systemLiveKind === 'issues' ? 'Limited' : systemLiveKind === 'offline' ? 'Offline' : 'Checking…'
  );

  function agentInitial(name?: string) {
    return (name || 'A').slice(0, 1).toUpperCase();
  }

  function formatTimer(seconds: number) {
    return new Date(seconds * 1000).toISOString().slice(11, 19);
  }

  // Start call / chat / workforce require a signed-in customer session.
  // Policy still drives backend enforcement; UI always gates on session presence.
  const authRequired = $derived(!customer);
  const autoRegister = $derived(!!portalPolicy?.customer_auth?.auto_register_on_conversation_otp);
  const quotaExhausted = $derived(quota?.state === 'quota_exhausted');
  const quotaLabel = $derived(formatQuota(quota));

  async function refreshPortalState() {
    portalPolicy = await loadCustomerPortalPolicy(tenantId ? { tenantId } : undefined);
    if (!authRequired) {
      agents = await loadWorkforce(tenantId ? { tenantId } : undefined);
      selectedAgent =
        agents.find((a) => a.id === selectedAgent?.id) || agents.find((a) => a.popular) || agents[0] || null;
    } else {
      agents = [];
      selectedAgent = null;
    }
    try {
      quota = await loadCustomerQuota(tenantId ? { tenantId } : undefined);
    } catch {
      quota = portalPolicy.quota;
    }
  }

  function formatQuota(value: CustomerQuotaSummary | null) {
    if (!value) return 'quota loading';
    if (value.daily_limit_seconds <= 0 && value.max_call_seconds <= 0) return 'quota not capped';
    const parts: string[] = [];
    if (value.daily_limit_seconds > 0) {
      const remaining = value.daily_remaining_seconds ?? value.daily_limit_seconds;
      parts.push(`${Math.floor(remaining / 60)}m daily left`);
    }
    if (value.max_call_seconds > 0) parts.push(`${Math.floor(value.max_call_seconds / 60)}m max/call`);
    return parts.join(' · ');
  }

  function startTimer() {
    startedAt = Date.now();
    activeCallLimitSeconds = Math.max(0, quota?.max_call_seconds || 0);
    remainingSeconds = activeCallLimitSeconds;
    remainingTimer = formatTimer(activeCallLimitSeconds);
    clearInterval(timerId);
    clearTimeout(warningTimerId);
    clearTimeout(timeoutTimerId);
    timerId = setInterval(() => {
      const elapsed = Math.floor((Date.now() - startedAt) / 1000);
      timer = formatTimer(elapsed);
      if (activeCallLimitSeconds > 0) {
        remainingSeconds = Math.max(0, activeCallLimitSeconds - elapsed);
        remainingTimer = formatTimer(remainingSeconds);
      }
    }, 1000);
    if (activeCallLimitSeconds > 10) {
      warningTimerId = setTimeout(() => {
        if (!live || !voice) return;
        voiceState = 'This call will end in 10 seconds. Please finish your question.';
        voice.sendText(
          'System notice: this call will end in 10 seconds because the customer time limit is nearly reached. Tell the customer to finish their question and let them know they can rate this call from 1 to 5 after it ends.'
        );
      }, (activeCallLimitSeconds - 10) * 1000);
    } else if (activeCallLimitSeconds > 0) {
      warningTimerId = setTimeout(() => {
        if (!live || !voice) return;
        voiceState = 'This call will end soon. Please finish your question.';
        voice.sendText(
          'System notice: this call will end very soon because the customer time limit is nearly reached. Ask the customer to finish and prepare to rate this call from 1 to 5 after it ends.'
        );
      }, 0);
    }
    if (activeCallLimitSeconds > 0) {
      timeoutTimerId = setTimeout(() => void endActiveCall('timeout'), activeCallLimitSeconds * 1000);
    }
  }

  function stopTimer() {
    clearInterval(timerId);
    clearTimeout(warningTimerId);
    clearTimeout(timeoutTimerId);
    clearInterval(autoCloseTimerId);
    clearTimeout(customerCloseFallbackTimerId);
    timer = '00:00:00';
    remainingTimer = '00:00:00';
    remainingSeconds = 0;
    activeCallLimitSeconds = 0;
    autoCloseTimerId = undefined;
    customerCloseFallbackTimerId = undefined;
    customerEndRequested = false;
    autoClosePending = false;
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
    if (live) await endActiveCall();
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
    if (uiRole === 'assistant') {
      showTone(content);
      if (customerEndRequested && assistantConfirmedFarewell(content)) {
        startCustomerFinishedCountdown();
      }
    }
    const offer = uiRole === 'user' ? ticketOfferForText(content) : null;
    if (offer) {
      openTicketOffer(offer);
    }
  }

  function requestCustomerFinishedClose() {
    if (!live || !session || customerEndRequested || autoClosePending) return;
    customerEndRequested = true;
    voiceState = 'กำลังกล่าวลาและจะวางสายภายใน 5 วินาที';
    const sent = voice?.sendText(
      'The caller said there is nothing else and thanked you. Respond in Thai: "ขออนุญาตวางสายก่อนนะครับ ขอบคุณครับ". Do not ask another question. The call will close in five seconds.'
    );
    // Start even if output transcription is unavailable or the model does not
    // repeat the farewell exactly. Normal calls start from the assistant caption.
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
    voiceState = `ขออนุญาตวางสายก่อนนะครับ ปิดสายภายใน ${seconds} วินาที`;
    autoCloseTimerId = setInterval(() => {
      seconds -= 1;
      if (seconds <= 0) {
        clearInterval(autoCloseTimerId);
        autoCloseTimerId = undefined;
        // Use the same handler as the visible End call button so recording,
        // archive, session end, and rating all follow one path.
        void endActiveCall('customer_finished');
        return;
      }
      voiceState = `ขออนุญาตวางสายก่อนนะครับ ปิดสายภายใน ${seconds} วินาที`;
    }, 1000);
  }

  function ticketOfferForText(text: string): TicketOffer | null {
    const normalized = text.toLowerCase();
    const signals = ['human agent', 'live agent', 'real person', 'speak to a person', 'talk to someone', 'escalate', 'มนุษย์', 'เจ้าหน้าที่', 'คุยกับคน', 'ขอคน'];
    if (!signals.some((signal) => normalized.includes(signal))) return null;
    return {
      subject: 'Human follow-up requested',
      category: topic === 'billing' || topic === 'technical' ? topic : 'general',
      reason: `Customer context: ${boundedCustomerContext(text)}`
    };
  }

  function boundedCustomerContext(text: string) {
    const value = text.trim().replace(/\s+/g, ' ');
    return value.length > 500 ? `${value.slice(0, 500)}…` : value;
  }

  function openTicketOffer(offer: TicketOffer) {
    if (ticketOffer) return;
    ticketOffer = offer;
    ticketError = '';
    ticketContactEmail = customer?.email || '';
    ticketContactName = customer?.display_name || '';
  }

  function declineTicketOffer() {
    ticketOffer = null;
    ticketError = '';
  }

  async function confirmTicketOffer() {
    if (!ticketOffer || ticketBusy) return;
    const callId = session?.id || chatSessionId;
    if (!callId) {
      ticketError = 'Start a chat or call before requesting follow-up.';
      return;
    }
    if (!customer && !ticketContactEmail.trim()) {
      ticketError = 'Enter an email so the tenant team can contact you.';
      return;
    }
    ticketBusy = true;
    ticketError = '';
    try {
      const result = await createCustomerTicket(
        {
          call_id: callId,
          confirm_escalation: true,
          subject: ticketOffer.subject,
          description: ticketOffer.reason,
          category: ticketOffer.category,
          contact_name: ticketContactName.trim() || undefined,
          contact_email: ticketContactEmail.trim() || undefined
        },
        { tenantId: tenantId || undefined, idempotencyKey: `customer-escalation:${callId}` }
      );
      addMessage('assistant', `Your follow-up request is confirmed. Reference ${result.ticket.id}.`, agentInitial(selectedAgent?.name));
      ticketOffer = null;
    } catch (err) {
      ticketError = err instanceof Error ? err.message : 'Could not create follow-up ticket';
    } finally {
      ticketBusy = false;
    }
  }

  async function startCall() {
    if (!selectedAgent) {
      error = authRequired ? 'Sign in before selecting an AI agent.' : 'Select an AI agent first.';
      return;
    }
    if (authRequired || quotaExhausted) {
      error = authRequired ? 'Sign in with OTP before starting a call.' : 'Customer quota exhausted.';
      return;
    }
    error = '';
    busy = true;
    callControlsExpanded = false;
    transcriptKeys.clear();
    customerEndRequested = false;
    autoClosePending = false;
    voiceState = 'Connecting…';
    let gemini: GeminiVoice | undefined;
    try {
      // Establish the call session before opening Gemini. Final caller
      // transcription can arrive immediately after the voice socket is ready.
      const created = await createCall(
        tenantId ? { tenantId, agentId: selectedAgent.id } : { agentId: selectedAgent.id }
      );
      session = created;
      chatSessionId = created.id;

      gemini = new GeminiVoice();
      // Make the live client available to callbacks while start() is still
      // negotiating. Gemini can deliver the first caller transcript before
      // start() resolves its ready promise.
      voice = gemini;
      // Show greeting text immediately while audio path connects.
      if (selectedAgent.greeting) {
        upsertVoiceTurn('agent', selectedAgent.greeting);
      }
      await gemini.start(
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
            if (role === 'caller' && customerConfirmedEnd(text)) {
              requestCustomerFinishedClose();
            }
            // Persist only finalized turns (not every short partial fragment).
            if (meta?.final && session) void persistTurn(session.id, role, text);
          },
          onCustomerEndRequested: () => requestCustomerFinishedClose(),
          onError: (message) => {
            error = message;
          }
        },
        {
          lang: 'auto',
          tenantId: tenantId || undefined,
          audioInputId: selectedMicId || undefined,
          audioOutputId: selectedSpeakerId || undefined
        }
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
        tenantId ? { tenantId } : undefined
      );

      live = true;
      startTimer();
      if (!voiceState.startsWith('On call')) {
        voiceState = `On call with ${selectedAgent.name} — agent greets first.`;
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Call failed';
      await gemini?.stop().catch(() => {});
      if (session) {
        await endCall(session.id, tenantId ? { tenantId } : undefined).catch(() => {});
      }
      await cleanup(true);
    } finally {
      busy = false;
    }
  }

  async function endActiveCall(reason: 'manual' | 'timeout' | 'customer_finished' = 'manual') {
    if (!session) return;
    const endedCallId = session.id;
    busy = true;
    try {
      const recordings = await voice?.stop();
      if (recordings && recordings.length > 0) {
        await archiveCallAudio(session.id, recordings, tenantId ? { tenantId } : undefined).catch((err) => {
          error = err instanceof Error ? err.message : 'Failed to archive call audio';
        });
      }
      await endCall(session.id, tenantId ? { tenantId } : undefined);
      void loadCustomerQuota(tenantId ? { tenantId } : undefined).then((q) => (quota = q)).catch(() => {});
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to end call';
    } finally {
      await cleanup(true);
      busy = false;
      ratingCallId = endedCallId;
      ratingScore = 0;
      ratingError = '';
      ratingOpen = true;
      if (reason === 'timeout') {
        voiceState = 'The call ended because the customer time limit was reached. Please rate the call.';
      } else if (reason === 'customer_finished') {
        voiceState = 'The call ended. Please rate the call.';
      }
    }
  }

  async function submitRating(event: Event) {
    event.preventDefault();
    if (!ratingCallId || ratingScore < 1 || ratingScore > 5 || ratingBusy) return;
    ratingBusy = true;
    ratingError = '';
    try {
      await submitCallRating(
        ratingCallId,
        { score: ratingScore },
        tenantId ? { tenantId } : undefined
      );
      ratingOpen = false;
    } catch (err) {
      ratingError = err instanceof Error ? err.message : 'Failed to save rating';
    } finally {
      ratingBusy = false;
    }
  }

  function finishChat() {
    if (!chatSessionId || live || ratingBusy) return;
    ratingCallId = chatSessionId;
    ratingScore = 0;
    ratingError = '';
    ratingOpen = true;
  }

  async function cleanup(resetSession: boolean) {
    live = false;
    stopTimer();
    unsubscribe?.();
    unsubscribe = undefined;
    voice = null;
    clearTimeout(customerCloseFallbackTimerId);
    customerCloseFallbackTimerId = undefined;
    customerEndRequested = false;
    if (resetSession) session = null;
    callControlsExpanded = false;
    voiceState = selectedAgent
      ? `Ready to call ${selectedAgent.name}.`
      : 'Select an agent, then start an inbound voice call.';
  }

  async function toggleCall() {
    if (live) await endActiveCall();
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
      const willRegister = res.customer_hint?.will_auto_register;
      authStatus = willRegister
        ? `OTP sent to ${res.delivery?.to || customerEmail} · will create account on verify`
        : `OTP sent to ${res.delivery?.to || customerEmail}`;
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
      await refreshPortalState();
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
    await refreshPortalState();
  }

  async function submitChat(event: Event) {
    event.preventDefault();
    if (!selectedAgent) {
      error = authRequired ? 'Sign in before selecting an AI agent.' : 'Select an AI agent first.';
      return;
    }
    if (authRequired || quotaExhausted) {
      error = authRequired ? 'Sign in with OTP before starting chat.' : 'Customer quota exhausted.';
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
      if (data.ticket_offer) openTicketOffer(data.ticket_offer);
      messages = messages.map((m) =>
        m.id === thinking.id
          ? { ...m, content: data.reply, sources: data.sources, missingKm: data.missing_km }
          : m
      );
      chatHistory = [...chatHistory, { role: 'assistant', content: data.reply }];
      showTone(data.reply);
      void loadCustomerQuota(tenantId ? { tenantId } : undefined).then((q) => (quota = q)).catch(() => {});
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
  const callStarted = $derived(live || !!session);
  const customerLabel = $derived(customer?.display_name || customer?.email || 'Customer');
  const canOpenPicker = $derived(!live && !authRequired && !quotaExhausted && agents.length > 0);
  const showCallDetails = $derived(!callStarted || callControlsExpanded);
  // Hide agent picker, Start call, and orb until the customer is signed in.
  const hideAgentSurfaceBeforeLogin = $derived(authRequired && !callStarted);
  const callTimerLabel = $derived(activeCallLimitSeconds > 0 ? remainingTimer : timer);
  const callTimerWarning = $derived(activeCallLimitSeconds > 0 && remainingSeconds <= 10);
</script>

<main class="app">
  <aside
    class="panel control-panel"
    class:live-collapsed={callStarted && !callControlsExpanded}
    class:live-expanded={callStarted && callControlsExpanded}
  >
    <!-- 1. Monti brand header -->
    <header class="desk-branding">
      <div class="monti-desk-hero">
        <div class="monti-desk-ring">
          <img class="monti-desk-logo" src="/images/monti-logo.png" width="96" height="96" alt="Monti" />
        </div>
        <div class="monti-desk-copy">
          <div class="monti-title-row">
            <span class="monti-desk-title">MONTI</span>
            <span
              class="online-pill"
              class:ok={systemLiveKind === 'ok'}
              class:issues={systemLiveKind === 'issues'}
              class:offline={systemLiveKind === 'offline'}
            >
              <i></i>
              {onlineLabel}
            </span>
          </div>
          <span class="monti-desk-tag">AI CALL CENTER</span>
          <p class="monti-desk-tagline">Your AI assistant. <em>Always here to help.</em></p>
        </div>
      </div>

      <!-- 2. Selected tenant -->
      <section class="company-card" aria-label="Selected brand">
        <div class="company-card-top">
          <span class="company-card-label">Selected tenant</span>
          {#if onChangeTenant && !live}
            <button class="link-btn" type="button" onclick={() => onChangeTenant?.()}>
              Change tenant ⌃
            </button>
          {/if}
        </div>
        <div class="company-card-row">
          <div class="company-logo">
            {#if brand.logo_url && !brand.logo_url.includes('monti-logo')}
              <img src={brand.logo_url} alt={brand.logo_alt || tenantLabel} />
            {:else}
              <span class="company-monogram">{companyMonogram}</span>
            {/if}
          </div>
          <div class="company-meta">
            <strong>{tenantName || brand.brand_name || tenantLabel}</strong>
            <span>AI · Text &amp; Voice</span>
            {#if tenantSlug || tenantLabel}
              <span class="company-brand-line">Brand · {tenantName || tenantLabel}</span>
            {/if}
            <span class="active-pill">Active</span>
          </div>
        </div>
      </section>
    </header>

    {#if callStarted}
      <section class="live-call-strip" aria-label="Active call controls">
        <div class="live-call-summary" style="--assistant-color:{selectedAgent?.color || 'var(--cyan)'}">
          <div class="agent-dot">{agentInitial(selectedAgent?.name)}</div>
          <div>
            <strong>{selectedAgent?.name || 'Agent'}</strong>
            <span>{callTimerLabel} · {customerLabel}</span>
          </div>
        </div>
        <button class="strip-button" type="button" onclick={() => (callControlsExpanded = !callControlsExpanded)}>
          {callControlsExpanded ? 'Collapse' : 'Expand'}
        </button>
        <button class="strip-button end" type="button" disabled={busy} onclick={() => void endActiveCall()}>End</button>
      </section>
    {/if}

    {#if showCallDetails}
      <!-- 3. Quick actions -->
      <nav class="quick-actions" aria-label="Quick actions">
        <button type="button" class="qa-btn" disabled={!onChangeTenant || live} onclick={() => onChangeTenant?.()}>
          <span class="qa-ico" aria-hidden="true">🏢</span>
          <span>My tenants</span>
        </button>
        <button
          type="button"
          class="qa-btn"
          onclick={() => {
            audioOpen = true;
            void startAudioTest();
          }}
        >
          <span class="qa-ico" aria-hidden="true">🎧</span>
          <span>Audio test</span>
        </button>
        <button
          type="button"
          class="qa-btn"
          onclick={() => {
            const el = document.getElementById('desk-account');
            el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
          }}
        >
          <span class="qa-ico" aria-hidden="true">👤</span>
          <span>Profile</span>
        </button>
        <button type="button" class="qa-btn" onclick={() => (audioOpen = !audioOpen)}>
          <span class="qa-ico" aria-hidden="true">⚙️</span>
          <span>Settings</span>
        </button>
      </nav>

      <!-- 4. Audio settings (collapsible) -->
      <section class="voice-card audio-card" aria-label="Audio settings">
        <button type="button" class="collapse-head" onclick={() => (audioOpen = !audioOpen)} aria-expanded={audioOpen}>
          <div>
            <strong>🔊 Audio settings</strong>
            <p>Configure your microphone and speaker</p>
          </div>
          <span class="chev">{audioOpen ? '⌃' : '⌄'}</span>
        </button>
        {#if audioOpen}
          <div class="audio-body">
            <label class="audio-field audio-device-row">
              <span class="audio-device-label">
                <span class="dev-ico" aria-hidden="true">🎤</span>
                <span>
                  <b>Microphone</b>
                  <small>{selectedMicLabel()}</small>
                </span>
                <span class="level-bars" aria-hidden="true">
                  {#each [0, 1, 2, 3, 4] as i}
                    <i class:on={micLevel > i * 18}></i>
                  {/each}
                </span>
              </span>
              <select
                value={selectedMicId}
                disabled={audioBusy || live}
                onchange={(e) => onMicChange((e.currentTarget as HTMLSelectElement).value)}
              >
                {#if audioInputs.length === 0}
                  <option value="">Default microphone</option>
                {:else}
                  {#each audioInputs as dev (dev.deviceId)}
                    <option value={dev.deviceId}>{dev.label}</option>
                  {/each}
                {/if}
              </select>
            </label>

            <label class="audio-field audio-device-row">
              <span class="audio-device-label">
                <span class="dev-ico" aria-hidden="true">🔈</span>
                <span>
                  <b>Speaker</b>
                  <small>{selectedSpeakerLabel()}</small>
                </span>
                <span class="level-bars idle" aria-hidden="true">
                  {#each [0, 1, 2, 3, 4] as i}
                    <i class:on={i < 3}></i>
                  {/each}
                </span>
              </span>
              <select
                value={selectedSpeakerId}
                disabled={audioBusy || !speakerSelectable || live}
                onchange={(e) => onSpeakerChange((e.currentTarget as HTMLSelectElement).value)}
              >
                {#if !speakerSelectable}
                  <option value="">System default speaker</option>
                {:else if audioOutputs.length === 0}
                  <option value="">Default speaker</option>
                {:else}
                  {#each audioOutputs as dev (dev.deviceId)}
                    <option value={dev.deviceId}>{dev.label}</option>
                  {/each}
                {/if}
              </select>
            </label>

            <div class="audio-test-row">
              <div>
                <strong>Test your audio</strong>
                <p>Make sure your mic and speaker work properly.</p>
              </div>
              <button class="voice-button test-btn" type="button" disabled={audioBusy || live} onclick={() => void startAudioTest()}>
                {audioTesting ? 'Stop test' : 'Start test'}
              </button>
            </div>
            {#if audioNote}
              <div class="voice-state">{audioNote}</div>
            {/if}
          </div>
        {/if}
      </section>

      <!-- 5. Collapsible utility sections -->
      <div class="util-stack">
        <button type="button" class="util-row" onclick={() => togglePanel('notifications')} aria-expanded={openPanel === 'notifications'}>
          <span>🔔 Notifications</span>
          <span class="chev">{openPanel === 'notifications' ? '⌃' : '⌄'}</span>
        </button>
        {#if openPanel === 'notifications'}
          <div class="util-body">Alert preferences stay on your device for this session. Email alerts follow tenant policy after sign-in.</div>
        {/if}

        <button type="button" class="util-row" onclick={() => togglePanel('language')} aria-expanded={openPanel === 'language'}>
          <span>🌐 Language</span>
          <span class="chev">{openPanel === 'language' ? '⌃' : '⌄'}</span>
        </button>
        {#if openPanel === 'language'}
          <div class="util-body">Conversation language follows the agent and topic. UI labels support EN · TH.</div>
        {/if}

        <button type="button" class="util-row" onclick={() => togglePanel('security')} aria-expanded={openPanel === 'security'}>
          <span>🛡️ Security</span>
          <span class="chev">{openPanel === 'security' ? '⌃' : '⌄'}</span>
        </button>
        {#if openPanel === 'security'}
          <div class="util-body">Your call data is tenant-scoped and encrypted in transit. Sign out clears this browser session.</div>
        {/if}

        <button type="button" class="util-row" onclick={() => togglePanel('about')} aria-expanded={openPanel === 'about'}>
          <span>ℹ️ About</span>
          <span class="chev">{openPanel === 'about' ? '⌃' : '⌄'}</span>
        </button>
        {#if openPanel === 'about'}
          <div class="util-body">Monti AI Call Center{appVersion ? ` · ${appVersion}` : ''}. Multi-tenant inbound chat and voice.</div>
        {/if}
      </div>

      <!-- 6. Account / sign-in -->
      <section id="desk-account" class={`voice-card auth-card ${callStarted ? 'auth-card-compact' : ''}`} aria-label="Customer account">
        {#if customer}
          <div class="account-card">
            <div class="account-avatar">{customerLabel.slice(0, 1).toUpperCase()}</div>
            <div class="account-meta">
              <div class="account-name-row">
                <strong>{customerLabel}</strong>
                <span class="agent-badge">Customer</span>
              </div>
              <div class="voice-state customer-meta">{customer.email}</div>
              <div class="voice-state customer-meta">Signed in</div>
            </div>
            <button class="voice-button signout-button" type="button" onclick={signOutCustomer}>Sign out</button>
          </div>
        {:else}
          <form onsubmit={challengeId ? verifyOTP : sendOTP} style="display:grid;gap:10px">
            <div class="voice-state">
              {#if autoRegister}
                Enter email + OTP to sign in (new customers are registered automatically for this brand).
              {:else}
                Sign in required before starting a call or chat.
              {/if}
            </div>
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
        {#if authStatus && !callStarted}
          <div class="voice-state" style="margin-top:8px">{authStatus}</div>
        {/if}
      </section>

      <!-- 7. Footer -->
      <footer class="desk-footer">
        <span>🛡️ Your data is encrypted and secure.</span>
        <span>{appVersion || 'Monti'}</span>
      </footer>
    {/if}

    {#if authRequired}
      <section class="voice-card auth-required-card">
        <strong>Sign in required</strong>
        <div class="voice-state">Verify your email OTP to unlock AI agents and Start call.</div>
      </section>
    {/if}

    {#if showCallDetails && !hideAgentSurfaceBeforeLogin}
      <section class="voice-card call-card">
        <div class="agent-select-row">
          {#if selectedAgent}
            <div class="selected-agent" style="--assistant-color:{selectedAgent.color}">
              <div class="agent-dot">{agentInitial(selectedAgent.name)}</div>
              <div class="selected-agent-copy">
                <strong>{selectedAgent.name}</strong>
                <span>{selectedAgent.role}</span>
              </div>
            </div>
            <button
              class="picker-trigger picker-trigger-compact"
              type="button"
              disabled={!canOpenPicker}
              onclick={() => (pickerOpen = true)}
            >
              Change avatar
            </button>
          {:else}
            <button
              class="picker-trigger picker-trigger-wide"
              type="button"
              disabled={!canOpenPicker}
              onclick={() => (pickerOpen = true)}
            >
              Choose avatar
            </button>
          {/if}
        </div>
        <div class="voice-row">
          <div class="status-pill" class:warning={callTimerWarning}>
            {callTimerLabel}
          </div>
          <button
            class="voice-button"
            class:live={live}
            type="button"
            disabled={busy || authRequired || quotaExhausted}
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
        <div class="voice-state">Quota · {quotaLabel}</div>
      </section>
    {/if}

    {#if selectedAgent && !hideAgentSurfaceBeforeLogin}
      <section class="assistant-orb">
        <div class="halo" style="--assistant-color:{selectedAgent.color}">
          <Portrait agent={selectedAgent} speaking={live} {tone} />
        </div>
        <div class="assistant-name">
          <h2>{selectedAgent.name}</h2>
          <p>{selectedAgent.role} · {selectedAgent.trait}</p>
        </div>
        <Waveform color={selectedAgent.color} />
      </section>
    {/if}
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
            <p>Choose who will answer this customer session. Quota · {quotaLabel}</p>
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
      <div class="system-live" class:ok={systemLiveKind === 'ok'} class:issues={systemLiveKind === 'issues'} class:offline={systemLiveKind === 'offline'} aria-live="polite">
        <span class="system-live-dot" aria-hidden="true"></span>
        <span>{systemLive}</span>
      </div>
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
      {#if ticketOffer}
        <section class="escalation-offer" aria-live="polite">
          <div>
            <p class="eyebrow">Human follow-up</p>
            <h3>Would you like the tenant team to contact you?</h3>
            <p>{ticketOffer.reason}</p>
          </div>
          {#if !customer}
            <div class="ticket-contact">
              <input bind:value={ticketContactEmail} type="email" placeholder="Contact email" autocomplete="email" />
              <input bind:value={ticketContactName} placeholder="Name (optional)" autocomplete="name" />
            </div>
          {/if}
          {#if ticketError}<div class="error">{ticketError}</div>{/if}
          <div class="escalation-actions">
            <button class="send" type="button" disabled={ticketBusy} onclick={confirmTicketOffer}>{ticketBusy ? 'Creating…' : 'Request follow-up'}</button>
            <button class="plain-button" type="button" disabled={ticketBusy} onclick={declineTicketOffer}>No thanks</button>
          </div>
        </section>
      {/if}
      <form onsubmit={submitChat}>
        <div class="composer">
          <textarea
            bind:value={input}
            placeholder={authRequired ? 'Sign in with OTP first...' : quotaExhausted ? 'Customer quota exhausted' : 'Ask your question...'}
            autocomplete="off"
            disabled={busy || authRequired || quotaExhausted}
            onkeydown={handleKeydown}
          ></textarea>
          <button class="send" type="submit" disabled={busy || authRequired || quotaExhausted}>Send</button>
        </div>
        {#if chatSessionId && !live}
          <button class="plain-button finish-chat" type="button" onclick={finishChat}>Finish chat &amp; rate</button>
        {/if}
        <div class="error">{error}</div>
      </form>
      <div class="infra">{sessionLabel}</div>
    </section>
  </section>
</main>

{#if ratingOpen}
  <div class="rating-backdrop">
    <div class="rating-dialog" role="dialog" aria-modal="true" aria-labelledby="rating-title">
      <div class="rating-kicker">Call complete</div>
      <h2 id="rating-title">How was your call?</h2>
      <p>Choose a score from 1 to 5 before closing this review.</p>
      <form onsubmit={submitRating}>
        <div class="rating-scale" role="radiogroup" aria-label="Call score">
          {#each [1, 2, 3, 4, 5] as score}
            <button
              type="button"
              class:active={ratingScore >= score}
              class="rating-score"
              aria-label={`${score} out of 5`}
              aria-pressed={ratingScore === score}
              onclick={() => (ratingScore = score)}
            >{ratingScore >= score ? '★' : '☆'}</button>
          {/each}
        </div>
        {#if ratingError}<div class="rating-error">{ratingError}</div>{/if}
        <button class="rating-submit" type="submit" disabled={ratingScore === 0 || ratingBusy}>
          {ratingBusy ? 'Saving…' : 'Submit review'}
        </button>
        <button class="rating-skip" type="button" disabled={ratingBusy} onclick={() => (ratingOpen = false)}>Not now</button>
      </form>
    </div>
  </div>
{/if}
