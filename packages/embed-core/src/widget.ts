import type { EmbedPosition } from "./types.js";

const POSITIONS: Record<
  EmbedPosition,
  { right: string; bottom: string; left: string; top: string }
> = {
  "bottom-right": { right: "20px", bottom: "20px", left: "auto", top: "auto" },
  "bottom-left": { left: "20px", bottom: "20px", right: "auto", top: "auto" },
  "top-right": { right: "20px", top: "20px", left: "auto", bottom: "auto" },
  "top-left": { left: "20px", top: "20px", right: "auto", bottom: "auto" },
};

const IFRAME_ALLOW =
  "microphone *; autoplay *; camera *; clipboard-write *; display-capture *";

export interface WidgetHandles {
  root: HTMLElement;
  panel: HTMLElement;
  iframe: HTMLIFrameElement;
  openBtn: HTMLButtonElement | null;
  closeBtn: HTMLButtonElement | null;
  setOpen: (open: boolean) => void;
  destroy: () => void;
}

export function createFloatingWidget(options: {
  iframeSrc: string;
  position: EmbedPosition;
  zIndex: number;
  onOpenClick: () => void;
  onCloseClick: () => void;
}): WidgetHandles {
  const pos = POSITIONS[options.position] ?? POSITIONS["bottom-right"];
  const z = options.zIndex;

  const root = document.createElement("div");
  root.id = "monti-embed-root";
  root.setAttribute("data-monti-embed", "1");
  root.style.cssText = `all:initial;position:fixed;z-index:${z};font-family:system-ui,sans-serif;`;

  const panel = document.createElement("div");
  panel.style.cssText = [
    "display:none",
    "position:fixed",
    `right:${pos.right}`,
    `bottom:${pos.bottom}`,
    `left:${pos.left}`,
    `top:${pos.top}`,
    "width:min(400px,calc(100vw - 24px))",
    "height:min(680px,calc(100vh - 40px))",
    "border-radius:16px",
    "overflow:hidden",
    "box-shadow:0 16px 48px rgba(0,0,0,.35)",
    "border:1px solid rgba(0,183,255,.35)",
    "background:#05101f",
    `z-index:${z + 1}`,
  ].join(";");

  const iframe = document.createElement("iframe");
  iframe.title = "Monti AI Assistant";
  iframe.setAttribute("allow", IFRAME_ALLOW);
  iframe.allow = IFRAME_ALLOW;
  iframe.setAttribute("allowfullscreen", "true");
  iframe.setAttribute("referrerpolicy", "strict-origin-when-cross-origin");
  iframe.style.cssText = "width:100%;height:100%;border:0;display:block;background:#05101f;";
  iframe.src = options.iframeSrc;
  panel.appendChild(iframe);

  const closeBtn = document.createElement("button");
  closeBtn.type = "button";
  closeBtn.setAttribute("aria-label", "Close Monti chat");
  closeBtn.textContent = "✕";
  closeBtn.style.cssText = [
    "position:absolute",
    "top:10px",
    "right:10px",
    "z-index:3",
    "width:32px",
    "height:32px",
    "border-radius:50%",
    "border:1px solid rgba(0,183,255,.4)",
    "cursor:pointer",
    "background:rgba(8,20,36,.92)",
    "color:#f7fbff",
    "font-size:14px",
    "line-height:1",
    "box-shadow:0 4px 12px rgba(0,0,0,.35)",
    "display:none",
    "padding:0",
  ].join(";");
  panel.appendChild(closeBtn);

  const openBtn = document.createElement("button");
  openBtn.type = "button";
  openBtn.setAttribute("aria-label", "Open Monti chat");
  openBtn.textContent = "💬";
  openBtn.style.cssText = [
    "position:fixed",
    `right:${pos.right}`,
    `bottom:${pos.bottom}`,
    `left:${pos.left}`,
    `top:${pos.top}`,
    "width:56px",
    "height:56px",
    "border-radius:50%",
    "border:0",
    "cursor:pointer",
    "background:linear-gradient(135deg,#0084ff,#00b7ff)",
    "color:#fff",
    "font-size:22px",
    "box-shadow:0 8px 24px rgba(0,120,255,.45)",
    `z-index:${z + 2}`,
  ].join(";");

  const onOpen = () => options.onOpenClick();
  const onClose = () => options.onCloseClick();
  openBtn.addEventListener("click", onOpen);
  closeBtn.addEventListener("click", onClose);

  root.appendChild(panel);
  root.appendChild(openBtn);
  document.body.appendChild(root);

  return {
    root,
    panel,
    iframe,
    openBtn,
    closeBtn,
    setOpen(open: boolean) {
      panel.style.display = open ? "block" : "none";
      closeBtn.style.display = open ? "block" : "none";
      openBtn.style.display = open ? "none" : "block";
    },
    destroy() {
      openBtn.removeEventListener("click", onOpen);
      closeBtn.removeEventListener("click", onClose);
      root.remove();
    },
  };
}

export function createInlineWidget(options: {
  container: HTMLElement;
  iframeSrc: string;
}): WidgetHandles {
  const root = document.createElement("div");
  root.setAttribute("data-monti-embed", "inline");
  root.style.cssText = "width:100%;height:100%;min-height:480px;position:relative;";

  const panel = document.createElement("div");
  panel.style.cssText =
    "display:block;position:absolute;inset:0;border-radius:16px;overflow:hidden;" +
    "border:1px solid rgba(0,183,255,.35);background:#05101f;";

  const iframe = document.createElement("iframe");
  iframe.title = "Monti AI Assistant";
  iframe.setAttribute("allow", IFRAME_ALLOW);
  iframe.allow = IFRAME_ALLOW;
  iframe.setAttribute("allowfullscreen", "true");
  iframe.setAttribute("referrerpolicy", "strict-origin-when-cross-origin");
  iframe.style.cssText = "width:100%;height:100%;border:0;display:block;background:#05101f;";
  iframe.src = options.iframeSrc;
  panel.appendChild(iframe);
  root.appendChild(panel);
  options.container.appendChild(root);

  return {
    root,
    panel,
    iframe,
    openBtn: null,
    closeBtn: null,
    setOpen(_open: boolean) {
      /* inline is always visible */
    },
    destroy() {
      root.remove();
    },
  };
}

export function warnInsecureMontiHost(apiBase: string): void {
  try {
    const montiUrl = new URL(apiBase, typeof window !== "undefined" ? window.location.href : "http://localhost");
    const host = (montiUrl.hostname || "").toLowerCase();
    const secure =
      montiUrl.protocol === "https:" ||
      host === "localhost" ||
      host === "127.0.0.1" ||
      host.endsWith(".localhost");
    if (!secure) {
      console.warn(
        `[monti-embed] Voice/mic will fail: Monti host is not a secure context (${montiUrl.origin}). ` +
          "Use https:// or http://localhost:PORT.",
      );
    }
  } catch {
    /* ignore */
  }
}
