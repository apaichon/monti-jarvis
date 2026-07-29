<script lang="ts">
  import { listPublicBrands, type PublicBrand } from '$lib/api/brands';
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
      loadError = err instanceof Error ? err.message : 'Failed to load brands';
      items = [];
      total = 0;
    } finally {
      loading = false;
    }
  }

  onMount(() => {
    void load();
  });

  function onSearchInput(value: string) {
    q = value;
    clearTimeout(searchTimer);
    searchTimer = setTimeout(() => void load(value.trim()), 250);
  }
</script>

<main class="tenant-picker" aria-label="Choose a brand to call">
  <header class="picker-hero">
    <img class="picker-logo" src="/images/monti-logo.png" width="56" height="56" alt="Monti" />
    <div>
      <h1>Choose who to call</h1>
      <p class="picker-sub">เลือกแบรนด์ที่ต้องการติดต่อ · Select a brand to start chat or voice</p>
    </div>
  </header>

  <div class="picker-search">
    <label class="sr-only" for="brand-search">Search brands</label>
    <input
      id="brand-search"
      type="search"
      placeholder="Search brands… · ค้นหาแบรนด์…"
      value={q}
      oninput={(e) => onSearchInput((e.currentTarget as HTMLInputElement).value)}
      autocomplete="off"
    />
  </div>

  {#if error || loadError}
    <div class="picker-error" role="alert">{error || loadError}</div>
  {/if}

  {#if loading}
    <div class="picker-status">Loading brands…</div>
  {:else if items.length === 0}
    <div class="picker-empty">
      <strong>No brands available</strong>
      <p>ยังไม่มีแบรนด์ที่เปิดให้บริการ · No public brands are listed yet.</p>
    </div>
  {:else}
    <ul class="picker-grid">
      {#each items as brand (brand.id)}
        <li>
          <button type="button" class="brand-card" onclick={() => onSelect(brand)}>
            <div class="brand-card-logo">
              {#if brand.logo_url}
                <img src={brand.logo_url} alt="" width="48" height="48" />
              {:else}
                <span>{(brand.name || brand.slug || '?').slice(0, 1).toUpperCase()}</span>
              {/if}
            </div>
            <div class="brand-card-body">
              <strong>{brand.name || brand.slug}</strong>
              {#if brand.blurb}
                <span class="brand-blurb">{brand.blurb}</span>
              {/if}
              <span class="brand-meta">{brand.slug}{#if brand.category} · {brand.category}{/if}</span>
            </div>
            <span class="brand-cta">Call →</span>
          </button>
        </li>
      {/each}
    </ul>
    {#if total > items.length}
      <div class="picker-status">Showing {items.length} of {total}</div>
    {/if}
  {/if}
</main>

<style>
  .tenant-picker {
    max-width: 960px;
    margin: 0 auto;
    padding: clamp(20px, 5vw, 48px);
    height: 100vh;
    overflow: auto;
  }
  .picker-hero {
    display: flex;
    gap: 16px;
    align-items: center;
    margin-bottom: 24px;
  }
  .picker-logo {
    border-radius: 14px;
    background: rgb(8 23 42 / 80%);
  }
  .picker-hero h1 {
    margin: 0;
    font-size: clamp(1.4rem, 3vw, 1.85rem);
  }
  .picker-sub {
    margin: 6px 0 0;
    color: var(--muted);
    font-size: 0.95rem;
  }
  .picker-search input {
    width: 100%;
    box-sizing: border-box;
    border: 1px solid var(--line);
    border-radius: 16px;
    background: rgb(7 17 32 / 90%);
    color: var(--ink);
    padding: 14px 16px;
    margin-bottom: 20px;
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
    grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
    gap: 14px;
  }
  .brand-card {
    width: 100%;
    text-align: left;
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 12px;
    align-items: center;
    padding: 14px;
    border-radius: 18px;
    border: 1px solid var(--line);
    background:
      linear-gradient(180deg, rgb(8 20 38 / 88%), rgb(3 11 23 / 94%));
    color: inherit;
  }
  .brand-card:hover {
    border-color: rgb(0 183 255 / 45%);
  }
  .brand-card-logo {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    overflow: hidden;
    display: grid;
    place-items: center;
    background: rgb(0 109 255 / 18%);
    font-weight: 700;
  }
  .brand-card-logo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
  }
  .brand-card-body {
    display: grid;
    gap: 4px;
    min-width: 0;
  }
  .brand-card-body strong {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .brand-blurb,
  .brand-meta {
    color: var(--muted);
    font-size: 0.85rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .brand-cta {
    color: var(--cyan);
    font-weight: 600;
    white-space: nowrap;
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
</style>
