<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';
  import { listGroups, listTiers, type CustomerGroup, type CustomerTier } from '$lib/api/tiers';
  import {
    createCustomer,
    createDomainRule,
    deactivateCustomer,
    deleteDomainRule,
    importCustomers,
    listCustomers,
    listDomainRules,
    updateCustomer,
    updateDomainRule,
    type Customer,
    type CustomerImport,
    type DomainRule
  } from '$lib/api/customers';

  type Tab = 'customers' | 'import' | 'domains';

  let tab = $state<Tab>('customers');
  let customers = $state<Customer[]>([]);
  let tiers = $state<CustomerTier[]>([]);
  let groups = $state<CustomerGroup[]>([]);
  let rules = $state<DomainRule[]>([]);
  let loading = $state(true);
  let saving = $state(false);
  let query = $state('');
  let statusFilter = $state('active');
  let tierFilter = $state('');

  let editingId = $state('');
  let displayName = $state('');
  let email = $state('');
  let phone = $state('');
  let locale = $state('');
  let tierId = $state('');
  let groupIds = $state<string[]>([]);
  let source = $state('manual');
  let externalId = $state('');
  let customerStatus = $state<'active' | 'inactive'>('active');

  let selectedFile = $state<File | null>(null);
  let validatedFileKey = $state('');
  let importResult = $state<CustomerImport | null>(null);
  let importing = $state(false);

  let editingRuleId = $state('');
  let domain = $state('');
  let policy = $state<'allow' | 'deny'>('allow');
  let defaultTierId = $state('');
  let defaultGroupId = $state('');
  let ruleActive = $state(true);

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/customers`)}`);
      return;
    }
    await loadAll();
  });

  function message(err: unknown, fallback: string) {
    return err instanceof ApiError ? err.message : fallback;
  }

  async function loadAll() {
    loading = true;
    try {
      const [customerRes, tierRes, groupRes, ruleRes] = await Promise.all([
        listCustomers({ status: statusFilter }), listTiers(), listGroups(), listDomainRules()
      ]);
      customers = customerRes.customers || [];
      tiers = tierRes.tiers || [];
      groups = groupRes.groups || [];
      rules = ruleRes.rules || [];
    } catch (err) {
      feedback.error(message(err, 'Failed to load customers'));
    } finally {
      loading = false;
    }
  }

  async function searchCustomers() {
    loading = true;
    try {
      const res = await listCustomers({ q: query.trim(), status: statusFilter, tier_id: tierFilter });
      customers = res.customers || [];
    } catch (err) {
      feedback.error(message(err, 'Search failed'));
    } finally {
      loading = false;
    }
  }

  function resetCustomer() {
    editingId = '';
    displayName = '';
    email = '';
    phone = '';
    locale = '';
    tierId = '';
    groupIds = [];
    source = 'manual';
    externalId = '';
    customerStatus = 'active';
  }

  function editCustomer(customer: Customer) {
    editingId = customer.id;
    displayName = customer.display_name;
    email = customer.email || '';
    phone = customer.phone || '';
    locale = customer.locale || '';
    tierId = customer.tier_id || '';
    groupIds = [...(customer.group_ids || [])];
    source = customer.source || 'manual';
    externalId = customer.external_id || '';
    customerStatus = customer.status;
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }

  function toggleGroup(id: string, checked: boolean) {
    groupIds = checked ? [...new Set([...groupIds, id])] : groupIds.filter((item) => item !== id);
  }

  async function saveCustomer() {
    if (!displayName.trim() || (!email.trim() && !externalId.trim())) {
      feedback.error('Display name and either email or external ID are required');
      return;
    }
    saving = true;
    const body = {
      display_name: displayName.trim(), email: email.trim(), phone: phone.trim(), locale,
      tier_id: tierId, group_ids: groupIds, source: source.trim() || 'manual',
      external_id: externalId.trim(), status: customerStatus, metadata: {}
    };
    try {
      if (editingId) {
        await updateCustomer(editingId, body);
        feedback.success('Customer updated');
      } else {
        const res = await createCustomer(body);
        feedback.success(res.outcome === 'updated' ? 'Existing customer updated' : 'Customer created');
      }
      resetCustomer();
      await searchCustomers();
    } catch (err) {
      feedback.error(message(err, 'Save failed'));
    } finally {
      saving = false;
    }
  }

  async function deactivate(customer: Customer) {
    if (!confirm(`Deactivate “${customer.display_name}”?`)) return;
    try {
      await deactivateCustomer(customer.id);
      feedback.success('Customer deactivated');
      await searchCustomers();
    } catch (err) {
      feedback.error(message(err, 'Deactivate failed'));
    }
  }

  function fileKey(file: File | null) {
    return file ? `${file.name}:${file.size}:${file.lastModified}` : '';
  }

  function chooseFile(event: Event) {
    selectedFile = (event.currentTarget as HTMLInputElement).files?.[0] || null;
    validatedFileKey = '';
    importResult = null;
  }

  async function runImport(dryRun: boolean) {
    if (!selectedFile) {
      feedback.error('Choose a CSV file first');
      return;
    }
    if (!dryRun && validatedFileKey !== fileKey(selectedFile)) {
      feedback.error('Validate this file before importing');
      return;
    }
    importing = true;
    try {
      importResult = await importCustomers(selectedFile, dryRun);
      if (dryRun) {
        validatedFileKey = fileKey(selectedFile);
        feedback.success(`Validated ${importResult.accepted_rows} accepted rows`);
      } else {
        feedback.success(`Imported ${importResult.created_rows} new and ${importResult.updated_rows} updated customers`);
        validatedFileKey = '';
        await searchCustomers();
      }
    } catch (err) {
      feedback.error(message(err, dryRun ? 'Validation failed' : 'Import failed'));
    } finally {
      importing = false;
    }
  }

  function resetRule() {
    editingRuleId = '';
    domain = '';
    policy = 'allow';
    defaultTierId = '';
    defaultGroupId = '';
    ruleActive = true;
  }

  function editRule(rule: DomainRule) {
    editingRuleId = rule.id;
    domain = rule.domain;
    policy = rule.policy;
    defaultTierId = rule.default_tier_id || '';
    defaultGroupId = rule.default_group_id || '';
    ruleActive = rule.active;
  }

  async function saveRule() {
    if (!domain.trim()) {
      feedback.error('Domain is required');
      return;
    }
    saving = true;
    const body = { domain: domain.trim(), policy, default_tier_id: defaultTierId, default_group_id: defaultGroupId, active: ruleActive };
    try {
      if (editingRuleId) await updateDomainRule(editingRuleId, body);
      else await createDomainRule(body);
      feedback.success(editingRuleId ? 'Domain rule updated' : 'Domain rule created');
      resetRule();
      rules = (await listDomainRules()).rules || [];
    } catch (err) {
      feedback.error(message(err, 'Save rule failed'));
    } finally {
      saving = false;
    }
  }

  async function removeRule(rule: DomainRule) {
    if (!confirm(`Delete rule for ${rule.domain}?`)) return;
    try {
      await deleteDomainRule(rule.id);
      rules = (await listDomainRules()).rules || [];
      feedback.success('Domain rule deleted');
    } catch (err) {
      feedback.error(message(err, 'Delete failed'));
    }
  }

  function tierName(id?: string) { return tiers.find((item) => item.id === id)?.name || '—'; }
  function groupNames(ids: string[]) { return ids.map((id) => groups.find((g) => g.id === id)?.name).filter(Boolean).join(', ') || '—'; }
</script>

<div class="page-head">
  <div>
    <span class="eyebrow">IDENTITY DIRECTORY</span>
    <h1>Customers</h1>
    <p>Import and organize customer identities before customer sign-in is enabled.</p>
  </div>
  <button class="btn" type="button" onclick={() => { tab = 'customers'; resetCustomer(); }}>+ Add customer</button>
</div>

<nav class="tabs" aria-label="Customer sections">
  <button class:active={tab === 'customers'} onclick={() => (tab = 'customers')}>Customers</button>
  <button class:active={tab === 'import'} onclick={() => (tab = 'import')}>CSV imports</button>
  <button class:active={tab === 'domains'} onclick={() => (tab = 'domains')}>Domain rules</button>
</nav>

{#if tab === 'customers'}
  <section class="customer-grid">
    <div class="card editor">
      <h2>{editingId ? 'Edit customer' : 'New customer'}</h2>
      <div class="form-grid">
        <label><span>Display name *</span><input bind:value={displayName} placeholder="Jane Doe" /></label>
        <label><span>Email</span><input type="email" bind:value={email} placeholder="jane@example.com" /></label>
        <label><span>Phone</span><input bind:value={phone} placeholder="+66…" /></label>
        <label><span>Locale</span><select bind:value={locale}><option value="">Auto</option><option value="en">English</option><option value="th">ไทย</option></select></label>
        <label><span>Tier</span><select bind:value={tierId}><option value="">No tier</option>{#each tiers.filter((t) => t.active) as tier}<option value={tier.id}>{tier.name}</option>{/each}</select></label>
        <label><span>Status</span><select bind:value={customerStatus}><option value="active">Active</option><option value="inactive">Inactive</option></select></label>
        <label><span>Source</span><input bind:value={source} placeholder="manual" /></label>
        <label><span>External ID</span><input bind:value={externalId} placeholder="crm-42" /></label>
      </div>
      <fieldset>
        <legend>Groups</legend>
        <div class="checks">
          {#each groups as group}
            <label><input type="checkbox" checked={groupIds.includes(group.id)} onchange={(event) => toggleGroup(group.id, event.currentTarget.checked)} /> {group.name}</label>
          {:else}<span class="muted">No groups configured.</span>{/each}
        </div>
      </fieldset>
      <div class="actions"><button class="btn" disabled={saving} onclick={saveCustomer}>{saving ? 'Saving…' : editingId ? 'Save changes' : 'Create customer'}</button>{#if editingId}<button class="btn ghost" onclick={resetCustomer}>Cancel</button>{/if}</div>
    </div>

    <div class="directory">
      <div class="filters card">
        <input aria-label="Search customers" bind:value={query} placeholder="Search name, email, phone…" onkeydown={(event) => event.key === 'Enter' && searchCustomers()} />
        <select bind:value={statusFilter} onchange={searchCustomers}><option value="">All statuses</option><option value="active">Active</option><option value="inactive">Inactive</option></select>
        <select bind:value={tierFilter} onchange={searchCustomers}><option value="">All tiers</option>{#each tiers as tier}<option value={tier.id}>{tier.name}</option>{/each}</select>
        <button class="btn ghost" onclick={searchCustomers}>Search</button>
      </div>
      {#if loading}<div class="card state">Loading customers…</div>
      {:else if !customers.length}<div class="card state">No customers match these filters.</div>
      {:else}<div class="card table-wrap"><table><thead><tr><th>Customer</th><th>Tier / groups</th><th>Source</th><th>Status</th><th></th></tr></thead><tbody>{#each customers as customer}<tr><td><strong>{customer.display_name}</strong><small>{customer.email || customer.phone || customer.external_id}</small></td><td>{tierName(customer.tier_id)}<small>{groupNames(customer.group_ids)}</small></td><td><span class="badge">{customer.source}</span><small>{customer.external_id || '—'}</small></td><td><span class:success={customer.status === 'active'} class="status">{customer.status}</span></td><td><div class="row-actions"><button class="btn ghost" onclick={() => editCustomer(customer)}>Edit</button>{#if customer.status === 'active'}<button class="btn ghost danger" onclick={() => deactivate(customer)}>Deactivate</button>{/if}</div></td></tr>{/each}</tbody></table></div>{/if}
    </div>
  </section>
{:else if tab === 'import'}
  <section class="card import-card">
    <div class="section-title"><div><span class="eyebrow">CSV WORKFLOW</span><h2>Import customers</h2><p>Validate first. Customer records are written only after you confirm the same file.</p></div><a class="link" href="data:text/csv;charset=utf-8,display_name%2Cemail%2Cphone%2Clocale%2Ctier_slug%2Cgroup_slugs%2Csource%2Cexternal_id%0AJane%20Doe%2Cjane%40example.com%2C%2Cen%2Cvip%2Cretail%7Cbeta%2Ccsv%2Ccrm-42" download="monti-customer-import.csv">Download template</a></div>
    <div class="upload"><input type="file" accept=".csv,text/csv" onchange={chooseFile} /><span>{selectedFile ? `${selectedFile.name} · ${Math.ceil(selectedFile.size / 1024)} KB` : 'UTF-8 CSV · maximum 2 MiB / 5,000 rows'}</span></div>
    <div class="steps"><span class:done={!!selectedFile}>1 <b>Upload</b></span><i></i><span class:done={!!validatedFileKey}>2 <b>Validate</b></span><i></i><span class:done={importResult?.status === 'completed'}>3 <b>Commit</b></span></div>
    <div class="actions"><button class="btn ghost" disabled={!selectedFile || importing} onclick={() => runImport(true)}>{importing ? 'Working…' : 'Validate CSV'}</button><button class="btn" disabled={!selectedFile || importing || validatedFileKey !== fileKey(selectedFile)} onclick={() => runImport(false)}>Import accepted rows</button></div>
    {#if importResult}<div class="result"><div><strong>{importResult.accepted_rows}</strong><span>accepted</span></div><div><strong>{importResult.rejected_rows}</strong><span>rejected</span></div><div><strong>{importResult.created_rows}</strong><span>created</span></div><div><strong>{importResult.updated_rows}</strong><span>updated</span></div></div>{#if importResult.errors.length}<div class="table-wrap errors"><table><thead><tr><th>Row</th><th>Field</th><th>Issue</th></tr></thead><tbody>{#each importResult.errors as item}<tr><td>{item.row}</td><td><code>{item.field}</code></td><td>{item.message}</td></tr>{/each}</tbody></table></div>{/if}{/if}
  </section>
{:else}
  <section class="domain-grid">
    <div class="card editor"><h2>{editingRuleId ? 'Edit domain rule' : 'New domain rule'}</h2><label><span>Domain *</span><input bind:value={domain} placeholder="example.com" /></label><label><span>Policy</span><select bind:value={policy}><option value="allow">Allow</option><option value="deny">Deny</option></select></label><label><span>Default tier</span><select bind:value={defaultTierId}><option value="">No default</option>{#each tiers.filter((t) => t.active) as tier}<option value={tier.id}>{tier.name}</option>{/each}</select></label><label><span>Default group</span><select bind:value={defaultGroupId}><option value="">No default</option>{#each groups as group}<option value={group.id}>{group.name}</option>{/each}</select></label><label class="check"><input type="checkbox" bind:checked={ruleActive} /> Active</label><div class="actions"><button class="btn" disabled={saving} onclick={saveRule}>{saving ? 'Saving…' : editingRuleId ? 'Save rule' : 'Add rule'}</button>{#if editingRuleId}<button class="btn ghost" onclick={resetRule}>Cancel</button>{/if}</div></div>
    <div class="card"><h2>Domain rules</h2><p class="notice">Allow/deny policies are stored now and enforced when customer authentication ships in SPRINT-020.</p>{#if !rules.length}<p class="muted">No domain rules configured.</p>{:else}<div class="rule-list">{#each rules as rule}<article><div><strong>{rule.domain}</strong><span class:allow={rule.policy === 'allow'} class="policy">{rule.policy}</span><small>{tierName(rule.default_tier_id)} · {groups.find((g) => g.id === rule.default_group_id)?.name || 'No group'} · {rule.active ? 'active' : 'inactive'}</small></div><div class="row-actions"><button class="btn ghost" onclick={() => editRule(rule)}>Edit</button><button class="btn ghost danger" onclick={() => removeRule(rule)}>Delete</button></div></article>{/each}</div>{/if}</div>
  </section>
{/if}

<style>
  .page-head,.section-title{display:flex;justify-content:space-between;align-items:flex-start;gap:16px;flex-wrap:wrap;margin-bottom:20px}.page-head h1{margin:5px 0;font-size:28px}.page-head p,.section-title p{margin:0;color:var(--muted);font-size:13px}.eyebrow{color:var(--cyan);font-size:9px;font-weight:700;letter-spacing:.18em}.tabs{display:flex;gap:4px;border-bottom:1px solid var(--line);margin-bottom:20px}.tabs button{border:0;border-bottom:2px solid transparent;padding:11px 15px;color:var(--muted);background:transparent}.tabs button.active{color:white;border-bottom-color:var(--cyan)}.customer-grid,.domain-grid{display:grid;grid-template-columns:minmax(280px,360px) minmax(0,1fr);gap:18px;align-items:start}.editor{position:sticky;top:77px}.editor h2,.card h2{margin:0 0 16px;font-size:16px}.form-grid{display:grid;grid-template-columns:1fr 1fr;gap:0 10px}.editor label{display:grid;gap:6px;margin-bottom:11px;color:var(--muted);font-size:11px}.editor input,.editor select,.filters input,.filters select{width:100%;padding:9px 10px;border:1px solid var(--line);border-radius:9px}.editor fieldset{border:1px solid var(--line);border-radius:10px;margin:4px 0 14px;padding:10px}.editor legend{padding:0 5px;color:var(--muted);font-size:11px}.checks{display:flex;gap:9px;flex-wrap:wrap}.checks label,.check{display:flex!important;align-items:center;gap:5px;margin:0!important}.checks input,.check input{width:auto}.actions,.row-actions{display:flex;gap:8px;flex-wrap:wrap}.filters{display:grid;grid-template-columns:minmax(160px,1fr) 130px 140px auto;gap:8px;margin-bottom:14px;padding:12px}.table-wrap{overflow:auto;padding:0}.table-wrap table{width:100%;border-collapse:collapse;font-size:12px}.table-wrap th,.table-wrap td{padding:12px;text-align:left;border-bottom:1px solid var(--line);vertical-align:top}.table-wrap th{color:var(--muted);font-size:9px;text-transform:uppercase;letter-spacing:.08em}.table-wrap td small{display:block;color:var(--muted);margin-top:4px}.status,.policy{display:inline-flex;border-radius:99px;padding:3px 8px;color:var(--danger);background:rgb(255 92 122 / 10%);font-size:9px;text-transform:uppercase}.status.success,.policy.allow{color:var(--success);background:rgb(61 214 140 / 10%)}.danger{color:var(--danger)!important}.state{text-align:center;color:var(--muted);padding:42px}.import-card{max-width:920px}.section-title h2{margin:4px 0;font-size:20px}.upload{display:grid;place-items:center;gap:8px;min-height:140px;border:1px dashed rgb(60 132 255 / 42%);border-radius:13px;background:rgb(19 49 94 / 10%);padding:22px;margin-bottom:18px}.upload span,.muted{color:var(--muted);font-size:12px}.steps{display:flex;align-items:center;margin:20px 0}.steps span{display:flex;align-items:center;gap:7px;color:var(--muted);font-size:11px}.steps span:first-letter{width:24px;height:24px}.steps span.done{color:var(--success)}.steps i{height:1px;flex:1;margin:0 10px;background:var(--line)}.result{display:grid;grid-template-columns:repeat(4,1fr);gap:10px;margin-top:20px}.result div{display:grid;gap:3px;padding:14px;border:1px solid var(--line);border-radius:10px;background:rgb(6 13 26 / 62%)}.result strong{font-size:22px}.result span{color:var(--muted);font-size:10px;text-transform:uppercase}.errors{margin-top:14px;border:1px solid var(--line);border-radius:10px}.rule-list{display:grid;gap:8px}.rule-list article{display:flex;justify-content:space-between;gap:12px;padding:13px;border:1px solid var(--line);border-radius:10px}.rule-list article>div:first-child{display:grid;grid-template-columns:auto auto;gap:6px;align-items:center}.rule-list small{grid-column:1/-1;color:var(--muted)}.notice{padding:11px;border:1px solid rgb(22 199 255 / 18%);border-radius:10px;color:#9fb9d5;background:rgb(22 199 255 / 5%);font-size:11px}@media(max-width:900px){.customer-grid,.domain-grid{grid-template-columns:1fr}.editor{position:static}.filters{grid-template-columns:1fr 1fr}}@media(max-width:600px){.form-grid,.filters{grid-template-columns:1fr}.result{grid-template-columns:1fr 1fr}.table-wrap table{min-width:680px}.page-head{align-items:stretch}.page-head>.btn{width:100%}}
</style>
