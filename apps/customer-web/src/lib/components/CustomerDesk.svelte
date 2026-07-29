<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { listPublicBrands, type PublicBrand } from '$lib/api/brands';
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
  let speakerLevel = $state(0);
  let tenantItems = $state<PublicBrand[]>([]);
  let appVersion = $state('');
  let signedInAt = $state<number | null>(null);
  let lastActiveAt = $state<number | null>(null);
  let audioTestStream: MediaStream | null = null;
  let audioTestRaf = 0;
  let audioTestCtx: AudioContext | null = null;
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
      if (profile) {
        signedInAt = Date.now();
        lastActiveAt = Date.now();
      }
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
    void loadTenantSwitcher();
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

  async function loadTenantSwitcher() {
    try {
      const res = await listPublicBrands({ limit: 8 });
      tenantItems = res.items;
      const selected = res.items.find((item) => item.id === tenantId || item.slug === tenantSlug);
      if (
        selected?.logo_url &&
        (!brand.logo_url || brand.logo_url.includes('monti-logo'))
      ) {
        brand = {
          ...brand,
          brand_name: selected.name || brand.brand_name,
          logo_url: selected.logo_url,
          logo_alt: selected.name || brand.logo_alt
        };
      }
    } catch {
      tenantItems = [];
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

  /** Open audio panel and load devices so mic/speaker can actually be chosen. */
  async function openAudioSettings(requestPermission = true) {
    audioOpen = true;
    await tick();
    const el = document.getElementById('desk-audio-settings');
    el?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    await refreshAudioDevices(requestPermission);
  }

  function stopAudioTest() {
    if (audioTestRaf) cancelAnimationFrame(audioTestRaf);
    audioTestRaf = 0;
    if (audioTestStream) {
      for (const t of audioTestStream.getTracks()) t.stop();
      audioTestStream = null;
    }
    if (audioTestCtx) {
      void audioTestCtx.close().catch(() => {});
      audioTestCtx = null;
    }
    audioTesting = false;
    micLevel = 0;
    speakerLevel = 0;
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
      audioTestCtx = ctx;
      const source = ctx.createMediaStreamSource(audioTestStream);
      const analyser = ctx.createAnalyser();
      analyser.fftSize = 256;
      source.connect(analyser);

      // Soft tone on selected speaker path when setSinkId is available.
      const osc = ctx.createOscillator();
      const gain = ctx.createGain();
      gain.gain.value = 0.04;
      osc.frequency.value = 660;
      osc.connect(gain);
      gain.connect(ctx.destination);
      const sink = ctx as AudioContext & { setSinkId?: (id: string) => Promise<void> };
      if (selectedSpeakerId && typeof sink.setSinkId === 'function') {
        try {
          await sink.setSinkId(selectedSpeakerId);
        } catch {
          /* default output */
        }
      }
      osc.start();
      window.setTimeout(() => {
        try {
          osc.stop();
        } catch {
          /* already stopped */
        }
      }, 450);

      const data = new Uint8Array(analyser.frequencyBinCount);
      audioTesting = true;
      audioNote = 'Testing… speak into the mic. A short tone played on the speaker.';
      const loop = () => {
        analyser.getByteFrequencyData(data);
        let sum = 0;
        for (const v of data) sum += v;
        const level = Math.min(100, Math.round((sum / data.length / 255) * 140));
        micLevel = level;
        speakerLevel = Math.max(level * 0.65, audioTesting ? 25 : 0);
        audioTestRaf = requestAnimationFrame(loop);
      };
      loop();
      window.setTimeout(() => {
        if (audioTesting) {
          stopAudioTest();
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

  function touchActivity() {
    lastActiveAt = Date.now();
  }

  function formatLastActive(ts: number | null): string {
    if (!ts) return 'just now';
    const sec = Math.max(0, Math.floor((Date.now() - ts) / 1000));
    if (sec < 45) return 'just now';
    if (sec < 90) return '1 min ago';
    if (sec < 3600) return `${Math.floor(sec / 60)} mins ago`;
    if (sec < 7200) return '1 hour ago';
    return `${Math.floor(sec / 3600)} hours ago`;
  }

  const onlineLabel = $derived(
    systemLiveKind === 'ok' ? 'Online' : systemLiveKind === 'issues' ? 'Limited' : systemLiveKind === 'offline' ? 'Offline' : 'Checking…'
  );
  const roleLabel = $derived(customer?.role === 'customer' ? 'Customer' : customer?.role || 'Guest');
  const lastActiveLabel = $derived(formatLastActive(lastActiveAt));

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

  async function callAgentNow(agent: Agent) {
    touchActivity();
    if (busy) return;
    if (selectedAgent?.id !== agent.id) {
      await selectAgent(agent);
      await tick();
    }
    if (live && selectedAgent?.id === agent.id) {
      await endActiveCall();
      return;
    }
    if (!live) await startCall();
  }

  function switchTenant(item: PublicBrand) {
    touchActivity();
    if (live) return;
    const slug = item.slug || item.id;
    if (!slug || slug === tenantSlug || item.id === tenantId) return;
    window.location.href = `/t/${encodeURIComponent(slug)}`;
  }

  function openTenantDirectory() {
    touchActivity();
    if (live) return;
    if (onChangeTenant) {
      onChangeTenant();
      return;
    }
    window.location.href = '/';
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
      signedInAt = Date.now();
      lastActiveAt = Date.now();
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
    signedInAt = null;
    lastActiveAt = null;
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

    {#if callStarted && !callControlsExpanded && selectedAgent}
      <section class="voice-card avatar-call-section live-avatar-section" aria-label="Live avatar">
        <div class="section-title-row">
          <span class="section-ico sparkle" aria-hidden="true">✦</span>
          <div>
            <strong>AI AVATAR LIVE</strong>
            <p>{voiceState}</p>
          </div>
          <div class="status-pill compact" class:warning={callTimerWarning}>{callTimerLabel}</div>
        </div>

        <div
          class="avatar-live-stage"
          class:live={live}
          class:connecting={busy && !live}
          style="--assistant-color:{selectedAgent.color}"
          aria-label={`Live avatar ${selectedAgent.name}`}
        >
          <div class="avatar-live-visual" aria-hidden="true">
            <span class="avatar-pulse-ring ring-a"></span>
            <span class="avatar-pulse-ring ring-b"></span>
            <span class="avatar-pulse-ring ring-c"></span>
            <div class="avatar-live-halo">
              <Portrait agent={selectedAgent} speaking={live} {tone} />
            </div>
          </div>
          <div class="avatar-live-copy">
            <strong>{selectedAgent.name}</strong>
            <span>{selectedAgent.role} · {selectedAgent.trait}</span>
          </div>
          <Waveform color={selectedAgent.color} count={34} />
          <div class="avatar-live-state">
            {live ? 'Speaking now' : busy ? 'Connecting...' : 'Ready to call'}
          </div>
        </div>
      </section>
    {/if}

    {#if showCallDetails}
      <!-- 3. Audio settings (collapsible) -->
      <section id="desk-audio-settings" class="voice-card audio-card" aria-label="Audio settings">
        <button
          type="button"
          class="collapse-head"
          onclick={() => {
            if (audioOpen) audioOpen = false;
            else void openAudioSettings(true);
          }}
          aria-expanded={audioOpen}
        >
          <div class="collapse-title">
            <span class="collapse-ico" aria-hidden="true">
              <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M3 10v4h4l5 5V5L7 10H3zm13.5 2a3.5 3.5 0 0 0-1.8-3.1v6.2A3.5 3.5 0 0 0 16.5 12zM14 4.2v2.1a6 6 0 0 1 0 11.4v2.1a8 8 0 0 0 0-15.6z"/></svg>
            </span>
            <div>
              <strong>AUDIO SETTINGS</strong>
              <p>Configure your microphone and speaker</p>
            </div>
          </div>
          <span class="chev">{audioOpen ? '⌃' : '⌄'}</span>
        </button>
        {#if audioOpen}
          <div class="audio-body">
            <div class="audio-field audio-device-row">
              <div class="audio-device-label">
                <span class="dev-ico" aria-hidden="true">
                  <svg viewBox="0 0 24 24" width="16" height="16"><path fill="currentColor" d="M12 14a3 3 0 0 0 3-3V6a3 3 0 0 0-6 0v5a3 3 0 0 0 3 3zm5-3a5 5 0 0 1-10 0H5a7 7 0 0 0 6 6.9V21h2v-3.1A7 7 0 0 0 19 11h-2z"/></svg>
                </span>
                <span class="dev-copy">
                  <b>Microphone</b>
                  <small>{selectedMicLabel()}</small>
                </span>
                <span class="level-bars" aria-hidden="true" title="Live mic level">
                  {#each [0, 1, 2, 3, 4] as i}
                    <i class:on={micLevel > i * 18}></i>
                  {/each}
                </span>
              </div>
              <select
                bind:value={selectedMicId}
                disabled={live}
                aria-label="Microphone device"
                onchange={(e) => {
                  touchActivity();
                  onMicChange((e.currentTarget as HTMLSelectElement).value);
                }}
              >
                {#if audioInputs.length === 0}
                  <option value="">Default microphone</option>
                {:else}
                  {#each audioInputs as dev (dev.deviceId)}
                    <option value={dev.deviceId}>{dev.label}</option>
                  {/each}
                {/if}
              </select>
            </div>

            <div class="audio-field audio-device-row">
              <div class="audio-device-label">
                <span class="dev-ico" aria-hidden="true">
                  <svg viewBox="0 0 24 24" width="16" height="16"><path fill="currentColor" d="M3 10v4h4l5 5V5L7 10H3zm13.5 2a3.5 3.5 0 0 0-1.8-3.1v6.2A3.5 3.5 0 0 0 16.5 12z"/></svg>
                </span>
                <span class="dev-copy">
                  <b>Speaker</b>
                  <small>{selectedSpeakerLabel()}</small>
                </span>
                <span class="level-bars" class:idle={!audioTesting} aria-hidden="true" title="Speaker activity">
                  {#each [0, 1, 2, 3, 4] as i}
                    <i class:on={speakerLevel > i * 18 || (!audioTesting && i < 3)}></i>
                  {/each}
                </span>
              </div>
              <select
                bind:value={selectedSpeakerId}
                disabled={live || !speakerSelectable}
                aria-label="Speaker device"
                onchange={(e) => {
                  touchActivity();
                  onSpeakerChange((e.currentTarget as HTMLSelectElement).value);
                }}
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
              {#if !speakerSelectable}
                <p class="voice-state">Speaker selection needs Chrome/Edge (setSinkId). Mic selection still works.</p>
              {/if}
            </div>

            <div class="audio-actions">
              <button
                class="ghost-btn"
                type="button"
                disabled={audioBusy}
                onclick={() => {
                  touchActivity();
                  void refreshAudioDevices(true);
                }}
              >
                {audioBusy ? 'Refreshing…' : '↻ Refresh devices'}
              </button>
            </div>

            <div class="audio-test-row">
              <div class="audio-test-copy">
                <span class="dev-ico wave" aria-hidden="true">
                  <svg viewBox="0 0 24 24" width="16" height="16"><path fill="currentColor" d="M4 12h2v4H4v-4zm4-4h2v12H8V8zm4-3h2v18h-2V5zm4 5h2v8h-2v-8zm4-2h2v12h-2V8z"/></svg>
                </span>
                <div>
                  <strong>Test your audio</strong>
                  <p>Make sure your mic and speaker work properly.</p>
                </div>
              </div>
              <button
                class="voice-button test-btn"
                type="button"
                disabled={live}
                onclick={() => {
                  touchActivity();
                  void startAudioTest();
                }}
              >
                {audioTesting ? 'Stop test' : 'Start test'}
              </button>
            </div>
            {#if audioNote}
              <div class="voice-state">{audioNote}</div>
            {/if}
          </div>
        {/if}
      </section>

      <!-- 4. AI avatar call grid -->
      {#if !hideAgentSurfaceBeforeLogin}
        <section class="voice-card avatar-call-section" aria-label="AI avatar call">
          <div class="section-title-row">
            <span class="section-ico sparkle" aria-hidden="true">✦</span>
            <div>
              <strong>AI AVATAR CALL</strong>
              <p>Choose an AI avatar to start a conversation</p>
            </div>
            <div class="status-pill compact" class:warning={callTimerWarning}>{callTimerLabel}</div>
          </div>

          {#if agents.length === 0}
            <div class="voice-state">Sign in to show available AI avatars.</div>
          {:else}
            {#if selectedAgent}
              <div
                class="avatar-live-stage"
                class:live={live && selectedAgent}
                class:connecting={busy && !live}
                style="--assistant-color:{selectedAgent.color}"
                aria-label={`Selected avatar ${selectedAgent.name}`}
              >
                <div class="avatar-live-visual" aria-hidden="true">
                  <span class="avatar-pulse-ring ring-a"></span>
                  <span class="avatar-pulse-ring ring-b"></span>
                  <span class="avatar-pulse-ring ring-c"></span>
                  <div class="avatar-live-halo">
                    <Portrait agent={selectedAgent} speaking={live} {tone} />
                  </div>
                </div>
                <div class="avatar-live-copy">
                  <strong>{selectedAgent.name}</strong>
                  <span>{selectedAgent.role} · {selectedAgent.trait}</span>
                </div>
                <Waveform color={selectedAgent.color} count={30} />
                <div class="avatar-live-state">
                  {live ? 'Live now' : busy ? 'Connecting...' : 'Ready to call'}
                </div>
              </div>
            {/if}

            <div class="avatar-call-grid">
              {#each agents.slice(0, 4) as agent (agent.id)}
                <article
                  class="avatar-call-card"
                  class:active={selectedAgent?.id === agent.id}
                  style="--assistant-color:{agent.color}"
                >
                  {#if selectedAgent?.id === agent.id}
                    <span class="avatar-online-dot" aria-label="Selected avatar"></span>
                  {/if}
                  <button
                    class="avatar-select-button"
                    type="button"
                    disabled={live}
                    onclick={() => {
                      touchActivity();
                      void selectAgent(agent);
                    }}
                  >
                    <Portrait {agent} />
                    <strong>{agent.name}</strong>
                  </button>
                  <button
                    class="avatar-call-button"
                    class:active={selectedAgent?.id === agent.id}
                    class:live={live && selectedAgent?.id === agent.id}
                    type="button"
                    disabled={busy || authRequired || quotaExhausted}
                    onclick={() => void callAgentNow(agent)}
                  >
                    <svg viewBox="0 0 24 24" width="15" height="15" aria-hidden="true">
                      <path
                        fill="currentColor"
                        d="M6.6 10.8c1.4 2.8 3.8 5.1 6.6 6.6l2.2-2.2c.3-.3.7-.4 1.1-.3 1.2.4 2.5.6 3.8.6.6 0 1 .4 1 1V20c0 .6-.4 1-1 1C10.6 21 3 13.4 3 4c0-.6.4-1 1-1h3.5c.6 0 1 .4 1 1 0 1.3.2 2.6.6 3.8.1.4 0 .8-.3 1.1L6.6 10.8z"
                      />
                    </svg>
                    {live && selectedAgent?.id === agent.id
                      ? 'End call'
                      : selectedAgent?.id === agent.id
                        ? busy
                          ? 'Connecting…'
                          : 'Call now'
                        : 'Call'}
                  </button>
                </article>
              {/each}
            </div>
          {/if}

          {#if busy && !live}
            <div class="voice-state loading" aria-live="polite">⏳ {voiceState}</div>
          {:else}
            <div class="voice-state">{voiceState}</div>
          {/if}
          <div class="voice-state">Quota · {quotaLabel}</div>
        </section>
      {/if}

      <!-- 5. My tenants -->
      <section class="voice-card tenant-switcher-section" aria-label="My tenants">
        <div class="section-title-row tenant-title-row">
          <span class="section-ico" aria-hidden="true">
            <svg viewBox="0 0 24 24" width="18" height="18"><path fill="currentColor" d="M3 21V5l8-3v19H3zm10 0V8l8 3v10h-8zM6 8h2v2H6V8zm0 4h2v2H6v-2zm0 4h2v2H6v-2zm10-3h2v2h-2v-2zm0 4h2v2h-2v-2z"/></svg>
          </span>
          <div>
            <strong>MY TENANTS</strong>
            <p>Manage and switch between your tenants</p>
          </div>
          <button class="tenant-next" type="button" disabled={live} onclick={openTenantDirectory} aria-label="Open tenant directory">›</button>
        </div>

        <div class="tenant-switcher-grid">
          {#if tenantItems.length > 0}
            {#each tenantItems.slice(0, 4) as item (item.id)}
              <button
                type="button"
                class="tenant-mini-card"
                class:active={item.slug === tenantSlug || item.id === tenantId}
                disabled={live}
                onclick={() => switchTenant(item)}
              >
                <span class="tenant-mini-logo">
                  {#if item.logo_url && !item.logo_url.includes('monti-logo')}
                    <img src={item.logo_url} alt="" />
                  {:else}
                    <span>{brandMonogram(item.name, item.slug)}</span>
                  {/if}
                </span>
                <strong>{item.name || item.slug}</strong>
              </button>
            {/each}
          {:else}
            <button type="button" class="tenant-mini-card active" disabled>
              <span class="tenant-mini-logo">
                {#if brand.logo_url && !brand.logo_url.includes('monti-logo')}
                  <img src={brand.logo_url} alt="" />
                {:else}
                  <span>{companyMonogram}</span>
                {/if}
              </span>
              <strong>{tenantName || brand.brand_name || tenantLabel}</strong>
            </button>
          {/if}
          <button type="button" class="tenant-mini-card add" disabled={live} onclick={openTenantDirectory}>
            <span class="tenant-add-plus">＋</span>
            <strong>Add tenant</strong>
          </button>
        </div>
      </section>

      <!-- 6. User account -->
      <section id="desk-account" class={`voice-card auth-card ${callStarted ? 'auth-card-compact' : ''}`} aria-label="Customer account">
        {#if customer}
          <div class="account-card">
            <div class="account-avatar" aria-hidden="true" title={customer.display_name || customer.email}>
              {(customer.display_name || customer.email || 'U').slice(0, 1).toUpperCase()}
            </div>
            <div class="account-meta">
              <div class="account-name-row">
                <strong>{customer.display_name || customerLabel}</strong>
                <span class="agent-badge">{roleLabel}</span>
              </div>
              <div class="voice-state customer-meta">{customer.email}</div>
              <div class="voice-state customer-meta">
                Signed in · Last active <em>{lastActiveLabel}</em>
              </div>
            </div>
            <button class="voice-button signout-button" type="button" onclick={() => void signOutCustomer()}>
              Sign out
            </button>
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

      <!-- 6. Footer -->
      <footer class="desk-footer">
        <span class="footer-secure">
          <svg viewBox="0 0 24 24" width="14" height="14" aria-hidden="true"><path fill="currentColor" d="M12 2 4 5v6c0 5 3.4 9.7 8 11 4.6-1.3 8-6 8-11V5l-8-3zm0 10.9 4-2.3V7.1L12 9.4 8 7.1v3.5l4 2.3z"/></svg>
          Your data is encrypted and secure.
        </span>
        <span class="footer-version">{appVersion || 'Monti'}</span>
      </footer>
    {/if}

    {#if authRequired}
      <section class="voice-card auth-required-card">
        <strong>Sign in required</strong>
        <div class="voice-state">Verify your email OTP to unlock AI agents and Start call.</div>
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
