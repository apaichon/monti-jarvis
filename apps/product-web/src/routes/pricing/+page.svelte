<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { getPublicPackages, postFunnelEvent, type PublicPackage, ApiError } from '$lib/api';
  import { registerHref } from '$lib/attribution';
  import { t, msg } from '$lib/i18n';

  type Period = 'monthly' | 'annual';

  type DisplayTier = {
    id: string;
    name: string;
    blurb: string;
    priceLabel: string;
    periodSuffix: string;
    features: string[];
    cta: string;
    ctaHref: string;
    featured?: boolean;
    badge?: string;
    mutedCta?: boolean;
    isCustom?: boolean;
    quoteOnly?: boolean;
  };

  let packages = $state<PublicPackage[]>([]);
  let loading = $state(true);
  let error = $state('');
  let period = $state<Period>('monthly');

  onMount(async () => {
    loading = true;
    error = '';
    try {
      const res = await getPublicPackages();
      packages = res.packages ?? [];
    } catch (err) {
      packages = [];
      error =
        err instanceof ApiError
          ? err.message || 'Package catalog is temporarily unavailable.'
          : 'Package catalog is temporarily unavailable.';
    } finally {
      loading = false;
    }
  });

  function formatAmount(amount: number, currency: string) {
    const cur = (currency || 'THB').toUpperCase();
    try {
      return new Intl.NumberFormat('th-TH', {
        style: 'currency',
        currency: cur,
        maximumFractionDigits: 0
      }).format(amount);
    } catch {
      if (cur === 'THB') return `฿${amount.toLocaleString('en-US')}`;
      return `${cur} ${amount}`;
    }
  }

  function featuresFromPkg(pkg: PublicPackage): string[] {
    if (pkg.highlights?.length) return pkg.highlights;
    if (pkg.rules_summary) {
      return Object.entries(pkg.rules_summary).map(([k, v]) => `${String(v)} ${k.replaceAll('_', ' ')}`);
    }
    return ['Full package details after registration'];
  }

  function displayPrice(pkg: PublicPackage) {
    let amount = Number(pkg.price_amount) || 0;
    if (period === 'annual') {
      // Marketing annual display: ~20% off monthly equivalent shown as monthly rate
      amount = Math.round(amount * 0.8);
    }
    return formatAmount(amount, pkg.price_currency || 'THB');
  }

  function mapPkg(pkg: PublicPackage, i: number, quoteOnly: boolean): DisplayTier {
    const featured = !quoteOnly && i === 2;
    const blurbs = [
      msg().pricing_blurb_starter,
      msg().pricing_blurb_growth,
      msg().pricing_blurb_pro,
      msg().pricing_blurb_ent
    ];
    return {
      id: pkg.id,
      name: pkg.name,
      blurb: pkg.description?.split('·')[0]?.trim() || blurbs[Math.min(i, 3)] || '',
      priceLabel: displayPrice(pkg),
      periodSuffix: '/mo',
      features: featuresFromPkg(pkg),
      cta: quoteOnly
        ? msg().pricing_request_quote
        : `${msg().pricing_choose} ${pkg.name}`,
      ctaHref: quoteOnly
        ? `${base}/contact?kind=contact&package_interest_id=${encodeURIComponent(pkg.id)}&use_case=${encodeURIComponent('Dedicated VM quote / capacity check: ' + pkg.name)}`
        : registerHref({ package_id: pkg.id }),
      featured,
      badge: featured ? msg().pricing_most_popular : undefined,
      mutedCta: !quoteOnly && i === 0,
      quoteOnly,
      isCustom: quoteOnly
    };
  }

  const sharedTiers = $derived.by((): DisplayTier[] => {
    void $t;
    const shared = packages
      .filter(
        (p) =>
          (p.purchase_mode || 'self_serve') === 'self_serve' &&
          (p.deployment || 'shared_cloud') === 'shared_cloud'
      )
      .sort((a, b) => (Number(a.price_amount) || 0) - (Number(b.price_amount) || 0));
    if (shared.length) return shared.map((p, i) => mapPkg(p, i, false));
    // Fallback Montti Shared Cloud sheet when catalog empty
    const annual = period === 'annual';
    return [
      {
        id: 'pkg-shared-launch',
        name: 'Launch',
        blurb: 'Shared Cloud',
        priceLabel: annual ? '฿400' : '฿500',
        periodSuffix: '/mo',
        features: ['1 concurrent voice', '2 AI avatars', 'Unlimited KM (up to 1 GB storage)', 'BYOK AI'],
        cta: `${msg().pricing_choose} Launch`,
        ctaHref: registerHref({ package_id: 'pkg-shared-launch' }),
        mutedCta: true
      },
      {
        id: 'pkg-shared-starter',
        name: 'Starter',
        blurb: 'Shared Cloud',
        priceLabel: annual ? '฿720' : '฿900',
        periodSuffix: '/mo',
        features: ['2 concurrent voice', '5 AI avatars', 'Unlimited KM (up to 5 GB storage)', 'BYOK AI'],
        cta: `${msg().pricing_choose} Starter`,
        ctaHref: registerHref({ package_id: 'pkg-shared-starter' })
      },
      {
        id: 'pkg-shared-growth',
        name: 'Growth',
        blurb: 'Shared Cloud',
        priceLabel: annual ? '฿1,200' : '฿1,500',
        periodSuffix: '/mo',
        features: ['4 concurrent voice', '10 AI avatars', 'Unlimited KM (up to 10 GB storage)', 'BYOK AI'],
        cta: `${msg().pricing_choose} Growth`,
        ctaHref: registerHref({ package_id: 'pkg-shared-growth' }),
        featured: true,
        badge: msg().pricing_most_popular
      },
      {
        id: 'pkg-shared-business',
        name: 'Business',
        blurb: 'Shared Cloud',
        priceLabel: annual ? '฿1,600' : '฿2,000',
        periodSuffix: '/mo',
        features: ['6 concurrent voice', '20 AI avatars', 'Unlimited KM (up to 20 GB storage)', 'BYOK AI'],
        cta: `${msg().pricing_choose} Business`,
        ctaHref: registerHref({ package_id: 'pkg-shared-business' })
      }
    ];
  });

  const dedicatedTiers = $derived.by((): DisplayTier[] => {
    void $t;
    const dedicated = packages
      .filter(
        (p) => p.purchase_mode === 'quote' || p.deployment === 'dedicated_vm'
      )
      .sort((a, b) => (Number(a.price_amount) || 0) - (Number(b.price_amount) || 0));
    return dedicated.map((p, i) => mapPkg(p, i, true));
  });

  function onCta(tier: DisplayTier) {
    void postFunnelEvent({
      event_name: tier.quoteOnly || tier.isCustom ? 'cta_click' : 'register_start',
      page_path: $page.url.pathname,
      cta_id: `pricing_${tier.id}_${period}`
    }).catch(() => {});
  }

  const assurances = $derived([
    { key: 'clock', title: $t.pricing_no_setup_t, body: $t.pricing_no_setup_b },
    { key: 'refresh', title: $t.pricing_cancel_t, body: $t.pricing_cancel_b },
    { key: 'shield', title: $t.pricing_secure_t, body: $t.pricing_secure_b },
    { key: 'headset', title: $t.pricing_support_t, body: $t.pricing_support_b }
  ]);
</script>

<svelte:head>
  <title>{$t.pricing_title}</title>
  <meta
    name="description"
    content={$t.pricing_p}
  />
</svelte:head>

<section class="pricing-page">
  <div class="glow" aria-hidden="true"></div>

  <div class="intro">
    <h1>{$t.pricing_h1}</h1>
    <p>{$t.pricing_p}</p>
  </div>

  <div class="period-toggle" role="group" aria-label="Billing period">
    <button
      type="button"
      class:active={period === 'monthly'}
      onclick={() => (period = 'monthly')}
    >
      {$t.pricing_monthly}
    </button>
    <button
      type="button"
      class:active={period === 'annual'}
      onclick={() => (period = 'annual')}
    >
      {$t.pricing_annual} <span class="save">{$t.pricing_save}</span>
    </button>
  </div>

  {#if loading}
    <div class="status-card">{$t.pricing_loading}</div>
  {:else if error && packages.length === 0}
    <div class="status-card warn">
      <p>{error}</p>
      <p class="muted">{$t.pricing_error_hint}</p>
    </div>
  {/if}

  <div class="catalog-block">
    <div class="catalog-head">
      <h2>{$t.pricing_shared_h2}</h2>
      <p>{$t.pricing_shared_p}</p>
    </div>
    <div
      class="tier-grid"
      class:cols-4={sharedTiers.length >= 4}
      class:cols-3={sharedTiers.length === 3}
    >
      {#each sharedTiers as tier (tier.id)}
        <article class="tier-card" class:featured={tier.featured}>
          {#if tier.badge}
            <span class="badge">{tier.badge}</span>
          {/if}
          <h2>{tier.name}</h2>
          <p class="blurb">{tier.blurb}</p>
          <div class="price-row">
            <strong>{tier.priceLabel}</strong>
            <span class="suffix">{tier.periodSuffix}</span>
          </div>
          <ul>
            {#each tier.features as line}
              <li>
                <span class="check" aria-hidden="true">✓</span>
                {line}
              </li>
            {/each}
          </ul>
          <a
            class="tier-cta"
            class:primary={tier.featured || !tier.mutedCta}
            class:outline={tier.mutedCta}
            href={tier.ctaHref}
            onclick={() => onCta(tier)}
          >
            {tier.cta}
          </a>
        </article>
      {/each}
    </div>
    <p class="catalog-note">{$t.pricing_catalog_note}</p>
  </div>

  {#if dedicatedTiers.length > 0}
    <div class="catalog-block dedicated">
      <div class="catalog-head">
        <h2>{$t.pricing_dedicated_h2}</h2>
        <p>{$t.pricing_dedicated_p}</p>
      </div>
      <div
        class="tier-grid"
        class:cols-4={dedicatedTiers.length >= 4}
        class:cols-3={dedicatedTiers.length === 3}
      >
        {#each dedicatedTiers as tier (tier.id)}
          <article class="tier-card quote-card">
            <h2>{tier.name}</h2>
            <p class="blurb">{tier.blurb}</p>
            <div class="price-row">
              <strong>{tier.priceLabel}</strong>
              <span class="suffix">{tier.periodSuffix}</span>
            </div>
            <ul>
              {#each tier.features as line}
                <li>
                  <span class="check" aria-hidden="true">✓</span>
                  {line}
                </li>
              {/each}
            </ul>
            <a
              class="tier-cta outline"
              href={tier.ctaHref}
              onclick={() => onCta(tier)}
            >
              {tier.cta}
            </a>
          </article>
        {/each}
      </div>
    </div>
  {/if}

  <div class="assurances">
    {#each assurances as item}
      <div class="assurance">
        <div class="a-icon" aria-hidden="true">
          {#if item.key === 'clock'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="8" />
              <path d="M12 8v4.2l2.8 1.6" />
            </svg>
          {:else if item.key === 'refresh'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 12a8 8 0 0 1 13.7-5.6L20 9" />
              <path d="M20 4v5h-5" />
              <path d="M20 12a8 8 0 0 1-13.7 5.6L4 15" />
              <path d="M4 20v-5h5" />
            </svg>
          {:else if item.key === 'shield'}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 3 5 6v5.5c0 4.4 2.9 7.7 7 9 4.1-1.3 7-4.6 7-9V6l-7-3Z" />
              <path d="m9 12 2 2 4-4" />
            </svg>
          {:else}
            <svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <path d="M4 14v-1a8 8 0 0 1 16 0v1" />
              <path d="M4 14a2 2 0 0 0 2 2h1v-4H6a2 2 0 0 0-2 2Z" />
              <path d="M20 14a2 2 0 0 1-2 2h-1v-4h1a2 2 0 0 1 2 2Z" />
              <path d="M17 16v1a3 3 0 0 1-3 3h-1" />
            </svg>
          {/if}
        </div>
        <div class="a-copy">
          <strong>{item.title}</strong>
          <span>{item.body}</span>
        </div>
      </div>
    {/each}
  </div>

  <p class="legal">{$t.pricing_legal}</p>
</section>

<style>
  .pricing-page {
    position: relative;
    overflow: hidden;
    padding: 52px 24px 72px;
    background:
      radial-gradient(circle at 50% 0%, rgb(30 100 255 / 12%), transparent 32%),
      linear-gradient(180deg, #020713 0%, #031025 50%, #020713 100%);
  }

  .glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    opacity: 0.22;
    background-image:
      linear-gradient(rgb(30 90 180 / 12%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(30 90 180 / 12%) 1px, transparent 1px);
    background-size: 56px 56px;
    mask-image: linear-gradient(180deg, #000, transparent 65%);
  }

  .intro {
    position: relative;
    z-index: 1;
    text-align: center;
    width: min(720px, 100%);
    margin: 0 auto 22px;
  }

  .intro h1 {
    margin: 0 0 12px;
    font-size: clamp(1.9rem, 3.6vw, 2.6rem);
    letter-spacing: -0.03em;
    font-weight: 750;
  }

  .intro p {
    margin: 0;
    color: #93a6c4;
    font-size: 1.05rem;
    line-height: 1.55;
  }

  .period-toggle {
    position: relative;
    z-index: 1;
    width: fit-content;
    margin: 0 auto 32px;
    display: inline-flex;
    padding: 4px;
    border-radius: 999px;
    border: 1px solid rgb(70 120 190 / 28%);
    background: rgb(8 16 34 / 80%);
    gap: 2px;
  }

  .period-toggle button {
    border: 0;
    background: transparent;
    color: #9aacc8;
    font-size: 13px;
    font-weight: 650;
    padding: 9px 18px;
    border-radius: 999px;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 8px;
  }

  .period-toggle button.active {
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    box-shadow: 0 8px 20px rgb(31 123 255 / 28%);
  }

  .save {
    font-size: 11px;
    font-weight: 700;
    color: #7ec0ff;
  }

  .period-toggle button.active .save {
    color: #d6ebff;
  }

  .status-card {
    position: relative;
    z-index: 1;
    width: min(720px, 100%);
    margin: 0 auto 18px;
    padding: 14px 16px;
    border-radius: 12px;
    border: 1px solid rgb(70 120 190 / 22%);
    background: rgb(8 16 34 / 75%);
    color: #a8b8d2;
    font-size: 13px;
    text-align: center;
  }

  .status-card.warn {
    border-color: rgb(240 184 63 / 30%);
  }

  .status-card p {
    margin: 0 0 6px;
  }

  .status-card .muted {
    color: #7f93b3;
    margin: 0;
  }

  .tier-grid {
    width: 100%;
    margin: 0 auto;
    display: grid;
    gap: 14px;
    align-items: stretch;
  }

  .tier-grid.cols-4 {
    grid-template-columns: repeat(4, minmax(0, 1fr));
  }

  .tier-grid.cols-3 {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .tier-grid:not(.cols-4):not(.cols-3) {
    grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  }

  .tier-card {
    position: relative;
    display: flex;
    flex-direction: column;
    padding: 24px 20px 20px;
    border-radius: 16px;
    border: 1px solid rgb(70 120 190 / 22%);
    background: linear-gradient(165deg, rgb(10 22 48 / 90%), rgb(5 12 28 / 94%));
    box-shadow:
      0 16px 40px rgb(0 10 30 / 28%),
      inset 0 1px 0 rgb(255 255 255 / 3%);
  }

  .tier-card.featured {
    border-color: rgb(70 150 255 / 50%);
    box-shadow:
      0 20px 50px rgb(20 70 180 / 22%),
      0 0 0 1px rgb(60 140 255 / 18%),
      inset 0 1px 0 rgb(140 190 255 / 10%);
    background: linear-gradient(165deg, rgb(14 36 80 / 92%), rgb(6 16 40 / 96%));
  }

  .badge {
    position: absolute;
    top: 14px;
    right: 14px;
    border-radius: 999px;
    padding: 4px 10px;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.04em;
    color: #fff;
    background: linear-gradient(100deg, #1f7bff, #3d7dff);
  }

  .tier-card h2 {
    margin: 0 0 6px;
    font-size: 1.2rem;
    letter-spacing: -0.01em;
    padding-right: 88px;
  }

  .blurb {
    margin: 0 0 16px;
    color: #8ea0bd;
    font-size: 13px;
  }

  .price-row {
    display: flex;
    align-items: baseline;
    gap: 4px;
    margin-bottom: 18px;
  }

  .price-row strong {
    font-size: 1.85rem;
    letter-spacing: -0.03em;
    color: #4ea3ff;
    font-weight: 750;
  }

  .price-row .custom-price {
    color: #e8eefc;
  }

  .suffix {
    color: #8ea0bd;
    font-size: 14px;
    font-weight: 500;
  }

  .tier-card ul {
    list-style: none;
    margin: 0 0 20px;
    padding: 0;
    display: grid;
    gap: 10px;
    flex: 1;
  }

  .tier-card li {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    color: #c4d2e8;
    font-size: 13px;
    line-height: 1.4;
  }

  .check {
    flex: 0 0 18px;
    width: 18px;
    height: 18px;
    min-width: 18px;
    min-height: 18px;
    aspect-ratio: 1 / 1;
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    font-size: 10px;
    font-weight: 700;
    line-height: 0;
    color: #fff;
    background: #1f6fff;
    border: none;
    margin-top: 1px;
  }

  .tier-cta {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    border-radius: 10px;
    padding: 11px 14px;
    font-weight: 700;
    font-size: 14px;
    text-decoration: none;
    transition: filter 0.15s ease, transform 0.15s ease;
  }

  .tier-cta:hover {
    filter: brightness(1.06);
    transform: translateY(-1px);
  }

  .tier-cta.primary {
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    box-shadow: 0 10px 24px rgb(31 123 255 / 28%);
  }

  .tier-cta.outline {
    background: rgb(10 20 42 / 70%);
    border: 1px solid rgb(80 130 200 / 35%);
    color: #e8eefc;
  }

  .catalog-block {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 0 auto 36px;
  }

  .catalog-block.dedicated {
    margin-top: 8px;
    padding-top: 28px;
    border-top: 1px solid rgb(70 120 190 / 18%);
  }

  .catalog-head {
    text-align: center;
    margin-bottom: 20px;
  }

  .catalog-head h2 {
    margin: 0 0 8px;
    font-size: clamp(1.2rem, 2vw, 1.45rem);
    letter-spacing: -0.02em;
  }

  .catalog-head p {
    margin: 0 auto;
    max-width: 62ch;
    color: #93a6c4;
    font-size: 14px;
    line-height: 1.55;
  }

  .quote-card {
    border-style: dashed;
  }

  .catalog-note {
    text-align: center;
    margin: 16px auto 0;
    color: #6f84a8;
    font-size: 12px;
  }

  .assurances {
    position: relative;
    z-index: 1;
    width: min(900px, 100%);
    margin: 36px auto 0;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 16px;
  }

  .assurance {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 12px;
  }

  /* Force a perfect solid circle badge (not an oval / incomplete ring). */
  .a-icon {
    flex: 0 0 56px;
    width: 56px;
    height: 56px;
    min-width: 56px;
    min-height: 56px;
    max-width: 56px;
    max-height: 56px;
    aspect-ratio: 1 / 1;
    border-radius: 9999px;
    display: flex;
    align-items: center;
    justify-content: center;
    box-sizing: border-box;
    overflow: hidden;
    line-height: 0;
    color: #ffffff;
    background-color: #1f6fff;
    background-image: radial-gradient(circle at 32% 28%, #4d93ff 0%, #1f6fff 42%, #1554c9 100%);
    border: 0;
    outline: none;
    box-shadow:
      0 0 0 1px rgb(120 180 255 / 18%),
      0 10px 22px rgb(20 70 180 / 30%);
  }

  .a-icon svg {
    display: block;
    width: 22px;
    height: 22px;
    flex: 0 0 auto;
  }

  .a-copy {
    display: grid;
    gap: 4px;
  }

  .a-copy strong {
    display: block;
    font-size: 14px;
    font-weight: 700;
    color: #f0f5ff;
  }

  .a-copy span {
    display: block;
    font-size: 12px;
    color: #8ea0bd;
    line-height: 1.4;
  }

  .legal {
    position: relative;
    z-index: 1;
    width: min(720px, 100%);
    margin: 28px auto 0;
    text-align: center;
    color: #6a7d9c;
    font-size: 12px;
    line-height: 1.5;
  }

  @media (max-width: 980px) {
    .tier-grid.cols-4,
    .tier-grid.cols-3 {
      grid-template-columns: 1fr 1fr;
    }

    .assurances {
      grid-template-columns: 1fr 1fr;
    }
  }

  @media (max-width: 600px) {
    .pricing-page {
      padding: 32px 16px 48px;
    }

    .tier-grid.cols-4,
    .tier-grid.cols-3,
    .tier-grid:not(.cols-4):not(.cols-3) {
      grid-template-columns: 1fr;
    }

    .assurances {
      grid-template-columns: 1fr;
    }

    .tier-card h2 {
      padding-right: 0;
    }
  }
</style>
