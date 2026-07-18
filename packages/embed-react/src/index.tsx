import {
  useCallback,
  useEffect,
  useImperativeHandle,
  useRef,
  useState,
  forwardRef,
  type CSSProperties,
} from "react";
import {
  MontiEmbed,
  type EmbedError,
  type EmbedPosition,
  type EmbedResolveResult,
} from "@monti/embed-core";

export type { EmbedError, EmbedPosition, EmbedResolveResult };
export { MontiEmbed };

export interface MontiEmbedProps {
  embedKey: string;
  apiBase: string;
  parentOrigin?: string;
  position?: EmbedPosition;
  agentId?: string;
  theme?: string;
  locale?: string;
  /** Controlled open state (floating mode). */
  open?: boolean;
  /** Uncontrolled initial open. */
  defaultOpen?: boolean;
  /** Mount iframe inside the component host (inline). */
  inline?: boolean;
  skipResolve?: boolean;
  className?: string;
  style?: CSSProperties;
  onOpen?: () => void;
  onClose?: () => void;
  onReady?: (result?: EmbedResolveResult) => void;
  onError?: (error: EmbedError) => void;
  onDestroy?: () => void;
  onOpenChange?: (open: boolean) => void;
}

export interface MontiEmbedHandle {
  open: () => void;
  close: () => void;
  toggle: () => void;
  destroy: () => void;
  getInstance: () => MontiEmbed | null;
}

export const MontiEmbedReact = forwardRef<MontiEmbedHandle, MontiEmbedProps>(function MontiEmbedReact(
  props,
  ref,
) {
  const hostRef = useRef<HTMLDivElement | null>(null);
  const embedRef = useRef<MontiEmbed | null>(null);
  const [error, setError] = useState<EmbedError | null>(null);

  const {
    embedKey,
    apiBase,
    parentOrigin,
    position = "bottom-right",
    agentId,
    theme,
    locale,
    open,
    defaultOpen = false,
    inline = false,
    skipResolve = false,
    className,
    style,
    onOpen,
    onClose,
    onReady,
    onError,
    onDestroy,
    onOpenChange,
  } = props;

  // Keep latest callbacks without remounting
  const cb = useRef({ onOpen, onClose, onReady, onError, onDestroy, onOpenChange });
  cb.current = { onOpen, onClose, onReady, onError, onDestroy, onOpenChange };

  useEffect(() => {
    let cancelled = false;
    const embed = new MontiEmbed({
      embedKey,
      apiBase,
      parentOrigin,
      position,
      agentId,
      theme,
      locale,
      open: open ?? defaultOpen,
      skipResolve,
      container: inline ? hostRef.current : null,
      onOpen: () => {
        cb.current.onOpen?.();
        cb.current.onOpenChange?.(true);
      },
      onClose: () => {
        cb.current.onClose?.();
        cb.current.onOpenChange?.(false);
      },
      onReady: (r) => cb.current.onReady?.(r),
      onError: (e) => {
        if (!cancelled) setError(e);
        cb.current.onError?.(e);
      },
      onDestroy: () => cb.current.onDestroy?.(),
    });
    embedRef.current = embed;
    void embed.mount();

    return () => {
      cancelled = true;
      embed.destroy();
      if (embedRef.current === embed) embedRef.current = null;
    };
  }, [
    embedKey,
    apiBase,
    parentOrigin,
    position,
    agentId,
    theme,
    locale,
    inline,
    skipResolve,
    defaultOpen,
    // intentionally omit `open` — controlled via separate effect
  ]);

  useEffect(() => {
    if (open === undefined) return;
    const embed = embedRef.current;
    if (!embed) return;
    if (open) embed.open();
    else embed.close();
  }, [open]);

  useImperativeHandle(
    ref,
    () => ({
      open: () => embedRef.current?.open(),
      close: () => embedRef.current?.close(),
      toggle: () => embedRef.current?.toggle(),
      destroy: () => {
        embedRef.current?.destroy();
        embedRef.current = null;
      },
      getInstance: () => embedRef.current,
    }),
    [],
  );

  const hostStyle: CSSProperties = inline
    ? { width: "100%", minHeight: 480, height: "100%", ...style }
    : { display: "contents", ...style };

  return (
    <div
      ref={hostRef}
      data-monti-embed-react="1"
      data-error={error?.code}
      className={className}
      style={hostStyle}
    />
  );
});

/** Hook that owns a MontiEmbed instance for custom UI shells. */
export function useMontiEmbed(
  options: Omit<MontiEmbedProps, "className" | "style" | "inline"> & {
    container?: HTMLElement | null;
    enabled?: boolean;
  },
): {
  embed: MontiEmbed | null;
  error: EmbedError | null;
  open: () => void;
  close: () => void;
  toggle: () => void;
} {
  const embedRef = useRef<MontiEmbed | null>(null);
  const [error, setError] = useState<EmbedError | null>(null);
  const [tick, setTick] = useState(0);
  const enabled = options.enabled !== false;

  useEffect(() => {
    if (!enabled) {
      embedRef.current?.destroy();
      embedRef.current = null;
      return;
    }
    const embed = new MontiEmbed({
      embedKey: options.embedKey,
      apiBase: options.apiBase,
      parentOrigin: options.parentOrigin,
      position: options.position,
      agentId: options.agentId,
      theme: options.theme,
      locale: options.locale,
      open: options.open ?? options.defaultOpen,
      skipResolve: options.skipResolve,
      container: options.container ?? null,
      onOpen: options.onOpen,
      onClose: options.onClose,
      onReady: options.onReady,
      onError: (e) => {
        setError(e);
        options.onError?.(e);
      },
      onDestroy: options.onDestroy,
    });
    embedRef.current = embed;
    void embed.mount().then(() => setTick((t) => t + 1));
    return () => {
      embed.destroy();
      if (embedRef.current === embed) embedRef.current = null;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps -- remount on identity fields
  }, [
    enabled,
    options.embedKey,
    options.apiBase,
    options.parentOrigin,
    options.position,
    options.agentId,
    options.theme,
    options.locale,
    options.skipResolve,
    options.container,
  ]);

  const open = useCallback(() => embedRef.current?.open(), []);
  const close = useCallback(() => embedRef.current?.close(), []);
  const toggle = useCallback(() => embedRef.current?.toggle(), []);

  return {
    embed: tick >= 0 ? embedRef.current : null,
    error,
    open,
    close,
    toggle,
  };
}

export default MontiEmbedReact;
