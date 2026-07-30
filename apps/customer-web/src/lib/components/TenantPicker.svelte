<script lang="ts">
  import { listPublicBrands, type PublicBrand } from '$lib/api/brands';
  import { brandMonogram } from '$lib/brandMark';
  import LanguageSelector from '$lib/components/LanguageSelector.svelte';
  import { initLangFromUrl, t } from '$lib/i18n';
  import { onMount } from 'svelte';

  type Props = {
    onSelect: (brand: PublicBrand) => void;
    error?: string;
  };

  let { onSelect, error = '' }: Props = $props();

  let items = $state<PublicBrand[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let loadError = $state('');
  let q = $state('');
  let searchTimer: ReturnType<typeof setTimeout> | undefined;

  async function load(query = '') {
    loading = true;
    loadError = '';
    try {
      const res = await listPublicBrands({ q: query, limit: 50 });
      items = res.items;
      total = res.total;
    } catch (err) {
      loadError = err instanceof Error ? err.message : $t.picker_load_error;
      items = [];
      total = 0;
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    initLangFromUrl(new URLSearchParams(window.location.search));
    void load();
  });

  function onSearchInput(value: string) {
    q = value;
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => void load(value.trim()), 250);
  }
</script>

<main class="tenant-picker" aria-label={$t.picker_aria}>
  <header class="picker-hero">
    <div class="picker-hero-top">
      <div class="monti-hero">
        <div class="monti-hero-ring">
          <img class="monti-hero-img" src="/images/monti-logo.png" width="120" height="120" alt="Monti" />
        </div>
        <div class="monti-wordmark">
          <span class="monti-title">MONTI</span>
          <span class="monti-tag">AI CALL CENTER</span>
        </div>
      </div>
      <LanguageSelector />
    </div>
    <div class="picker-copy">
      <h1>{$t.picker_title}</h1>
      <p class="picker-sub">{$t.picker_sub}</p>
    </div>
  </header>

  <div class="picker-search">
    <label class="sr-only" for="brand-search">{$t.picker_search}</label>
    <input
      id="brand-search"
      type="search"
      placeholder={$t.picker_search_ph}
      value={q}
      oninput={(e) => onSearchInput((e.currentTarget as HTMLInputElement).value)}
      autocomplete="off"
    />
  </div>

  {#if error || loadError}
    <div class="picker-error" role="alert">{error || loadError}</div>
  {/if}

  {#if loading}
    <div class="picker-status">{$t.picker_loading}</div>
  {:else if items.length === 0}
    <div class="picker-empty">
      <strong>{$t.picker_empty_title}</strong>
      <p>{$t.picker_empty_body}</p>
    </div>
  {:else}
    <ul class="picker-grid">
      {#each items as brand (brand.id)}
        <li>
          <button type="button" class="brand-card" onclick={() => onSelect(brand)}>
            <div class="brand-card-logo">
              {#if brand.logo_url && !brand.logo_url.includes('monti-logo')}
                <img src={brand.logo_url} alt="" />
              {:else}
                <span class="brand-monogram">{brandMonogram(brand.name, brand.slug)}</span>
              {/if}
            </div>
            <strong class="brand-name">{brand.name || brand.slug}</strong>
            <span class="brand-meta">{brand.slug}</span>
            <span class="brand-badge">{$t.picker_badge}</span>
            <span class="brand-cta">
              <svg class="call-icon" viewBox="0 0 24 24" width="18" height="18" aria-hidden="true" focusable="false">
                <path
                  fill="currentColor"
                  d="M6.6 10.8c1.4 2.8 3.8 5.1 6.6 6.6l2.2-2.2c.3-.3.7-.4 1.1-.3 1.2.4 2.5.6 3.8.6.6 0 1 .4 1 1V20c0 .6-.4 1-1 1C10.6 21 3 13.4 3 4c0-.6.4-1 1-1h3.5c.6 0 1 .4 1 1 0 1.3.2 2.6.6 3.8.1.4 0 .8-.3 1.1L6.6 10.8z"
                />
              </svg>
              {$t.picker_call}
            </span>
          </button>
        </li>
      {/each}
    </ul>
    {#if total > items.length}
      <div class="picker-status">{$t.picker_showing} {items.length} / {total}</div>
    {/if}
  {/if}

  <footer class="picker-foot">
    <span>{$t.picker_foot_secure}</span>
    <span>{$t.picker_foot_ai}</span>
  </footer>
</main>

<style>
  .tenant-picker {
    max-width: 1100px;
    margin: 0 auto;
    padding: clamp(24px, 5vw, 56px);
    height: 100vh;
    overflow: auto;
    box-sizing: border-box;
  }
  .picker-hero {
    display: grid;
    gap: 20px;
    margin-bottom: 28px;
    justify-items: start;
  }
  .picker-hero-top {
    width: 100%;
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
  }
  .monti-hero {
    display: flex;
    align-items: center;
    gap: 18px;
  }
  .monti-hero-ring {
    width: 112px;
    height: 112px;
    border-radius: 50%;
    padding: 4px;
    background:
      radial-gradient(circle at 50% 30%, rgb(0 183 255 / 35%), transparent 55%),
      linear-gradient(145deg, rgb(0 183 255 / 50%), rgb(0 80 200 / 20%));
    box-shadow: 0 0 40px rgb(0 140 255 / 35%);
    display: grid;
    place-items: center;
  }
  .monti-hero-img {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    object-fit: cover;
    border: 2px solid rgb(0 183 255 / 40%);
    background: #04101f;
  }
  .monti-wordmark {
    display: grid;
    gap: 2px;
  }
  .monti-title {
    font-size: clamp(1.8rem, 4vw, 2.4rem);
    font-weight: 800;
    letter-spacing: 0.18em;
    line-height: 1;
  }
  .monti-tag {
    color: var(--cyan);
    font-size: 0.8rem;
    font-weight: 700;
    letter-spacing: 0.22em;
  }
  .picker-copy h1 {
    margin: 0;
    font-size: clamp(1.5rem, 3vw, 2rem);
  }
  .picker-sub {
    margin: 8px 0 0;
    color: var(--muted);
    font-size: 0.95rem;
  }
  .picker-search input {
    width: 100%;
    box-sizing: border-box;
    border: 1px solid var(--line);
    border-radius: 18px;
    background: rgb(7 17 32 / 90%);
    color: var(--ink);
    padding: 16px 18px;
    margin-bottom: 22px;
    font-size: 1rem;
  }
  .picker-error {
    color: var(--red);
    margin-bottom: 12px;
  }
  .picker-status {
    color: var(--muted);
    margin: 12px 0;
  }
  .picker-empty {
    border: 1px dashed var(--line);
    border-radius: 20px;
    padding: 28px;
    background: rgb(5 16 31 / 70%);
  }
  .picker-empty p {
    color: var(--muted);
    margin: 8px 0 0;
  }
  .picker-grid {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 16px;
  }
  .brand-card {
    width: 100%;
    text-align: center;
    display: grid;
    gap: 10px;
    justify-items: center;
    padding: 22px 16px 18px;
    border-radius: 22px;
    border: 1px solid var(--line);
    background:
      linear-gradient(180deg, rgb(8 20 38 / 92%), rgb(3 11 23 / 96%));
    color: inherit;
    min-height: 240px;
    transition: border-color 0.15s ease, box-shadow 0.15s ease, transform 0.15s ease;
  }
  .brand-card:hover {
    border-color: rgb(0 183 255 / 55%);
    box-shadow: 0 12px 40px rgb(0 100 255 / 18%);
    transform: translateY(-2px);
  }
  .brand-card-logo {
    width: 88px;
    height: 88px;
    border-radius: 50%;
    overflow: hidden;
    display: grid;
    place-items: center;
    background: rgb(0 109 255 / 16%);
    border: 1px solid rgb(0 183 255 / 28%);
    box-shadow: 0 0 24px rgb(0 140 255 / 18%);
  }
  .brand-card-logo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .brand-monogram {
    font-size: 1.6rem;
    font-weight: 800;
    letter-spacing: 0.04em;
    color: var(--cyan);
  }
  .brand-name {
    font-size: 1.05rem;
    line-height: 1.25;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
  }
  .brand-meta {
    color: var(--muted);
    font-size: 0.8rem;
    max-width: 100%;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .brand-badge {
    font-size: 0.72rem;
    font-weight: 600;
    color: var(--cyan);
    border: 1px solid rgb(0 183 255 / 30%);
    border-radius: 999px;
    padding: 4px 10px;
    background: rgb(0 110 255 / 10%);
  }
  .brand-cta {
    margin-top: 4px;
    width: 100%;
    box-sizing: border-box;
    border-radius: 12px;
    padding: 10px 12px;
    font-weight: 700;
    color: #fff;
    background: linear-gradient(135deg, #006dff, #00b7ff);
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
  }
  .call-icon {
    flex-shrink: 0;
  }
  .picker-foot {
    margin-top: 28px;
    display: flex;
    flex-wrap: wrap;
    gap: 16px 28px;
    color: var(--muted);
    font-size: 0.85rem;
  }
  .sr-only {
    position: absolute;
    width: 1px;
    height: 1px;
    padding: 0;
    margin: -1px;
    overflow: hidden;
    clip: rect(0, 0, 0, 0);
    border: 0;
  }
  @media (max-width: 640px) {
    .monti-hero {
      flex-direction: column;
      align-items: flex-start;
    }
    .picker-grid {
      grid-template-columns: 1fr 1fr;
    }
  }
</style>
