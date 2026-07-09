<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { getStoredUser, hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import {
    getKYCProfile,
    updateKYCProfile,
    uploadKYCPhoto,
    uploadKYCDocument,
    submitKYC,
    type KYCProfile
  } from '$lib/api/kyc';

  let profile = $state<KYCProfile | null>(null);
  let contactName = $state('');
  let contactPhone = $state('');
  let contactAddress = $state('');
  let loading = $state(true);
  let saving = $state(false);
  let submitting = $state(false);

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/backoffice`)}`);
      return;
    }
    const user = getStoredUser();
    contactName = user?.display_name ?? '';
    await load();
  });

  async function load() {
    loading = true;
    try {
      profile = await getKYCProfile();
      contactName = profile.contact_name || contactName;
      contactPhone = profile.contact_phone;
      contactAddress = profile.contact_address;
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load KYC profile');
    } finally {
      loading = false;
    }
  }

  async function saveContact(e: Event) {
    e.preventDefault();
    saving = true;
    try {
      profile = await updateKYCProfile({
        contact_name: contactName.trim(),
        contact_phone: contactPhone.trim(),
        contact_address: contactAddress.trim()
      });
      feedback.success('Contact details saved');
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function onPhotoChange(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      await uploadKYCPhoto(file);
      await load();
      feedback.success('Photo uploaded');
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Photo upload failed');
    }
  }

  async function onDocChange(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      await uploadKYCDocument(file);
      await load();
      feedback.success('Document uploaded');
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Document upload failed');
    }
  }

  async function submitForReview() {
    submitting = true;
    try {
      profile = await submitKYC();
      feedback.success('KYC package submitted for platform review');
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Submit failed');
    } finally {
      submitting = false;
    }
  }
</script>

<div class="login-wrap" style="align-items:flex-start;padding-top:32px">
  <div class="card login-card" style="width:min(640px,100%)">
    <div style="display:flex;justify-content:space-between;gap:12px;align-items:center;margin-bottom:16px">
      <div>
        <h1 style="margin:0;font-size:22px">Tenant backoffice</h1>
        <p style="margin:4px 0 0;color:var(--muted);font-size:13px">
          Submit verification evidence while your workspace is <strong>pending KYC</strong>.
        </p>
      </div>
      <a class="link" href="{base}/login">Account</a>
    </div>

    {#if loading}
      <p style="color:var(--muted)">Loading…</p>
    {:else}
      <form onsubmit={saveContact}>
        <h2 style="font-size:16px;margin:0 0 12px">Contact information</h2>
        <div class="field">
          <label for="contact-name">Contact name</label>
          <input id="contact-name" bind:value={contactName} disabled={saving} />
        </div>
        <div class="field">
          <label for="contact-phone">Phone</label>
          <input id="contact-phone" bind:value={contactPhone} disabled={saving} />
        </div>
        <div class="field">
          <label for="contact-address">Business address</label>
          <textarea id="contact-address" bind:value={contactAddress} disabled={saving} rows="3" style="border:1px solid rgb(70 132 190 / 24%);border-radius:10px;background:rgb(3 11 23 / 88%);color:var(--ink);padding:10px 12px"></textarea>
        </div>
        <button class="btn" type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save contact info'}</button>
      </form>

      <hr style="border:none;border-top:1px solid var(--line);margin:24px 0" />

      <h2 style="font-size:16px;margin:0 0 12px">Representative photo</h2>
      {#if profile?.photo_url}
        <img src={profile.photo_url} alt="Representative" style="max-width:160px;border-radius:12px;margin-bottom:12px" />
      {/if}
      <input type="file" accept="image/*" onchange={onPhotoChange} />

      <hr style="border:none;border-top:1px solid var(--line);margin:24px 0" />

      <h2 style="font-size:16px;margin:0 0 12px">Business registration documents</h2>
      {#if profile?.documents?.length}
        <ul style="margin:0 0 12px;padding-left:18px;font-size:13px">
          {#each profile.documents as doc}
            <li><a class="link" href={doc.url} target="_blank" rel="noreferrer">Document</a></li>
          {/each}
        </ul>
      {/if}
      <input type="file" accept="image/*,application/pdf" onchange={onDocChange} />

      <div style="margin-top:24px;display:flex;gap:10px;flex-wrap:wrap;align-items:center">
        <button class="btn" type="button" disabled={submitting || profile?.status === 'submitted'} onclick={submitForReview}>
          {profile?.status === 'submitted' ? 'Submitted for review' : submitting ? 'Submitting…' : 'Submit for KYC review'}
        </button>
        <span style="color:var(--muted);font-size:12px">Status: {profile?.status ?? 'draft'}</span>
      </div>
    {/if}
  </div>
</div>