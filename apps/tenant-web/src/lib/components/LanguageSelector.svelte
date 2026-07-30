<script lang="ts">
  import { lang, setLang, t, type Lang } from '$lib/i18n';

  type Props = {
    compact?: boolean;
  };

  let { compact = false }: Props = $props();

  function switchLang(next: Lang) {
    setLang(next);
  }
</script>

<div class="lang-toggle" class:compact role="group" aria-label={$t.lang_label}>
  <button type="button" class:active={$lang === 'en'} aria-pressed={$lang === 'en'} onclick={() => switchLang('en')}
    >EN</button
  >
  <button type="button" class:active={$lang === 'th'} aria-pressed={$lang === 'th'} onclick={() => switchLang('th')}
    >TH</button
  >
  <button type="button" class:active={$lang === 'ja'} aria-pressed={$lang === 'ja'} onclick={() => switchLang('ja')}
    >JA</button
  >
</div>

<style>
  .lang-toggle {
    display: inline-flex;
    gap: 2px;
    padding: 3px;
    border-radius: 999px;
    border: 1px solid color-mix(in srgb, var(--border, #2a3550) 90%, transparent);
    background: color-mix(in srgb, var(--panel, #0d1528) 85%, #000);
  }
  .lang-toggle button {
    border: 0;
    background: transparent;
    color: var(--muted, #8b9bb8);
    font: inherit;
    font-size: 11px;
    font-weight: 700;
    letter-spacing: 0.04em;
    padding: 4px 8px;
    border-radius: 999px;
    cursor: pointer;
    line-height: 1.2;
  }
  .lang-toggle button.active {
    background: color-mix(in srgb, var(--accent, #3d7dff) 28%, transparent);
    color: var(--text, #eef3ff);
  }
  .lang-toggle button:hover:not(.active) {
    color: var(--text, #eef3ff);
  }
  .lang-toggle.compact button {
    padding: 3px 6px;
    font-size: 10px;
  }
</style>
