<script lang="ts">
  import { uploadAvatarImage } from '$lib/api/avatars';
  import { ApiError } from '$lib/api/http';

  let {
    avatarId,
    imageUrl = $bindable(''),
    onError = (_message: string) => {}
  }: {
    avatarId: string;
    imageUrl?: string;
    onError?: (message: string) => void;
  } = $props();

  let uploading = $state(false);
  let uploadStatus = $state('');
  let previewSrc = $derived(imageUrl || '');

  async function onFileChange(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file || !avatarId) return;

    uploading = true;
    uploadStatus = '';
    try {
      const res = await uploadAvatarImage(avatarId, file);
      imageUrl = res.image_url;
      uploadStatus =
        res.status === 'uploaded_and_saved'
          ? 'Image uploaded to MinIO and saved.'
          : 'Image uploaded to MinIO.';
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : 'Image upload failed';
      onError(msg);
      uploadStatus = '';
    } finally {
      uploading = false;
    }
  }
</script>

<div class="field">
  <label for="avatar-image-file">Portrait image</label>
  <div class="avatar-image-row">
    {#if previewSrc}
      <div class="avatar-preview" style="--preview-size:96px">
        <img src={previewSrc} alt="Avatar preview" loading="lazy" />
      </div>
    {:else}
      <div class="avatar-preview placeholder" aria-hidden="true">No image</div>
    {/if}
    <div class="avatar-image-controls">
      <input
        id="avatar-image-file"
        type="file"
        accept="image/jpeg,image/png,image/webp,image/gif"
        disabled={!avatarId || uploading}
        onchange={onFileChange}
      />
      <input
        id="image_url"
        bind:value={imageUrl}
        placeholder="/api/assets/avatars/{avatarId || 'slug'}/portrait.jpg"
      />
      <p class="hint">Upload stores to MinIO and sets the public asset URL. JPEG, PNG, WebP, or GIF up to 4MB.</p>
      {#if uploading}
        <p class="hint">Uploading…</p>
      {:else if uploadStatus}
        <p class="hint success-text">{uploadStatus}</p>
      {/if}
    </div>
  </div>
</div>

<style>
  .avatar-image-row {
    display: flex;
    gap: 16px;
    align-items: flex-start;
    flex-wrap: wrap;
  }

  .avatar-preview {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    overflow: hidden;
    border: 2px solid rgb(0 183 255 / 35%);
    flex-shrink: 0;
    background: rgb(3 11 23 / 88%);
  }

  .avatar-preview img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .avatar-preview.placeholder {
    display: grid;
    place-items: center;
    font-size: 11px;
    color: var(--muted);
    text-align: center;
    padding: 8px;
  }

  .avatar-image-controls {
    flex: 1;
    min-width: 220px;
    display: grid;
    gap: 8px;
  }

  .avatar-image-controls input[type='file'] {
    font-size: 13px;
    color: var(--muted);
  }

  .avatar-image-controls input:not([type='file']) {
    border: 1px solid rgb(70 132 190 / 24%);
    border-radius: 10px;
    background: rgb(3 11 23 / 88%);
    color: var(--ink);
    padding: 10px 12px;
  }

  .hint {
    margin: 0;
    font-size: 12px;
    color: var(--muted);
  }

  .success-text {
    color: var(--success);
  }
</style>