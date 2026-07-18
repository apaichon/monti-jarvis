import type { EmbedError, EmbedResolveResult } from "./types.js";

export function normalizeApiBase(apiBase: string): string {
  return apiBase.replace(/\/+$/, "");
}

export function buildEmbedIframeUrl(options: {
  apiBase: string;
  embedKey: string;
  parentOrigin: string;
  agentId?: string;
  theme?: string;
  locale?: string;
}): string {
  const base = normalizeApiBase(options.apiBase);
  const url = new URL(`${base}/embed`);
  url.searchParams.set("key", options.embedKey);
  if (options.parentOrigin) {
    url.searchParams.set("parent_origin", options.parentOrigin);
  }
  if (options.agentId) url.searchParams.set("agent", options.agentId);
  if (options.theme) url.searchParams.set("theme", options.theme);
  if (options.locale) url.searchParams.set("locale", options.locale);
  return url.toString();
}

export function buildResolveUrl(apiBase: string, embedKey: string, parentOrigin: string): string {
  const base = normalizeApiBase(apiBase);
  const url = new URL(`${base}/api/public/embed/${encodeURIComponent(embedKey)}`);
  if (parentOrigin) url.searchParams.set("parent_origin", parentOrigin);
  return url.toString();
}

export async function resolveEmbed(options: {
  apiBase: string;
  embedKey: string;
  parentOrigin: string;
  fetch?: typeof fetch;
}): Promise<EmbedResolveResult> {
  const fetcher = options.fetch ?? globalThis.fetch?.bind(globalThis);
  if (!fetcher) {
    throw toEmbedError({ code: "network_error", message: "fetch is not available" });
  }

  const url = buildResolveUrl(options.apiBase, options.embedKey, options.parentOrigin);
  let response: Response;
  try {
    response = await fetcher(url, {
      method: "GET",
      headers: {
        Accept: "application/json",
        ...(options.parentOrigin ? { "X-Embed-Parent-Origin": options.parentOrigin } : {}),
      },
      credentials: "omit",
    });
  } catch (err) {
    throw toEmbedError({
      code: "network_error",
      message: err instanceof Error ? err.message : "Network error resolving embed",
    });
  }

  const raw = await response.text();
  let payload: Record<string, unknown> = {};
  try {
    payload = raw ? (JSON.parse(raw) as Record<string, unknown>) : {};
  } catch {
    payload = {};
  }

  if (!response.ok) {
    const code =
      typeof payload.code === "string"
        ? payload.code
        : response.status === 404
          ? "embed_not_found"
          : response.status === 403
            ? "origin_not_allowed"
            : "resolve_failed";
    const message =
      typeof payload.error === "string"
        ? payload.error
        : typeof payload.message === "string"
          ? payload.message
          : code;
    throw toEmbedError({ code, message, status: response.status });
  }

  return payload as unknown as EmbedResolveResult;
}

export function toEmbedError(partial: EmbedError): EmbedError {
  return {
    code: partial.code,
    message: partial.message || String(partial.code),
    status: partial.status,
  };
}
