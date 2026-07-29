<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import CustomerDesk from '$lib/components/CustomerDesk.svelte';
  import { getPublicBrand } from '$lib/api/brands';
  import {
    clearSelectedTenant,
    setSelectedTenant,
    type SelectedTenant
  } from '$lib/tenantContext';

  let ready = $state(false);
  let error = $state('');
  let tenant = $state<SelectedTenant | null>(null);
  let resolving = $state(false);

  $effect(() => {
    const s = ($page.params.slug || '').trim();
    if (!s) {
      error = 'Missing brand';
      ready = false;
      tenant = null;
      return;
    }
    let cancelled = false;
    resolving = true;
    ready = false;
    error = '';
    void (async () => {
      try {
        const brand = await getPublicBrand(s);
        if (cancelled) return;
        const selected = { id: brand.id, slug: brand.slug, name: brand.name };
        setSelectedTenant(selected);
        tenant = selected;
        ready = true;
      } catch (err) {
        if (cancelled) return;
        tenant = null;
        clearSelectedTenant();
        error = err instanceof Error ? err.message : 'Brand not found';
        ready = false;
      } finally {
        if (!cancelled) resolving = false;
      }
    })();
    return () => {
      cancelled = true;
    };
  });

  function changeTenant() {
    clearSelectedTenant();
    void goto('/');
  }
</script>

{#if error && !ready}
  <main class="tenant-picker-page">
    <h1>Brand not available</h1>
    <p class="muted">{error}</p>
    <button type="button" class="back-btn" onclick={() => { clearSelectedTenant(); void goto('/'); }}>
      ← Back to brands
    </button>
  </main>
{:else if ready && tenant}
  {#key tenant.id}
    <CustomerDesk
      tenantId={tenant.id}
      tenantSlug={tenant.slug}
      tenantName={tenant.name}
      onChangeTenant={changeTenant}
    />
  {/key}
{:else}
  <main class="tenant-picker-page muted">{resolving ? 'Loading brand…' : 'Loading…'}</main>
{/if}

<style>
  .tenant-picker-page {
    max-width: 640px;
    margin: 0 auto;
    padding: 48px 20px;
    min-height: 100vh;
  }
  .muted {
    color: var(--muted);
  }
  .back-btn {
    margin-top: 16px;
    padding: 12px 18px;
    border-radius: 14px;
    border: 1px solid var(--line);
    background: var(--panel);
    color: var(--ink);
  }
</style>
