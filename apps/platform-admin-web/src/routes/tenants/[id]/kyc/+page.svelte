<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { approveTenantKYC, getTenantKYC, rejectTenantKYC, type TenantKYCReview } from '$lib/api/kyc';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  const tenantId = $derived($page.params.id);

  let data = $state<TenantKYCReview | null>(null);
  let rejectReason = $state('');
  let loading = $state(true);
  let approving = $state(false);
  let rejecting = $state(false);

  const canDecide = $derived(data?.kyc.status === 'submitted' && data?.tenant.status === 'pending_kyc');

  onMount(load);

  async function load() {
    loading = true;
    try {
      data = await getTenantKYC(tenantId);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load KYC package');
    } finally {
      loading = false;
    }
  }

  async function approve() {
    if (!canDecide) return;
    approving = true;
    try {
      await approveTenantKYC(tenantId);
      feedback.success('Tenant activated — KYC approved');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Approve failed');
    } finally {
      approving = false;
    }
  }

  async function reject() {
    if (!canDecide) return;
    const reason = rejectReason.trim();
    if (!reason) {
      feedback.error('Rejection reason is required');
      return;
    }
    rejecting = true;
    try {
      await rejectTenantKYC(tenantId, reason);
      feedback.success('KYC rejected — tenant notified');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Reject failed');
    } finally {
      rejecting = false;
    }
  }
</script>

<p><a class="link" href="{base}/tenants">← Tenants</a></p>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else if data}
  <h1 style="margin:8px 0 4px;font-size:24px">KYC review — {data.tenant.slug}</h1>
  <p style="color:var(--muted);font-size:14px;margin:0 0 20px">
    Tenant: <span class="badge">{data.tenant.status}</span>
    · KYC: <span class="badge">{data.kyc.status}</span>
    {#if data.registration}
      · Registration: <span class="badge">{data.registration.status}</span>
    {/if}
  </p>

  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 12px;font-size:16px">Registration</h2>
    {#if data.registration}
      <div class="field">
        <label>Company</label>
        <div>{data.registration.company_name}</div>
      </div>
      <div class="field">
        <label>Workspace</label>
        <div><code>monti.app/{data.tenant.slug}</code></div>
      </div>
      <div class="field">
        <label>Admin email</label>
        <div>{data.registration.admin_email}</div>
      </div>
    {:else}
      <p style="color:var(--muted);margin:0">No registration record.</p>
    {/if}
  </div>

  <div class="kyc-grid" style="margin-bottom:16px">
    <div class="card">
      <h2 style="margin:0 0 12px;font-size:16px">Photo</h2>
      {#if data.kyc.photo_url}
        <img class="kyc-photo" src={data.kyc.photo_url} alt="KYC portrait" />
      {:else}
        <p style="color:var(--muted);margin:0">No photo uploaded.</p>
      {/if}
    </div>
    <div class="card">
      <h2 style="margin:0 0 12px;font-size:16px">Contact</h2>
      <div class="field">
        <label>Name</label>
        <div>{data.kyc.contact_name || '—'}</div>
      </div>
      <div class="field">
        <label>Phone</label>
        <div>{data.kyc.contact_phone || '—'}</div>
      </div>
      <div class="field" style="margin-bottom:0">
        <label>Address</label>
        <div>{data.kyc.contact_address || '—'}</div>
      </div>
    </div>
  </div>

  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 12px;font-size:16px">Documents</h2>
    {#if data.kyc.documents.length === 0}
      <p style="color:var(--muted);margin:0">No documents uploaded.</p>
    {:else}
      <ul class="doc-list">
        {#each data.kyc.documents as doc (doc.object_key)}
          <li>
            <a class="link" href={doc.url} target="_blank" rel="noopener noreferrer">
              {doc.object_key.split('/').pop()}
            </a>
          </li>
        {/each}
      </ul>
    {/if}
  </div>

  {#if canDecide}
    <div class="card" style="margin-bottom:16px">
      <h2 style="margin:0 0 12px;font-size:16px">Decision</h2>
      <div class="field">
        <label for="reject-reason">Rejection reason (required to reject)</label>
        <textarea id="reject-reason" rows="3" bind:value={rejectReason} placeholder="Explain what the tenant must fix…"></textarea>
      </div>
      <div class="actions">
        <button class="btn danger" type="button" disabled={rejecting || approving} onclick={reject}>
          {rejecting ? 'Rejecting…' : 'Reject'}
        </button>
        <button class="btn" type="button" disabled={approving || rejecting} onclick={approve}>
          {approving ? 'Approving…' : 'Approve'}
        </button>
      </div>
    </div>
  {:else if data.kyc.rejection_reason}
    <div class="card">
      <h2 style="margin:0 0 8px;font-size:16px">Last rejection</h2>
      <p style="margin:0;color:var(--muted);font-size:14px">{data.kyc.rejection_reason}</p>
    </div>
  {/if}

  <p style="margin-top:16px">
    <a class="link" href="{base}/tenants/{tenantId}/entitlement">Entitlement</a>
    ·
    <a class="link" href="{base}/tenants/{tenantId}/avatars">Avatars</a>
  </p>
{/if}

<style>
  .kyc-grid {
    display: grid;
    grid-template-columns: 200px 1fr;
    gap: 16px;
  }

  @media (max-width: 720px) {
    .kyc-grid {
      grid-template-columns: 1fr;
    }
  }

  .kyc-photo {
    width: 160px;
    height: 160px;
    object-fit: cover;
    border-radius: 12px;
    border: 1px solid rgb(0 183 255 / 35%);
  }

  .doc-list {
    margin: 0;
    padding-left: 18px;
    font-size: 14px;
  }

  .actions {
    display: flex;
    gap: 10px;
    justify-content: flex-end;
    flex-wrap: wrap;
  }
</style>