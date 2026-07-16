export type Locale = "en" | "th";

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface TokenStore {
  getAccessToken(): string | undefined;
  getRefreshToken(): string | undefined;
  setTokens(tokens: TokenPair): void;
  clear(): void;
}

export interface MobileAvatar {
  id: string;
  name: string;
  role: string;
  trait?: string;
  voice?: string;
  image?: string;
  image_url?: string;
  greeting?: string;
  status?: "active" | string;
  is_default?: boolean;
}

export interface MobileBootstrap {
  version: "v1";
  tenant: { id: string; display_name: string; slug: string };
  auth: {
    enabled: boolean;
    mode: "disabled" | "optional" | "required";
    require_auth_for_workforce: boolean;
    allow_public_no_auth: boolean;
    otp: {
      channel: "email" | string;
      ttl_seconds: number;
      resend_after_seconds: number;
    };
  };
  locale: {
    default: Locale;
    customer: Locale | string;
    ai_reply: Locale;
    language: Locale;
    supported: Locale[];
    timezone: string;
  };
  avatars: MobileAvatar[];
  default_avatar_id: string;
  limits: {
    max_call_seconds: number;
    daily_limit_seconds: number;
    daily_remaining_seconds: number | null;
    warning_at_seconds: number;
    reset_at: string;
    state: string;
  };
}

export interface MobileCall {
  call_id: string;
  status: "active" | "ended" | string;
  avatar_id: string;
  started_at: string;
  ended_at?: string;
}

export interface PublicBrand {
  id: string;
  slug: string;
  name: string;
  blurb?: string;
  logo_url?: string;
  category?: string;
  languages: Locale[];
  listed: boolean;
  status: "active" | string;
}

export interface MobileTurn {
  id: number;
  role: string;
  content: string;
  created_at: string;
}

export type MobileEvent =
  | { type: "status"; message?: string }
  | { type: "ready"; call_id?: string; avatar_id?: string; agent_id?: string; agent_name?: string; voice?: string; encoding?: string }
  | { type: "transcript"; role?: "user" | "assistant" | "caller"; text?: string; final?: boolean }
  | { type: "audio"; data?: string; data_base64?: string }
  | { type: "text"; text?: string }
  | { type: "turn_complete" }
  | { type: "call_status"; status?: "active" | "ending" | "ended" | "failed" | string; message?: string }
  | { type: "customer_end_requested" }
  | { type: "error"; message?: string; code?: string; retryable?: boolean }
  | Record<string, unknown>;

export interface MobileFetch {
  (input: RequestInfo | URL, init?: RequestInit): Promise<Response>;
}

export interface MobileSocket {
  readyState: number;
  onopen: ((event: unknown) => void) | null;
  onmessage: ((event: { data: string }) => void) | null;
  onerror: ((event: unknown) => void) | null;
  onclose: ((event: unknown) => void) | null;
  send(data: string): void;
  close(code?: number, reason?: string): void;
}

export type MobileSocketFactory = (url: string) => MobileSocket;

export interface MobileClientOptions {
  baseUrl: string;
  tenantId?: string;
  tokenStore: TokenStore;
  fetch?: MobileFetch;
  websocket?: MobileSocketFactory;
  sdkVersion?: string;
}

export interface OTPRequest {
  email: string;
  display_name?: string;
  locale?: Locale;
  notification?: { platform: "ios" | "android"; push_token: string; app_version?: string };
}

export interface CallOptions {
  avatarId?: string;
  locale?: Locale;
  idempotencyKey?: string;
  notification?: { platform: "ios" | "android"; push_token: string; app_version?: string };
}

export class MobileApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message = code) {
    super(message);
    this.name = "MobileApiError";
    this.status = status;
    this.code = code;
  }
}

export class CallHandle {
  private listeners = new Set<(event: MobileEvent) => void>();

  constructor(
    readonly call: MobileCall,
    private readonly socket: MobileSocket | undefined,
    private readonly client: MontiMobileClient,
  ) {
    if (!socket) return;
    socket.onmessage = (event) => {
      try {
        const parsed = JSON.parse(event.data) as MobileEvent;
        this.listeners.forEach((listener) => listener(parsed));
      } catch {
        this.listeners.forEach((listener) => listener({ type: "error", code: "invalid_event" }));
      }
    };
  }

  onEvent(listener: (event: MobileEvent) => void): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  sendAudio(base64Pcm16: string): void {
    this.socket?.send(JSON.stringify({ type: "audio", data_base64: base64Pcm16 }));
  }

  startAudio(sampleRate = 16000, channels = 1): void {
    this.socket?.send(JSON.stringify({ type: "start_audio", sample_rate: sampleRate, channels, encoding: "pcm_s16le" }));
  }

  sendText(text: string): void {
    this.socket?.send(JSON.stringify({ type: "text", text }));
  }

  closeTransport(): void {
    this.socket?.close(1000, "client_closed");
  }

  async end(): Promise<MobileCall> {
    this.socket?.send(JSON.stringify({ type: "end" }));
    this.socket?.close(1000, "call_ended");
    return this.client.endCall(this.call.call_id);
  }

  async rate(score: 1 | 2 | 3 | 4 | 5, review = ""): Promise<void> {
    await this.client.rateCall(this.call.call_id, score, review);
  }

  async reconnect(): Promise<CallHandle> {
    return this.client.connectCall(this.call.call_id);
  }
}

export class MontiMobileClient {
  private readonly fetcher: MobileFetch;
  private readonly socketFactory?: MobileSocketFactory;

  constructor(private readonly options: MobileClientOptions) {
    this.fetcher = options.fetch ?? globalThis.fetch.bind(globalThis);
    this.socketFactory = options.websocket;
  }

  async requestOTP(request: OTPRequest): Promise<Record<string, unknown>> {
    return this.request("/api/customer/auth/request-otp", {
      method: "POST",
      body: JSON.stringify({ ...request, tenant_id: this.options.tenantId }),
      auth: false,
    });
  }

  async verifyOTP(challengeId: string, otp: string): Promise<TokenPair> {
    const result = await this.request<TokenPair>("/api/customer/auth/verify-otp", {
      method: "POST",
      body: JSON.stringify({ tenant_id: this.options.tenantId, challenge_id: challengeId, otp }),
      auth: false,
    });
    this.options.tokenStore.setTokens(result);
    return result;
  }

  async refreshToken(): Promise<TokenPair> {
    const refreshToken = this.options.tokenStore.getRefreshToken();
    if (!refreshToken) throw new MobileApiError(401, "refresh_token_missing");
    try {
      const result = await this.request<TokenPair>("/api/customer/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
        auth: false,
      });
      this.options.tokenStore.setTokens(result);
      return result;
    } catch (error) {
      this.options.tokenStore.clear();
      throw error;
    }
  }

  async getBootstrap(): Promise<MobileBootstrap> {
    return this.request<MobileBootstrap>("/api/mobile/v1/bootstrap");
  }

  async getBrands(query = "", limit = 50, offset = 0): Promise<{ items: PublicBrand[]; total: number; limit: number; offset: number }> {
    const params = new URLSearchParams({ q: query, limit: String(limit), offset: String(offset) });
    return this.request(`/api/public/brands?${params.toString()}`, { auth: false });
  }

  async getBrand(slug: string): Promise<{ item: PublicBrand }> {
    return this.request<{ item: PublicBrand }>(`/api/public/brands/${encodeURIComponent(slug)}`, { auth: false });
  }

  async createCall(options: CallOptions = {}): Promise<MobileCall> {
    const result = await this.request<{ call_id: string; status: string; avatar_id: string; started_at: string }>("/api/mobile/v1/calls", {
      method: "POST",
      body: JSON.stringify({ avatar_id: options.avatarId, locale: options.locale, notification: options.notification }),
      headers: { "Idempotency-Key": options.idempotencyKey ?? randomId() },
    });
    return result;
  }

  async getCall(callId: string): Promise<MobileCall> {
    return this.request<MobileCall>(`/api/mobile/v1/calls/${encodeURIComponent(callId)}`);
  }

  async getTranscript(callId: string): Promise<{ turns: MobileTurn[] }> {
    return this.request<{ turns: MobileTurn[] }>(`/api/mobile/v1/calls/${encodeURIComponent(callId)}/transcript`);
  }

  async connectCall(callId: string): Promise<CallHandle> {
    if (!this.socketFactory) throw new MobileApiError(0, "websocket_adapter_missing");
    const call = await this.getCall(callId);
    const token = this.options.tokenStore.getAccessToken();
    const url = new URL(`/ws/mobile/v1/calls/${encodeURIComponent(callId)}`, this.options.baseUrl);
    if (token) url.searchParams.set("access_token", token);
    const socket = this.socketFactory(toWebSocketURL(url).toString());
    return new CallHandle(call, socket, this);
  }

  async endCall(callId: string, idempotencyKey = randomId()): Promise<MobileCall> {
    return this.request<MobileCall>(`/api/mobile/v1/calls/${encodeURIComponent(callId)}/end`, {
      method: "POST",
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  async rateCall(callId: string, score: 1 | 2 | 3 | 4 | 5, review = "", idempotencyKey = randomId()): Promise<void> {
    await this.request(`/api/mobile/v1/calls/${encodeURIComponent(callId)}/rating`, {
      method: "POST",
      body: JSON.stringify({ score, review }),
      headers: { "Idempotency-Key": idempotencyKey },
    });
  }

  private async request<T = Record<string, unknown>>(path: string, init: RequestInit & { auth?: boolean } = {}): Promise<T> {
    const headers = new Headers(init.headers);
    headers.set("Content-Type", "application/json");
    headers.set("X-Monti-SDK-Version", this.options.sdkVersion ?? "0.1.0");
    if (this.options.tenantId) headers.set("X-Tenant-Id", this.options.tenantId);
    if (init.auth !== false) {
      const accessToken = this.options.tokenStore.getAccessToken();
      if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);
    }
    const response = await this.fetcher(new URL(path, this.options.baseUrl), { ...init, headers });
    const raw = await response.text();
    let payload: any = {};
    try { payload = raw ? JSON.parse(raw) : {}; } catch { payload = {}; }
    if (!response.ok) {
      const code = typeof payload.code === "string" ? payload.code : "mobile_request_failed";
      throw new MobileApiError(response.status, code, typeof payload.error === "string" ? payload.error : code);
    }
    return payload as T;
  }
}

function randomId(): string {
  const cryptoObject = globalThis.crypto as Crypto | undefined;
  if (cryptoObject?.randomUUID) return cryptoObject.randomUUID();
  return `mobile-${Date.now()}-${Math.random().toString(36).slice(2)}`;
}

function toWebSocketURL(url: URL): URL {
  url.protocol = url.protocol === "https:" ? "wss:" : "ws:";
  return url;
}
