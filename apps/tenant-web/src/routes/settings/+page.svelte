<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import {
    getSettings,
    putSettings,
    getUsage,
    putCallLimits,
    type TenantSettings,
    type UsageSnapshot
  } from '$lib/api/settings';

  let settings = $state<TenantSettings | null>(null);
  let usage = $state<UsageSnapshot | null>(null);

  let displayName = $state('');
  let locale = $state('en');
  let timezone = $state('Asia/Bangkok');
  let aiReplyLocale = $state('');
  let tierLabel = $state('');
  let groupLabel = $state('');

  let maxPerCall = $state(0);
  let maxPerDay = $state(0);

  let loading = $state(true);
  let savingSettings = $state(false);
  let savingLimits = $state(false);

  /** Simple EN/TH labels for this page only (full portal i18n is out of scope). */
  const uiLocale = $derived(locale === 'th' ? 'th' : 'en');
  const t = $derived(
    uiLocale === 'th'
      ? {
          title: 'การตั้งค่า',
          subtitle: 'ภาษา เขตเวลา การใช้งาน และขีดจำกัดการโทร',
          workspace: 'พื้นที่ทำงาน',
          displayName: 'ชื่อแสดง',
          locale: 'ภาษาพอร์ทัล',
          timezone: 'เขตเวลา',
          aiLocale: 'ภาษาที่ AI ตอบ (ว่าง = ตามผู้โทร)',
          saveSettings: 'บันทึกการตั้งค่า',
          usage: 'การใช้งานแพ็กเกจ',
          noPackage: 'ยังไม่มีแพ็กเกจที่ใช้งาน',
          callLimits: 'ขีดจำกัดการโทร',
          maxPerCall: 'นาทีสูงสุดต่อสาย (0 = ไม่จำกัด)',
          maxPerDay: 'นาทีสูงสุดต่อวัน (0 = ไม่จำกัด)',
          packageCeiling: 'เพดานแพ็กเกจ (นาที/เดือน)',
          dailyUsed: 'ใช้วันนี้',
          saveLimits: 'บันทึกขีดจำกัด',
          labels: 'ป้ายกำกับ (ชั่วคราว)',
          tier: 'ระดับลูกค้า (โน้ต)',
          group: 'กลุ่ม (โน้ต)',
          labelsHelp: 'โน้ตภายใน — จัดการระดับลูกค้าที่เมนู Tiers',
          i18nNote: 'แปลทั้งพอร์ทัลยังไม่อยู่ในสโคป — หน้านี้เท่านั้น'
        }
      : {
          title: 'Settings',
          subtitle: 'Locale, timezone, package usage, and call-time caps',
          workspace: 'Workspace',
          displayName: 'Display name',
          locale: 'Portal locale',
          timezone: 'Timezone',
          aiLocale: 'AI reply language (empty = follow caller)',
          saveSettings: 'Save settings',
          usage: 'Package usage',
          noPackage: 'No active package entitlement',
          callLimits: 'Call limits',
          maxPerCall: 'Max minutes per call (0 = unset)',
          maxPerDay: 'Max call minutes per day (0 = unset)',
          packageCeiling: 'Package ceiling (min/month)',
          dailyUsed: 'Used today',
          saveLimits: 'Save call limits',
          labels: 'Labels (scaffold)',
          tier: 'User tier label',
          group: 'User group label',
          labelsHelp: 'Ops notes — manage structured tiers under Tiers (menu)',
          i18nNote: 'Full portal i18n is out of scope — settings page only'
        }
  );

  const timezones = [
    'Asia/Bangkok',
    'Asia/Singapore',
    'Asia/Tokyo',
    'UTC',
    'America/New_York',
    'Europe/London',
    'Australia/Sydney'
  ];

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/settings`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    try {
      const [s, u] = await Promise.all([getSettings(), getUsage()]);
      settings = s;
      usage = u;
      displayName = s.display_name || '';
      locale = s.locale || 'en';
      timezone = s.timezone || 'Asia/Bangkok';
      aiReplyLocale = s.ai_reply_locale || '';
      tierLabel = s.user_tier_label || '';
      groupLabel = s.user_group_label || '';
      const lim = u.call_limits;
      maxPerCall = lim?.max_minutes_per_call ?? 0;
      maxPerDay = lim?.max_call_minutes_per_day ?? 0;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load settings');
    } finally {
      loading = false;
    }
  }

  async function saveWorkspace() {
    savingSettings = true;
    try {
      settings = await putSettings({
        display_name: displayName,
        locale,
        timezone,
        ai_reply_locale: aiReplyLocale,
        user_tier_label: tierLabel,
        user_group_label: groupLabel
      });
      feedback.success(uiLocale === 'th' ? 'บันทึกแล้ว' : 'Settings saved');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      savingSettings = false;
    }
  }

  async function saveLimits() {
    savingLimits = true;
    try {
      const lim = await putCallLimits({
        max_minutes_per_call: Number(maxPerCall) || 0,
        max_call_minutes_per_day: Number(maxPerDay) || 0
      });
      maxPerCall = lim.max_minutes_per_call;
      maxPerDay = lim.max_call_minutes_per_day;
      usage = await getUsage();
      feedback.success(uiLocale === 'th' ? 'บันทึกขีดจำกัดแล้ว' : 'Call limits saved');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      savingLimits = false;
    }
  }

  function meter(used: number, limit: number | undefined | null) {
    if (limit == null || limit <= 0) return `${used} / —`;
    return `${used} / ${limit}`;
  }

  function barPct(used: number, limit: number | undefined | null) {
    if (limit == null || limit <= 0) return 0;
    return Math.min(100, Math.round((used / limit) * 100));
  }
</script>

<div class="page-wrap">
<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">{t.title}</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">{t.subtitle}</p>
    <p style="margin:4px 0 0;color:var(--muted);font-size:11px">{t.i18nNote}</p>
  </div>
</div>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  <div class="grid">
    <section class="card">
      <h2>{t.workspace}</h2>
      <label>
        <span>{t.displayName}</span>
        <input type="text" bind:value={displayName} placeholder="Acme Support" />
      </label>
      <label>
        <span>{t.locale}</span>
        <select bind:value={locale}>
          <option value="en">English (en)</option>
          <option value="th">ไทย (th)</option>
        </select>
      </label>
      <label>
        <span>{t.timezone}</span>
        <select bind:value={timezone}>
          {#each timezones as tz}
            <option value={tz}>{tz}</option>
          {/each}
          {#if timezone && !timezones.includes(timezone)}
            <option value={timezone}>{timezone}</option>
          {/if}
        </select>
      </label>
      <label>
        <span>{t.aiLocale}</span>
        <select bind:value={aiReplyLocale}>
          <option value="">Auto / follow caller</option>
          <option value="en">English</option>
          <option value="th">ไทย</option>
        </select>
      </label>
      <button class="btn" type="button" disabled={savingSettings} onclick={saveWorkspace}>
        {savingSettings ? '…' : t.saveSettings}
      </button>
    </section>

    <section class="card">
      <h2>{t.usage}</h2>
      {#if !usage || usage.status === 'none' || !usage.limits}
        <p style="color:var(--muted);font-size:13px">{t.noPackage}</p>
        {#if usage?.package}
          <p style="font-size:13px">{usage.package.name}</p>
        {/if}
      {:else}
        <p style="font-size:13px;margin:0 0 12px">
          <strong>{usage.package?.name ?? 'Package'}</strong>
          · period {usage.period}
        </p>
        <div class="meter">
          <div class="meter-label">AI employees · {meter(usage.usage.ai_employees, usage.limits.max_ai_employees)}</div>
          <div class="bar"><div class="fill" style="width:{barPct(usage.usage.ai_employees, usage.limits.max_ai_employees)}%"></div></div>
        </div>
        <div class="meter">
          <div class="meter-label">
            Monthly call minutes · {meter(usage.usage.monthly_call_minutes, usage.limits.max_monthly_call_minutes)}
          </div>
          <div class="bar"><div class="fill" style="width:{barPct(usage.usage.monthly_call_minutes, usage.limits.max_monthly_call_minutes)}%"></div></div>
        </div>
        <div class="meter">
          <div class="meter-label">KM documents · {meter(usage.usage.km_documents, usage.limits.max_km_documents)}</div>
          <div class="bar"><div class="fill" style="width:{barPct(usage.usage.km_documents, usage.limits.max_km_documents)}%"></div></div>
        </div>
        <div class="meter">
          <div class="meter-label">
            Concurrent calls · {meter(usage.usage.concurrent_calls, usage.limits.max_concurrent_calls)}
          </div>
          <div class="bar"><div class="fill" style="width:{barPct(usage.usage.concurrent_calls, usage.limits.max_concurrent_calls)}%"></div></div>
        </div>
        {#if usage.daily_usage}
          <p style="font-size:12px;color:var(--muted);margin-top:12px">
            {t.dailyUsed}: {usage.daily_usage.call_minutes} min
            {#if usage.daily_usage.timezone}
              ({usage.daily_usage.timezone})
            {/if}
          </p>
        {/if}
      {/if}
    </section>

    <section class="card">
      <h2>{t.callLimits}</h2>
      {#if usage?.limits?.max_monthly_call_minutes}
        <p style="font-size:12px;color:var(--muted);margin:0 0 12px">
          {t.packageCeiling}: {usage.limits.max_monthly_call_minutes}
          (operational caps are clamped ≤ package)
        </p>
      {/if}
      <label>
        <span>{t.maxPerCall}</span>
        <input type="number" min="0" bind:value={maxPerCall} />
      </label>
      <label>
        <span>{t.maxPerDay}</span>
        <input type="number" min="0" bind:value={maxPerDay} />
      </label>
      <button class="btn" type="button" disabled={savingLimits} onclick={saveLimits}>
        {savingLimits ? '…' : t.saveLimits}
      </button>
    </section>

    <section class="card">
      <h2>{t.labels}</h2>
      <p style="font-size:12px;color:var(--muted);margin:0 0 12px">
        {t.labelsHelp}
        <a class="link" href="{base}/tiers" style="margin-left:6px">Open Tiers →</a>
      </p>
      <label>
        <span>{t.tier}</span>
        <input type="text" bind:value={tierLabel} placeholder="VIP / standard" />
      </label>
      <label>
        <span>{t.group}</span>
        <input type="text" bind:value={groupLabel} placeholder="retail / enterprise" />
      </label>
      <button class="btn ghost" type="button" disabled={savingSettings} onclick={saveWorkspace}>
        {savingSettings ? '…' : t.saveSettings}
      </button>
    </section>
  </div>
{/if}
</div>

<style>
  .page-wrap {
    max-width: 960px;
    margin: 0 auto;
    padding: 20px;
  }
  .grid {
    display: grid;
    gap: 16px;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  }
  .card {
    border: 1px solid var(--line);
    border-radius: 12px;
    padding: 16px 18px;
    background: rgb(12 18 32 / 80%);
  }
  .card h2 {
    margin: 0 0 14px;
    font-size: 15px;
    font-weight: 600;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 12px;
    font-size: 12px;
    color: var(--muted);
  }
  input,
  select {
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--line);
    background: rgb(8 12 22);
    color: inherit;
    font-size: 14px;
  }
  .meter {
    margin-bottom: 10px;
  }
  .meter-label {
    font-size: 12px;
    margin-bottom: 4px;
  }
  .bar {
    height: 6px;
    border-radius: 4px;
    background: rgb(30 40 60);
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: linear-gradient(90deg, var(--cyan), #6ee7ff);
    border-radius: 4px;
    min-width: 0;
  }
</style>
