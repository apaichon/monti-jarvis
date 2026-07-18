import {
  MontiEmbed,
  type EmbedError,
  type EmbedPosition,
  type EmbedResolveResult,
} from "@monti/embed-core";

export type { EmbedError, EmbedPosition, EmbedResolveResult };
export { MontiEmbed };

const OBSERVED = [
  "embed-key",
  "api-base",
  "parent-origin",
  "position",
  "agent-id",
  "theme",
  "locale",
  "open",
  "inline",
  "skip-resolve",
] as const;

/**
 * Framework-agnostic custom element.
 *
 * ```html
 * <script type="module">
 *   import '@monti/embed-web-component';
 * </script>
 * <monti-embed
 *   embed-key="emb_…"
 *   api-base="http://localhost:8091"
 *   position="bottom-right"
 * ></monti-embed>
 * ```
 */
export class MontiEmbedElement extends HTMLElement {
  static get observedAttributes(): string[] {
    return [...OBSERVED];
  }

  private embed: MontiEmbed | null = null;
  private mountToken = 0;
  private host: HTMLDivElement | null = null;

  connectedCallback(): void {
    if (!this.host) {
      this.host = document.createElement("div");
      this.host.style.display = this.hasAttribute("inline") ? "block" : "contents";
      if (this.hasAttribute("inline")) {
        this.host.style.width = "100%";
        this.host.style.minHeight = "480px";
        this.host.style.height = "100%";
      }
      this.appendChild(this.host);
    }
    void this.remount();
  }

  disconnectedCallback(): void {
    this.teardown();
  }

  attributeChangedCallback(): void {
    if (!this.isConnected) return;
    void this.remount();
  }

  /** Imperative API for Angular / plain JS hosts. */
  open(): void {
    this.embed?.open();
  }

  close(): void {
    this.embed?.close();
  }

  toggle(): void {
    this.embed?.toggle();
  }

  getInstance(): MontiEmbed | null {
    return this.embed;
  }

  private teardown(): void {
    this.mountToken += 1;
    this.embed?.destroy();
    this.embed = null;
  }

  private attr(name: string): string {
    return (this.getAttribute(name) || "").trim();
  }

  private boolAttr(name: string): boolean {
    if (!this.hasAttribute(name)) return false;
    const v = (this.getAttribute(name) || "").toLowerCase();
    return v === "" || v === "true" || v === "1" || v === "yes";
  }

  private async remount(): Promise<void> {
    const token = ++this.mountToken;
    this.embed?.destroy();
    this.embed = null;

    const embedKey = this.attr("embed-key");
    const apiBase = this.attr("api-base");
    if (!embedKey || !apiBase) {
      this.dispatchEvent(
        new CustomEvent("monti-error", {
          detail: {
            code: !embedKey ? "missing_embed_key" : "missing_api_base",
            message: !embedKey ? "embed-key is required" : "api-base is required",
          } satisfies EmbedError,
          bubbles: true,
        }),
      );
      return;
    }

    const inline = this.boolAttr("inline");
    if (this.host) {
      this.host.style.display = inline ? "block" : "contents";
    }

    const position = (this.attr("position") || "bottom-right") as EmbedPosition;
    const embed = new MontiEmbed({
      embedKey,
      apiBase,
      parentOrigin: this.attr("parent-origin") || undefined,
      position,
      agentId: this.attr("agent-id") || undefined,
      theme: this.attr("theme") || undefined,
      locale: this.attr("locale") || undefined,
      open: this.boolAttr("open"),
      skipResolve: this.boolAttr("skip-resolve"),
      container: inline ? this.host : null,
      onOpen: () => this.dispatchEvent(new CustomEvent("monti-open", { bubbles: true })),
      onClose: () => this.dispatchEvent(new CustomEvent("monti-close", { bubbles: true })),
      onReady: (result?: EmbedResolveResult) =>
        this.dispatchEvent(new CustomEvent("monti-ready", { detail: result, bubbles: true })),
      onError: (error: EmbedError) =>
        this.dispatchEvent(new CustomEvent("monti-error", { detail: error, bubbles: true })),
      onDestroy: () => this.dispatchEvent(new CustomEvent("monti-destroy", { bubbles: true })),
    });

    await embed.mount();
    if (token !== this.mountToken) {
      embed.destroy();
      return;
    }
    this.embed = embed;
  }
}

declare global {
  interface HTMLElementTagNameMap {
    "monti-embed": MontiEmbedElement;
  }
}

export function defineMontiEmbedElement(tagName = "monti-embed"): void {
  if (typeof customElements === "undefined") return;
  if (customElements.get(tagName)) return;
  customElements.define(tagName, MontiEmbedElement);
}

// Auto-register when imported in the browser
defineMontiEmbedElement();

export default MontiEmbedElement;
