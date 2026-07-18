/** Floating launcher corner (matches S14 monti-embed.js). */
export type EmbedPosition = "bottom-right" | "bottom-left" | "top-right" | "top-left";

/** Public resolve payload from GET /api/public/embed/{key}. */
export interface EmbedResolveResult {
  tenant_id: string;
  slug: string;
  name: string;
  embed_key: string;
  enabled: boolean;
  default_agent_id?: string;
  agents?: Array<{ id: string; name: string; role?: string }>;
}

export type EmbedErrorCode =
  | "missing_embed_key"
  | "missing_api_base"
  | "embed_not_found"
  | "embed_disabled"
  | "origin_not_allowed"
  | "resolve_failed"
  | "network_error"
  | "not_browser"
  | "destroyed";

export interface EmbedError {
  code: EmbedErrorCode | string;
  message: string;
  status?: number;
}

export type EmbedEventMap = {
  open: void;
  close: void;
  ready: EmbedResolveResult | undefined;
  error: EmbedError;
  destroy: void;
};

export type EmbedEventName = keyof EmbedEventMap;

export interface EmbedOptions {
  /** Public embed key (`emb_…`) from tenant Embed settings. */
  embedKey: string;
  /** Monti host origin, e.g. `https://monti.example.com` or `http://localhost:8091`. */
  apiBase: string;
  /**
   * Host-site origin for allowlist checks.
   * Defaults to `window.location.origin` in the browser.
   */
  parentOrigin?: string;
  /** Floating launcher position. Default `bottom-right`. Ignored when `container` is set (inline). */
  position?: EmbedPosition;
  /** Optional default workforce agent id. Passed as `agent` query param when set. */
  agentId?: string;
  /** Optional theme hint (forwarded as `theme` query param when set). */
  theme?: string;
  /** Optional locale hint (forwarded as `locale` query param when set). */
  locale?: string;
  /** Start with panel open. Default false. */
  open?: boolean;
  /**
   * When set, mount the iframe panel inside this element (inline embed).
   * When omitted, use floating launcher + fixed panel (S14 UX).
   */
  container?: HTMLElement | null;
  /** Skip pre-resolve (iframe still validates). Default false — resolve for clearer errors. */
  skipResolve?: boolean;
  /** Custom fetch (tests). Defaults to global fetch. */
  fetch?: typeof fetch;
  /** z-index base for floating UI. Default 2147483000. */
  zIndex?: number;
  onOpen?: () => void;
  onClose?: () => void;
  onReady?: (result?: EmbedResolveResult) => void;
  onError?: (error: EmbedError) => void;
  onDestroy?: () => void;
}
