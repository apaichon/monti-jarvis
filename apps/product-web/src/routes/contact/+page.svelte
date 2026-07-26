<script lang="ts">
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { getAttribution } from '$lib/attribution';
  import { postFunnelEvent, postLead, ApiError, type LeadKind } from '$lib/api';
  import { t, getLang, msg } from '$lib/i18n';

  let kind = $state<LeadKind>('book_demo');
  let fullName = $state('');
  let email = $state('');
  let companyName = $state('');
  let phone = $state('');
  let useCase = $state('');
  let preferredChannel = $state<'email' | 'phone' | 'line' | 'other' | ''>('email');
  let consentContact = $state(false);
  let consentMarketing = $state(false);
  let website = $state(''); // honeypot
  let submitting = $state(false);
  let error = $state('');
  let leadId = $state('');
  let deduped = $state(false);

  $effect(() => {
    const k = $page.url.searchParams.get('kind');
    if (k === 'contact' || k === 'book_demo' || k === 'newsletter') {
      kind = k;
    }
  });

  async function onSubmit(e: Event) {
    e.preventDefault();
    error = '';
    if (!email.trim()) {
      error = msg().contact_err_email;
      return;
    }
    if ((kind === 'contact' || kind === 'book_demo') && !consentContact) {
      error = msg().contact_err_consent;
      return;
    }
    if (kind === 'newsletter' && !consentMarketing) {
      error = msg().contact_err_marketing;
      return;
    }

    submitting = true;
    try {
      const attrs = getAttribution();
      const res = await postLead({
        kind,
        email: email.trim(),
        full_name: fullName.trim() || undefined,
        company_name: companyName.trim() || undefined,
        phone: phone.trim() || undefined,
        use_case: useCase.trim() || undefined,
        preferred_channel: preferredChannel || undefined,
        consent_contact: kind === 'newsletter' ? consentContact : true,
        consent_marketing: consentMarketing,
        language: getLang(),
        landing_path: $page.url.pathname,
        package_interest_id: attrs.package_id,
        website
      });
      leadId = res.lead_id;
      deduped = res.deduped;
      void postFunnelEvent({
        event_name: 'lead_submit',
        page_path: $page.url.pathname,
        cta_id: `contact_${kind}`
      }).catch(() => {});
    } catch (err) {
      error =
        err instanceof ApiError
          ? err.message || msg().contact_err_generic
          : msg().contact_err_generic;
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{$t.contact_title}</title>
</svelte:head>

<section class="section">
  <span class="badge">CONTACT</span>
  <h1 class="section-title" style="margin-top:12px">
    {kind === 'book_demo' ? $t.contact_book : kind === 'newsletter' ? $t.contact_newsletter : $t.contact_sales}
  </h1>
  <p class="section-lede">{$t.contact_lede}</p>

  {#if leadId}
    <div class="card confirm" role="status">
      <p class="success"><strong>{$t.contact_thanks}</strong></p>
      <p class="muted">
        {deduped ? $t.contact_deduped : $t.contact_received}
      </p>
      <!-- lead_id kept for ops correlation; not emphasized for visitors -->
      <p class="lead-ref muted" aria-hidden="true">{leadId}</p>
      <div class="cta-row">
        <a class="btn cyan" href="{base}/demo">{$t.contact_see_demo}</a>
        <a class="btn ghost" href="{base}/">{$t.contact_back_home}</a>
      </div>
    </div>
  {:else}
    <form class="card form" onsubmit={onSubmit}>
      <div class="kind-row" role="radiogroup" aria-label="Request type">
        <label class:active={kind === 'book_demo'}>
          <input type="radio" bind:group={kind} value="book_demo" /> {$t.contact_type_book}
        </label>
        <label class:active={kind === 'contact'}>
          <input type="radio" bind:group={kind} value="contact" /> {$t.contact_type_contact}
        </label>
      </div>

      <div class="grid-2 form-grid">
        <div class="field">
          <label for="full_name">{$t.contact_name}</label>
          <input id="full_name" name="full_name" autocomplete="name" bind:value={fullName} />
        </div>
        <div class="field">
          <label for="email">{$t.contact_email} *</label>
          <input
            id="email"
            name="email"
            type="email"
            required
            autocomplete="email"
            bind:value={email}
          />
        </div>
        <div class="field">
          <label for="company">{$t.contact_company}</label>
          <input id="company" name="company" autocomplete="organization" bind:value={companyName} />
        </div>
        <div class="field">
          <label for="phone">{$t.contact_phone}</label>
          <input id="phone" name="phone" type="tel" autocomplete="tel" bind:value={phone} />
        </div>
      </div>

      <div class="field">
        <label for="use_case">{$t.contact_usecase}</label>
        <textarea
          id="use_case"
          name="use_case"
          rows="4"
          maxlength="2000"
          bind:value={useCase}
        ></textarea>
      </div>

      <div class="field">
        <label for="channel">{$t.contact_channel}</label>
        <select id="channel" name="channel" bind:value={preferredChannel}>
          <option value="email">{$t.contact_channel_email}</option>
          <option value="phone">{$t.contact_channel_phone}</option>
          <option value="line">{$t.contact_channel_line}</option>
          <option value="other">{$t.contact_channel_other}</option>
        </select>
      </div>

      <!-- Honeypot -->
      <div class="field hp" aria-hidden="true">
        <label for="website">Website</label>
        <input id="website" name="website" tabindex="-1" autocomplete="off" bind:value={website} />
      </div>

      <label class="check">
        <input type="checkbox" bind:checked={consentContact} />
        <span>{$t.contact_consent_contact}</span>
      </label>
      <label class="check">
        <input type="checkbox" bind:checked={consentMarketing} />
        <span>{$t.contact_consent_marketing}</span>
      </label>

      {#if error}
        <p class="error" role="alert">{error}</p>
      {/if}

      <button class="btn" type="submit" disabled={submitting}>
        {submitting ? $t.contact_submitting : $t.contact_submit}
      </button>
    </form>
  {/if}
</section>

<style>
  .form {
    max-width: 720px;
  }

  .kind-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-bottom: 18px;
  }

  .kind-row label {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    border: 1px solid var(--line);
    border-radius: 999px;
    padding: 8px 14px;
    font-size: 13px;
    color: var(--muted);
    cursor: pointer;
  }

  .kind-row label.active {
    color: var(--ink);
    border-color: rgb(22 199 255 / 40%);
    background: rgb(22 199 255 / 10%);
  }

  .kind-row input {
    accent-color: var(--cyan);
  }

  .form-grid {
    margin-bottom: 4px;
  }

  .check {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    margin-bottom: 12px;
    font-size: 13px;
    color: #c4d0e4;
    line-height: 1.45;
  }

  .check input {
    margin-top: 3px;
    accent-color: var(--cyan);
  }

  .confirm {
    max-width: 640px;
  }

  .confirm .success {
    margin: 0 0 8px;
    font-size: 18px;
  }

  .lead-ref {
    margin: 12px 0 0;
    font-size: 10px;
    opacity: 0.45;
    word-break: break-all;
  }

  .cta-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    margin-top: 18px;
  }
</style>
