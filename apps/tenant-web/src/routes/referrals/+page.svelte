<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';
  import { getReferralCode, getReferrals, redeemReferralCode, validateReferralCode, type BonusBalance, type Referral, type ReferralCode, type ReferralRedemption } from '$lib/api/referrals';

  let code = $state<ReferralCode | null>(null);
  let referrals = $state<Referral[]>([]);
  let bonus = $state<BonusBalance[]>([]);
  let loading = $state(true);
  let redeemCode = $state('');
  let redeemBusy = $state(false);
  let redeemPreview = $state<string>('');
  let redemptions = $state<ReferralRedemption[]>([]);

  let error = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/referrals`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    error = '';
    try {
      [code, { referrals, bonus, redemptions }] = await Promise.all([getReferralCode(), getReferrals()]);
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Failed to load referral rewards';
      feedback.error(error);
    } finally {
      loading = false;
    }
  }

  async function copyLink() {
    if (!code) return;
    await navigator.clipboard?.writeText(`${window.location.origin}/register?ref=${encodeURIComponent(code.code)}`);
    feedback.success('Referral link copied');
  }

  function label(dimension: string) {
    return dimension.replaceAll('_', ' ');
  }

  async function checkCode() {
    redeemBusy = true;
    redeemPreview = '';
    try {
      const res = await validateReferralCode(redeemCode.trim());
      const parts = (res.preview_bonus || []).map((b) => `+${b.remaining} ${b.dimension.replaceAll('_', ' ')}`);
      redeemPreview = parts.length ? `Eligible: ${parts.join(' · ')}` : 'Eligible (no bonus rules configured)';
      feedback.success('Code is eligible');
    } catch (err) {
      redeemPreview = '';
      feedback.error(err instanceof ApiError ? err.message : 'Code is not eligible');
    } finally {
      redeemBusy = false;
    }
  }

  async function applyCode() {
    redeemBusy = true;
    try {
      await redeemReferralCode(redeemCode.trim());
      feedback.success('Referral code applied');
      redeemCode = '';
      redeemPreview = '';
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to redeem code');
    } finally {
      redeemBusy = false;
    }
  }
</script>

<svelte:head><title>Referral rewards | Monti Tenant</title></svelte:head>

<div class="referrals-page">
  <div class="page-head"><div><p class="eyebrow">Growth</p><h1>Referral rewards</h1><p class="muted">Invite another business and track earned bonus quota separately from your package.</p></div><span class="scope-badge">Tenant scoped</span></div>
  {#if loading}<p class="muted">Loading referral rewards...</p>
  {:else if error}<section class="card error"><strong>Unable to load referrals.</strong><span>{error}</span><button class="btn ghost" onclick={load}>Retry</button></section>
  {:else}
    <section class="card invite"><div><p class="eyebrow">Your referral link</p><strong>{code?.code ?? '—'}</strong><p class="muted">Attribution is captured once and cannot be reassigned.</p></div><button class="btn" onclick={copyLink} disabled={!code}>Copy invite link</button></section>
    <section class="card">
      <div class="section-head"><div><h2>Redeem a code</h2><p class="muted">Apply another tenant&apos;s referral code for bonus quota</p></div></div>
      <div style="display:flex;gap:10px;flex-wrap:wrap;margin-top:14px;align-items:center">
        <input style="flex:1;min-width:180px;padding:10px 12px;border-radius:9px;border:1px solid var(--line);background:transparent;color:inherit" placeholder="Enter referral code" bind:value={redeemCode} />
        <button class="btn ghost" disabled={redeemBusy || !redeemCode.trim()} onclick={() => void checkCode()}>Check</button>
        <button class="btn" disabled={redeemBusy || !redeemCode.trim()} onclick={() => void applyCode()}>Apply</button>
      </div>
      {#if redeemPreview}<p class="muted" style="margin-top:10px">{redeemPreview}</p>{/if}
      {#if redemptions.length}
        <div class="table" style="margin-top:16px">
          {#each redemptions as r (r.id || r.redemption_id || r.referral_code)}
            <div class="row"><span><strong>{r.referral_code || r.id}</strong><small>{r.status}</small></span></div>
          {/each}
        </div>
      {/if}
    </section>
    <section class="card"><div class="section-head"><div><h2>Bonus quota</h2><p class="muted">Granted, consumed, and remaining rewards</p></div><button class="btn ghost" onclick={load}>Refresh</button></div><div class="bonus-grid">{#each bonus as item (item.dimension)}<article class="bonus"><span>{label(item.dimension)}</span><strong>{item.remaining} {item.unit}</strong><small>{item.used} used · {item.granted} granted{item.expires_at ? ` · expires ${new Date(item.expires_at).toLocaleDateString()}` : ''}</small></article>{:else}<p class="muted">No bonus quota has been earned yet.</p>{/each}</div></section>
    <section class="card"><div class="section-head"><div><h2>Referral activity</h2><p class="muted">Qualification requires activation, approved KYC, and a paid non-voided order.</p></div><span>{referrals.length}</span></div>{#if referrals.length === 0}<p class="muted empty">No referrals yet.</p>{:else}<div class="table">{#each referrals as item (item.id)}<div class="row"><span><strong>{item.referred_tenant_id}</strong><small>{item.source || 'direct invite'}</small></span><b class:qualified={item.status === 'qualified'}>{item.status}</b></div>{/each}</div>{/if}</section>
  {/if}
</div>

<style>
  .referrals-page { display:grid; gap:18px; max-width:1100px; margin:0 auto; }
  .page-head,.section-head,.invite { display:flex; align-items:center; justify-content:space-between; gap:16px; }
  .page-head { align-items:start; }
  h1,h2,p { margin:0; } h1 { font-size:34px; } h2 { font-size:16px; }
  .eyebrow { margin-bottom:6px; color:var(--cyan); font-size:11px; letter-spacing:.12em; text-transform:uppercase; }
  .muted,.bonus small,.row small { color:var(--muted); } .muted { font-size:13px; }
  .scope-badge,.section-head > span { border:1px solid var(--line); border-radius:999px; padding:6px 10px; color:var(--muted); font-size:11px; }
  .invite { align-items:start; } .invite strong { display:block; margin-bottom:6px; font-size:24px; letter-spacing:.04em; }
  .btn { border:1px solid rgb(74 135 255 / 46%); border-radius:9px; padding:9px 13px; background:linear-gradient(100deg,var(--blue),var(--violet)); color:var(--ink); font-weight:650; white-space:nowrap; } .btn.ghost { background:rgb(13 23 42 / 62%); }
  .btn:disabled { opacity:.55; } .error { display:grid; gap:8px; } .error .btn { justify-self:start; }
  .bonus-grid { display:grid; grid-template-columns:repeat(3,minmax(0,1fr)); gap:10px; margin-top:18px; }
  .bonus { display:grid; gap:7px; padding:14px; border:1px solid var(--line); border-radius:10px; } .bonus span { color:var(--muted); text-transform:capitalize; } .bonus strong { font-size:22px; } .bonus small { font-size:11px; }
  .table { margin-top:16px; } .row { display:flex; justify-content:space-between; gap:12px; padding:13px 0; border-bottom:1px solid var(--line); } .row:last-child { border-bottom:0; } .row span { display:grid; gap:4px; } .row small { font-size:11px; } .row b { text-transform:capitalize; } .row b.qualified { color:var(--cyan); } .empty { padding-top:16px; }
  @media (max-width:700px) { .page-head,.invite { flex-direction:column; align-items:start; } .bonus-grid { grid-template-columns:1fr; } }
</style>
