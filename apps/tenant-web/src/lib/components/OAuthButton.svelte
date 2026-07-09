<script lang="ts">
  import OAuthProviderIcon from '$lib/components/OAuthProviderIcon.svelte';

  let {
    provider,
    label,
    disabled = false,
    onclick
  }: {
    provider: 'google' | 'github';
    label: string;
    disabled?: boolean;
    onclick: () => void;
  } = $props();
</script>

<button class="oauth-btn oauth-btn-{provider}" type="button" {disabled} {onclick}>
  <OAuthProviderIcon {provider} />
  <span>{label}</span>
</button>

<style>
  .oauth-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    width: 100%;
    border-radius: 10px;
    padding: 10px 16px;
    border: 1px solid rgb(70 132 190 / 28%);
    background: rgb(255 255 255 / 6%);
    color: var(--ink);
    font-size: 14px;
    font-weight: 500;
    transition: background 0.15s ease, border-color 0.15s ease;
  }

  .oauth-btn:hover:not(:disabled) {
    background: rgb(255 255 255 / 10%);
    border-color: rgb(0 183 255 / 35%);
  }

  .oauth-btn:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }

  .oauth-btn-google {
    background: rgb(255 255 255 / 92%);
    color: #1f1f1f;
    border-color: rgb(255 255 255 / 20%);
  }

  .oauth-btn-google:hover:not(:disabled) {
    background: #fff;
    border-color: rgb(255 255 255 / 40%);
  }

  .oauth-btn-github {
    background: #24292f;
    border-color: rgb(255 255 255 / 12%);
    color: #f0f6fc;
  }

  .oauth-btn-github:hover:not(:disabled) {
    background: #2f363d;
    border-color: rgb(255 255 255 / 20%);
  }
</style>