import { browser } from '$app/environment';
import { derived, get, writable } from 'svelte/store';
import { messages, type Lang, type Messages } from './messages';

const STORAGE_KEY = 'monti_product_lang';

const SUPPORTED: Lang[] = ['en', 'th', 'ja'];

function isLang(v: string | null | undefined): v is Lang {
  return v === 'en' || v === 'th' || v === 'ja';
}

function detectInitial(): Lang {
  if (!browser) return 'en';
  const stored = localStorage.getItem(STORAGE_KEY);
  if (isLang(stored)) return stored;
  const nav = (navigator.language || '').toLowerCase();
  if (nav.startsWith('th')) return 'th';
  if (nav.startsWith('ja')) return 'ja';
  return 'en';
}

export const lang = writable<Lang>(detectInitial());

export const t = derived(lang, ($lang) => messages[$lang]);

export function getLang(): Lang {
  return get(lang);
}

export function setLang(next: Lang) {
  lang.set(next);
  if (browser) {
    localStorage.setItem(STORAGE_KEY, next);
    document.documentElement.lang = next;
    document.documentElement.dataset.lang = next;
  }
}

export function initLangFromUrl(searchParams: URLSearchParams) {
  const q = searchParams.get('lang');
  if (isLang(q)) {
    setLang(q);
    return;
  }
  if (browser) {
    document.documentElement.lang = get(lang);
    document.documentElement.dataset.lang = get(lang);
  }
}

export function supportedLangs(): Lang[] {
  return [...SUPPORTED];
}

export function msg(): Messages {
  return messages[get(lang)];
}

export type { Lang, Messages };
export { messages };
