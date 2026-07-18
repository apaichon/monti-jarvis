import { buildEmbedIframeUrl, normalizeApiBase, resolveEmbed, toEmbedError } from "./resolve.js";
import { createFloatingWidget, createInlineWidget, warnInsecureMontiHost, type WidgetHandles } from "./widget.js";
import type {
  EmbedError,
  EmbedEventMap,
  EmbedEventName,
  EmbedOptions,
  EmbedPosition,
  EmbedResolveResult,
} from "./types.js";

export type {
  EmbedError,
  EmbedErrorCode,
  EmbedEventMap,
  EmbedEventName,
  EmbedOptions,
  EmbedPosition,
  EmbedResolveResult,
} from "./types.js";

export { buildEmbedIframeUrl, buildResolveUrl, normalizeApiBase, resolveEmbed } from "./resolve.js";

type Listener<K extends EmbedEventName> = (payload: EmbedEventMap[K]) => void;

/**
 * Programmatic Monti web embed (framework-agnostic).
 * Floating launcher by default; pass `container` for inline iframe.
 */
export class MontiEmbed {
  private readonly options: EmbedOptions;
  private readonly listeners = new Map<EmbedEventName, Set<Listener<EmbedEventName>>>();
  private widget: WidgetHandles | null = null;
  private openState = false;
  private destroyed = false;
  private resolveResult: EmbedResolveResult | undefined;
  private lastError: EmbedError | null = null;
  private mountPromise: Promise<void> | null = null;

  constructor(options: EmbedOptions) {
    this.options = options;
    if (options.onOpen) this.on("open", options.onOpen);
    if (options.onClose) this.on("close", options.onClose);
    if (options.onReady) this.on("ready", options.onReady);
    if (options.onError) this.on("error", options.onError);
    if (options.onDestroy) this.on("destroy", options.onDestroy);
  }

  get isOpen(): boolean {
    return this.openState;
  }

  get error(): EmbedError | null {
    return this.lastError;
  }

  get resolve(): EmbedResolveResult | undefined {
    return this.resolveResult;
  }

  get isDestroyed(): boolean {
    return this.destroyed;
  }

  /** Mount the widget (idempotent). Resolves after optional public resolve. */
  async mount(): Promise<void> {
    if (this.destroyed) {
      this.emitError(toEmbedError({ code: "destroyed", message: "Embed instance was destroyed" }));
      return;
    }
    if (this.widget) return;
    if (this.mountPromise) return this.mountPromise;

    this.mountPromise = this.doMount();
    try {
      await this.mountPromise;
    } finally {
      this.mountPromise = null;
    }
  }

  open(): void {
    if (this.destroyed || !this.widget) return;
    if (this.openState) return;
    this.openState = true;
    this.widget.setOpen(true);
    this.emit("open", undefined as void);
  }

  close(): void {
    if (this.destroyed || !this.widget) return;
    if (!this.openState) return;
    this.openState = false;
    this.widget.setOpen(false);
    this.emit("close", undefined as void);
  }

  toggle(): void {
    if (this.openState) this.close();
    else this.open();
  }

  /** Remove DOM, listeners, and iframe. Safe to call multiple times. */
  destroy(): void {
    if (this.destroyed) return;
    this.destroyed = true;
    this.openState = false;
    this.widget?.destroy();
    this.widget = null;
    this.emit("destroy", undefined as void);
    this.listeners.clear();
  }

  on<K extends EmbedEventName>(event: K, handler: Listener<K>): () => void {
    let set = this.listeners.get(event);
    if (!set) {
      set = new Set();
      this.listeners.set(event, set);
    }
    set.add(handler as Listener<EmbedEventName>);
    return () => {
      set!.delete(handler as Listener<EmbedEventName>);
    };
  }

  private async doMount(): Promise<void> {
    if (typeof document === "undefined" || typeof window === "undefined") {
      this.emitError(toEmbedError({ code: "not_browser", message: "MontiEmbed requires a browser environment" }));
      return;
    }

    const embedKey = (this.options.embedKey || "").trim();
    const apiBase = (this.options.apiBase || "").trim();
    if (!embedKey) {
      this.emitError(toEmbedError({ code: "missing_embed_key", message: "embedKey is required" }));
      return;
    }
    if (!apiBase) {
      this.emitError(toEmbedError({ code: "missing_api_base", message: "apiBase is required" }));
      return;
    }

    const parentOrigin =
      (this.options.parentOrigin || "").trim() ||
      (typeof window !== "undefined" ? window.location.origin : "");

    warnInsecureMontiHost(apiBase);

    if (!this.options.skipResolve) {
      try {
        this.resolveResult = await resolveEmbed({
          apiBase,
          embedKey,
          parentOrigin,
          fetch: this.options.fetch,
        });
      } catch (err) {
        const error =
          err && typeof err === "object" && "code" in err
            ? (err as EmbedError)
            : toEmbedError({
                code: "resolve_failed",
                message: err instanceof Error ? err.message : "Failed to resolve embed",
              });
        this.emitError(error);
        // Still mount iframe so operators can see in-widget error UI when key is only partially wrong —
        // except for missing key / clear client errors already handled.
        if (error.code === "embed_not_found" || error.code === "embed_disabled" || error.code === "origin_not_allowed") {
          this.mountErrorState(error);
          return;
        }
      }
    }

    const iframeSrc = buildEmbedIframeUrl({
      apiBase: normalizeApiBase(apiBase),
      embedKey,
      parentOrigin,
      agentId: this.options.agentId,
      theme: this.options.theme,
      locale: this.options.locale,
    });

    const position: EmbedPosition = this.options.position ?? "bottom-right";
    const zIndex = this.options.zIndex ?? 2147483000;
    const container = this.options.container ?? null;

    if (container) {
      this.widget = createInlineWidget({ container, iframeSrc });
      this.openState = true;
    } else {
      this.widget = createFloatingWidget({
        iframeSrc,
        position,
        zIndex,
        onOpenClick: () => this.open(),
        onCloseClick: () => this.close(),
      });
      this.widget.setOpen(false);
    }

    this.emit("ready", this.resolveResult);
    if (this.options.open || container) {
      this.open();
    }
  }

  private mountErrorState(error: EmbedError): void {
    if (typeof document === "undefined") return;
    const container = this.options.container;
    const root = document.createElement("div");
    root.setAttribute("data-monti-embed", "error");
    root.setAttribute("role", "alert");
    root.style.cssText = container
      ? "padding:16px;border-radius:12px;border:1px solid #f66;background:#2a1010;color:#fdd;font:14px system-ui,sans-serif;"
      : "position:fixed;right:20px;bottom:20px;z-index:2147483000;max-width:320px;padding:16px;border-radius:12px;border:1px solid #f66;background:#2a1010;color:#fdd;font:14px system-ui,sans-serif;box-shadow:0 8px 24px rgba(0,0,0,.4);";
    root.innerHTML = `<strong style="display:block;margin-bottom:6px">Monti embed error</strong><span style="opacity:.9">${escapeHtml(error.message)} <code style="font-size:12px">(${escapeHtml(String(error.code))})</code></span>`;

    if (container) container.appendChild(root);
    else document.body.appendChild(root);

    this.widget = {
      root,
      panel: root,
      iframe: document.createElement("iframe"),
      openBtn: null,
      closeBtn: null,
      setOpen() {},
      destroy() {
        root.remove();
      },
    };
  }

  private emitError(error: EmbedError): void {
    this.lastError = error;
    this.emit("error", error);
  }

  private emit<K extends EmbedEventName>(event: K, payload: EmbedEventMap[K]): void {
    const set = this.listeners.get(event);
    if (!set) return;
    for (const handler of [...set]) {
      try {
        (handler as Listener<K>)(payload);
      } catch (err) {
        console.error(`[monti-embed] listener error on ${event}`, err);
      }
    }
  }
}

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;");
}

/** Convenience: create + mount in one call. */
export async function createMontiEmbed(options: EmbedOptions): Promise<MontiEmbed> {
  const embed = new MontiEmbed(options);
  await embed.mount();
  return embed;
}
