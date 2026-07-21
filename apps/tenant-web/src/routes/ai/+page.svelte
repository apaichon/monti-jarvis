<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import { getEmbedConfig, putEmbedConfig, type EmbedConfig } from '$lib/api/embed';
  import {
    createTenantSkill,
    createTenantTool,
    deleteTenantGeminiKey,
    getTenantGeminiKey,
    getTenantPrompt,
    listTenantSkills,
    listTenantTools,
    putTenantGeminiKey,
    putTenantPrompt,
    type TenantGeminiKey,
    type TenantPrompt,
    type TenantSkill,
    type TenantTool
  } from '$lib/api/ai';

  const agents = [
    { id: 'ava', name: 'Ava' },
    { id: 'max', name: 'Max' },
    { id: 'luna', name: 'Luna' },
    { id: 'neo', name: 'Neo' }
  ];

  let loading = $state(true);
  let saving = $state(false);
  let embed = $state<EmbedConfig | null>(null);
  let keyMeta = $state<TenantGeminiKey | null>(null);
  let keyInput = $state('');
  let authRequired = $state(false);
  let agentId = $state('ava');
  let prompt = $state<TenantPrompt>({ agent_id: 'ava', enabled: true, system_prompt: '', max_length: 8000 });
  let tools = $state<TenantTool[]>([]);
  let skills = $state<TenantSkill[]>([]);
  let error = $state('');

  function report(err: unknown, fallback: string) {
    feedback.error(err instanceof ApiError ? err.message : fallback);
  }

  async function loadPrompt() {
    try {
      prompt = await getTenantPrompt(agentId);
    } catch (err) {
      report(err, 'Failed to load agent prompt');
    }
  }

  async function load() {
    loading = true;
    try {
      const [embedRow, keyRow, toolRow, skillRow] = await Promise.all([
        getEmbedConfig(),
        getTenantGeminiKey(),
        listTenantTools(),
        listTenantSkills()
      ]);
      embed = embedRow;
      authRequired = !!embedRow.auth_required;
      keyMeta = keyRow;
      tools = toolRow.tools || [];
      skills = skillRow.skills || [];
      await loadPrompt();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load AI configuration';
      report(err, error);
    } finally {
      loading = false;
    }
  }

  async function saveEmbed() {
    if (!embed) return;
    saving = true;
    try {
      embed = await putEmbedConfig({ auth_required: authRequired });
      feedback.success('Embed access settings saved');
    } catch (err) {
      report(err, 'Failed to save embed settings');
    } finally {
      saving = false;
    }
  }

  async function saveKey() {
    if (!keyInput.trim()) return;
    saving = true;
    try {
      keyMeta = await putTenantGeminiKey(keyInput.trim());
      keyInput = '';
      feedback.success('Tenant Gemini key encrypted and saved');
    } catch (err) {
      report(err, 'Failed to save Gemini key');
    } finally {
      saving = false;
    }
  }

  async function removeKey() {
    if (!confirm('Remove the tenant Gemini key and use the platform fallback?')) return;
    try {
      keyMeta = await deleteTenantGeminiKey();
      feedback.success('Tenant key removed');
    } catch (err) {
      report(err, 'Failed to remove Gemini key');
    }
  }

  async function savePrompt() {
    saving = true;
    try {
      prompt = await putTenantPrompt(agentId, {
        system_prompt: prompt.system_prompt,
        enabled: prompt.enabled
      });
      feedback.success('Agent prompt saved');
    } catch (err) {
      report(err, 'Failed to save prompt');
    } finally {
      saving = false;
    }
  }

  async function addTool() {
    try {
      const item = await createTenantTool({
        tool_key: 'create_support_ticket',
        display_name: 'Create support ticket',
        description: 'Create a ticket after the caller confirms human follow-up.',
        handler_key: 'create_ticket',
        input_schema: {
          type: 'object',
          required: ['subject', 'description', 'category'],
          properties: {
            subject: { type: 'string', maxLength: 160 },
            description: { type: 'string', maxLength: 2000 },
            category: { type: 'string', enum: ['general', 'billing', 'technical', 'other'] }
          }
        },
        enabled: false
      });
      tools = [...tools, item];
      feedback.success('Allowlisted tool created');
    } catch (err) {
      report(err, 'Failed to create tool');
    }
  }

  async function addSkill() {
    try {
      const item = await createTenantSkill({
        slug: 'support-follow-up',
        name: 'Support follow-up',
        prompt: 'Use the approved support follow-up checklist and ask one clarifying question at a time.',
        tool_ids: tools.filter((tool) => tool.enabled).slice(0, 1).map((tool) => tool.id),
        agent_ids: [agentId],
        enabled: true
      });
      skills = [...skills, item];
      feedback.success('Tenant skill created');
    } catch (err) {
      report(err, 'Failed to create skill');
    }
  }

  onMount(() => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/ai`)}`);
      return;
    }
    void load();
  });
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">AI configuration / การตั้งค่า AI</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">
      Configure embed access and tenant-scoped AI behavior. Secrets are never shown after save.
    </p>
  </div>
</div>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  {#if error}<div class="card" style="border-color:#8b3030;color:#ff9a9a;margin-bottom:16px">{error}</div>{/if}

  <div class="grid" style="align-items:start">
    <div class="col">
      <section class="card" style="margin-bottom:16px">
        <h2 style="margin:0 0 6px;font-size:16px">Embed access / การเข้าถึง Embed</h2>
        <p style="margin:0 0 14px;color:var(--muted);font-size:12px">When enabled, customers must complete the existing OTP session before chat or voice.</p>
        <label style="display:flex;gap:10px;align-items:center;font-size:14px">
          <input type="checkbox" bind:checked={authRequired} />
          Require customer login / บังคับลูกค้าเข้าสู่ระบบ
        </label>
        <button class="primary" style="margin-top:14px" disabled={saving} onclick={() => void saveEmbed()}>Save embed settings</button>
      </section>

      <section class="card" style="margin-bottom:16px">
        <h2 style="margin:0 0 6px;font-size:16px">Gemini provider / ผู้ให้บริการ Gemini</h2>
        <p style="margin:0 0 14px;color:var(--muted);font-size:12px">The key is encrypted at rest and is not returned to the browser.</p>
        {#if keyMeta?.configured}
          <div style="display:flex;justify-content:space-between;align-items:center;gap:10px">
            <span>Configured · ••••{keyMeta.last4}</span>
            <button class="secondary" onclick={() => void removeKey()}>Remove key</button>
          </div>
        {:else}
          <span style="color:var(--muted);font-size:13px">Using platform Gemini fallback</span>
        {/if}
        <div class="field" style="margin-top:14px">
          <label for="gemini-key">Replace key</label>
          <input id="gemini-key" type="password" bind:value={keyInput} autocomplete="new-password" placeholder="Paste tenant Gemini API key" />
        </div>
        <button class="primary" disabled={!keyInput.trim() || saving} onclick={() => void saveKey()}>Encrypt and save key</button>
      </section>

      <section class="card">
        <div style="display:flex;justify-content:space-between;align-items:center;gap:12px">
          <h2 style="margin:0;font-size:16px">Agent system prompt / System prompt</h2>
          <select bind:value={agentId} onchange={() => void loadPrompt()}>
            {#each agents as agent}<option value={agent.id}>{agent.name}</option>{/each}
          </select>
        </div>
        <p style="margin:8px 0;color:var(--muted);font-size:12px">Platform safety rules remain locked. Tenant text is bounded context.</p>
        <textarea rows="9" maxlength="8000" bind:value={prompt.system_prompt} placeholder="Use our approved support tone…"></textarea>
        <div style="display:flex;justify-content:space-between;align-items:center;margin-top:8px;font-size:12px;color:var(--muted)">
          <label><input type="checkbox" bind:checked={prompt.enabled} /> Enabled</label>
          <span>{prompt.system_prompt.length}/8000</span>
        </div>
        <button class="primary" style="margin-top:12px" disabled={saving} onclick={() => void savePrompt()}>Save prompt</button>
      </section>
    </div>

    <div class="col">
      <section class="card" style="margin-bottom:16px">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <h2 style="margin:0;font-size:16px">Tools / เครื่องมือ</h2>
          <button class="secondary" onclick={() => void addTool()}>+ Add tool</button>
        </div>
        <p style="margin:8px 0 14px;color:var(--muted);font-size:12px">Only compiled server handlers can execute; arbitrary code and URLs are rejected.</p>
        {#if tools.length === 0}
          <p style="color:var(--muted);font-size:13px">No tenant tools configured.</p>
        {:else}
          {#each tools as tool}
            <div style="display:flex;justify-content:space-between;gap:12px;padding:10px 0;border-top:1px solid var(--line)">
              <div><strong>{tool.display_name}</strong><small style="display:block;color:var(--muted)">{tool.handler_key} · {tool.tool_key}</small></div>
              <span style="color:{tool.enabled ? '#36d399' : 'var(--muted)'}">{tool.enabled ? 'Enabled' : 'Disabled'}</span>
            </div>
          {/each}
        {/if}
      </section>

      <section class="card">
        <div style="display:flex;justify-content:space-between;align-items:center">
          <h2 style="margin:0;font-size:16px">Skills / ทักษะ</h2>
          <button class="secondary" onclick={() => void addSkill()}>+ Add skill</button>
        </div>
        <p style="margin:8px 0 14px;color:var(--muted);font-size:12px">Skills combine bounded prompt guidance with tenant-owned tools and agent assignments.</p>
        {#if skills.length === 0}
          <p style="color:var(--muted);font-size:13px">No tenant skills configured.</p>
        {:else}
          {#each skills as skill}
            <div style="padding:10px 0;border-top:1px solid var(--line)">
              <strong>{skill.name}</strong>
              <small style="display:block;color:var(--muted)">{skill.slug} · agents: {(skill.agent_ids || []).join(', ') || 'none'} · tools: {(skill.tool_ids || []).length}</small>
            </div>
          {/each}
        {/if}
      </section>
    </div>
  </div>
{/if}
