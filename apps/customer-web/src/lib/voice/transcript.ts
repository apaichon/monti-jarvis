/**
 * Gemini Live streams input/output transcription as short partial fragments
 * (or growing cumulative strings). Merge into one readable caption per turn.
 */

export type PreferredLang = 'th' | 'en' | '';

/** Merge a new transcript chunk into the turn buffer. */
export function mergeTranscriptChunk(prev: string, chunk: string): string {
  const p = (prev || '').trim();
  let c = (chunk || '').trim();
  if (!c) return p;
  if (!p) return c;

  // Cumulative full string (API sometimes re-sends growing text).
  if (c.startsWith(p)) return c;
  if (p.startsWith(c)) return p;
  if (p.includes(c) && c.length < p.length) return p;
  if (c.includes(p) && p.length < c.length) return c;

  // Overlap: end of prev matches start of chunk (dedupe join).
  const maxOverlap = Math.min(p.length, c.length, 48);
  for (let n = maxOverlap; n >= 4; n--) {
    if (p.slice(-n) === c.slice(0, n)) {
      return p + c.slice(n);
    }
  }

  const hasThai = /[\u0E00-\u0E7F]/.test(p) || /[\u0E00-\u0E7F]/.test(c);
  if (hasThai) {
    // Thai usually joins without a space between chunks.
    return p + c;
  }
  const needSpace = !/[\s.,!?;:)]$/.test(p) && !/^[.,!?;:(]/.test(c);
  return needSpace ? `${p} ${c}` : p + c;
}

function thaiCharCount(s: string): number {
  return (s.match(/[\u0E00-\u0E7F]/g) || []).join('').length;
}

function latinWordishCount(s: string): number {
  return (s.match(/[A-Za-z]{2,}/g) || []).length;
}

/**
 * Prefer a single language for on-screen captions when the model mixes
 * Thai + English in one utterance (common with dual-language voice replies).
 */
export function preferMainLanguage(text: string, preferred: PreferredLang = ''): string {
  const raw = (text || '').trim();
  if (!raw) return raw;

  // Split dual-line captions: "Thai line\nEnglish line" or "… / …"
  let parts = raw
    .split(/\n+/)
    .map((s) => s.trim())
    .filter(Boolean);
  if (parts.length === 1 && raw.includes(' / ')) {
    parts = raw.split(/\s+\/\s+/).map((s) => s.trim()).filter(Boolean);
  }

  if (parts.length > 1) {
    const thaiParts = parts.filter((p) => thaiCharCount(p) >= 2);
    const enParts = parts.filter((p) => latinWordishCount(p) >= 2 && thaiCharCount(p) < 2);
    if (preferred === 'th' && thaiParts.length) return thaiParts.join(' ').trim();
    if (preferred === 'en' && enParts.length) return enParts.join(' ').trim();
    // Auto: keep the script that dominates the whole string.
    if (thaiCharCount(raw) >= latinWordishCount(raw) * 2 && thaiParts.length) {
      return thaiParts.join(' ').trim();
    }
    if (enParts.length && thaiParts.length) {
      // Mixed dual lines — keep majority script of full text.
      return thaiCharCount(raw) > 8 ? thaiParts.join(' ').trim() : enParts.join(' ').trim();
    }
  }

  // Same line mixed: "สวัสดี Hello how are you" — drop the secondary clause if preferred.
  if (preferred === 'th' && thaiCharCount(raw) >= 4) {
    // Remove long Latin runs (likely translation tail).
    const stripped = raw
      .replace(/(?:^|\s)[A-Za-z][A-Za-z0-9'’.,!?;:\-\s]{12,}$/u, '')
      .replace(/^[A-Za-z][A-Za-z0-9'’.,!?;:\-\s]{12,}(?=[\u0E00-\u0E7F])/u, '')
      .trim();
    if (thaiCharCount(stripped) >= 4) return stripped;
  }
  if (preferred === 'en' && latinWordishCount(raw) >= 2) {
    const stripped = raw.replace(/[\u0E00-\u0E7F][\u0E00-\u0E7F\s.,!?;:0-9]*/g, ' ').replace(/\s+/g, ' ').trim();
    if (latinWordishCount(stripped) >= 2) return stripped;
  }

  return raw;
}
