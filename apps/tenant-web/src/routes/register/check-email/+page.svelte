<script lang="ts">
  import { base } from '$app/paths';
  import { page } from '$app/stores';

  const email = $derived($page.url.searchParams.get('email') ?? 'your inbox');
</script>

<div class="login-wrap">
  <div class="card login-card" style="text-align:center">
    <h1 style="margin:0 0 8px;font-size:22px">Check your email</h1>
    <p style="color:var(--muted);font-size:14px;margin:0 0 16px">
      We sent a verification link to <strong>{email}</strong>. After you verify, you can sign in and complete KYC in the tenant backoffice.
    </p>
    <a class="link" href="{base}/login">Back to sign in</a>
  </div>
</div>