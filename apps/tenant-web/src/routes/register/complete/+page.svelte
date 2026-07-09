<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { completeOAuthRegistration } from '$lib/api/register';
  import { ApiError } from '$lib/api/http';
  import { saveSession } from '$lib/auth/session';
  import { suggestSlug } from '$lib/utils/slug';
  import { feedback } from '$lib/feedback.svelte';

  let sessionId = $state('');
  let companyName = $state('');
  let slug = $state('');
  let loading = $state(false);

  onMount(() => {
    sessionId = $page.url.searchParams.get('session') ?? '';
    if (!sessionId) feedback.error('OAuth session missing — start registration again');
  });

  function onCompanyInput() {
    slug = suggestSlug(companyName);
  }

  async function submit(e: Event) {
    e.preventDefault();
    if (!sessionId || !companyName.trim() || !slug.trim()) {
      feedback.error('Company name and workspace URL are required');
      return;
    }
    loading = true;
    try {
      const res = await completeOAuthRegistration({
        session_id: sessionId,
        company_name: companyName.trim(),
        slug: slug.trim()
      });
      if (res.access_token && res.refresh_token && res.user) {
        saveSession(
          {
            access_token: res.access_token,
            refresh_token: res.refresh_token,
            expires_in: res.expires_in ?? 0,
            token_type: res.token_type ?? 'Bearer',
            user: res.user
          },
          res.tenant_id,
          res.registration_id
        );
        goto(`${base}/register/success`);
      }
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Could not complete registration');
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-wrap">
  <div class="card login-card">
    <h1 style="margin:0 0 8px;font-size:20px">Finish your workspace</h1>
    <p style="color:var(--muted);font-size:14px;margin:0 0 16px">
      Your Google/GitHub account is linked. Choose your company workspace details.
    </p>
    <form onsubmit={submit}>
      <div class="field">
        <label for="company">Company name</label>
        <input id="company" bind:value={companyName} oninput={onCompanyInput} disabled={loading} />
      </div>
      <div class="field">
        <label for="slug">Workspace URL</label>
        <div class="slug-row">
          <span class="slug-prefix">monti.app/</span>
          <input id="slug" bind:value={slug} disabled={loading} />
        </div>
      </div>
      <button class="btn" type="submit" disabled={loading} style="width:100%">Create workspace</button>
    </form>
  </div>
</div>