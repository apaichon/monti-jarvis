<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import TenantPicker from '$lib/components/TenantPicker.svelte';
  import { getPublicBrand, type PublicBrand } from '$lib/api/brands';
  import {
    clearSelectedTenant,
    getSelectedTenant,
    setSelectedTenant
  } from '$lib/tenantContext';

  let booting = $state(true);
  let error = $state('');

  onMount(async () => {
    const params = new URLSearchParams(window.location.search);
    const queryTenant = params.get('tenant_id')?.trim() || '';

    // Legacy deep link: /?tenant_id=… → resolve → /t/{slug}
    if (queryTenant) {
      try {
        const brand = await getPublicBrand(queryTenant);
        setSelectedTenant({ id: brand.id, slug: brand.slug, name: brand.name });
        await goto(`/t/${encodeURIComponent(brand.slug)}`, { replaceState: true });
        return;
      } catch (err) {
        error = err instanceof Error ? err.message : 'Brand not found';
        clearSelectedTenant();
        booting = false;
        return;
      }
    }

    // Resume previous selection from session
    const selected = getSelectedTenant();
    if (selected?.slug) {
      await goto(`/t/${encodeURIComponent(selected.slug)}`, { replaceState: true });
      return;
    }

    booting = false;
  });

  async function onSelect(brand: PublicBrand) {
    setSelectedTenant({ id: brand.id, slug: brand.slug, name: brand.name });
    await goto(`/t/${encodeURIComponent(brand.slug)}`);
  }
</script>

{#if booting}
  <main class="tenant-picker" style="padding:48px;color:var(--muted)">Loading…</main>
{:else}
  <TenantPicker {error} onSelect={onSelect} />
{/if}
