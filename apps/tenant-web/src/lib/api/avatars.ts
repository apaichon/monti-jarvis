import { apiFetch, ApiError } from './http';
import { getAccessToken } from '$lib/auth/session';

/** Portrait upload guidance shown in the tenant avatar UI. */
export const AVATAR_IMAGE_HINT = {
  accept: 'image/jpeg,image/png,image/webp,image/gif',
  acceptLabel: 'JPEG, PNG, WebP, or GIF',
  maxBytes: 4 * 1024 * 1024,
  maxLabel: '4 MB',
  recommendSize: '512×512 px',
  recommendNote: 'Square portrait recommended (512×512). Minimum 256×256 for clear display.',
  aspect: '1:1 square'
} as const;

export type TenantAvatar = {
  id: string;
  slug: string;
  name: string;
  role: string;
  trait: string;
  color: string;
  image_url: string;
  greeting: string;
  status: string;
  owner_tenant_id?: string;
  assignment_status: 'active' | 'disabled' | string;
  /** Gemini AI Studio speaker setting name (e.g. Aoede, Puck). */
  voice?: string;
  flags?: Record<string, unknown>;
};

export type GeminiSpeakerVoice = {
  name: string;
  style: string;
  label: string;
  voice_provider_id: string;
  voice_id: string;
};

export type AvatarCap = {
  active: number;
  limit: number;
  remaining: number;
};

export type TenantAvatarsResponse = {
  avatars: TenantAvatar[];
  cap: AvatarCap;
};

export type CreateTenantAvatarBody = {
  name: string;
  role?: string;
  trait?: string;
  color?: string;
  image_url?: string;
  greeting?: string;
  /** Gemini AI Studio speaker name from generate-speech. */
  voice?: string;
};

export function listGeminiSpeakerVoices() {
  return apiFetch<{
    source: string;
    docs: string;
    provider: string;
    voices: GeminiSpeakerVoice[];
  }>('/api/tenant/avatar-voices');
}

export function listTenantAvatars() {
  return apiFetch<TenantAvatarsResponse>('/api/tenant/avatars');
}

export function createTenantAvatar(body: CreateTenantAvatarBody) {
  return apiFetch<TenantAvatar>('/api/tenant/avatars', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function updateTenantAvatar(id: string, body: Partial<CreateTenantAvatarBody>) {
  return apiFetch<TenantAvatar>(`/api/tenant/avatars/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function activateTenantAvatar(id: string) {
  return apiFetch<TenantAvatar>(`/api/tenant/avatars/${encodeURIComponent(id)}/activate`, {
    method: 'POST'
  });
}

export function deactivateTenantAvatar(id: string) {
  return apiFetch<TenantAvatar>(`/api/tenant/avatars/${encodeURIComponent(id)}/deactivate`, {
    method: 'POST'
  });
}

export function archiveTenantAvatar(id: string) {
  return apiFetch<{ status: string }>(`/api/tenant/avatars/${encodeURIComponent(id)}`, {
    method: 'DELETE'
  });
}

export type TenantAvatarImageUploadResponse = {
  image_url: string;
  status: 'uploaded' | 'uploaded_and_saved';
  tenant_avatar?: TenantAvatar;
};

export async function uploadTenantAvatarImage(
  avatarId: string,
  file: File
): Promise<TenantAvatarImageUploadResponse> {
  if (file.size > AVATAR_IMAGE_HINT.maxBytes) {
    throw new ApiError(400, `Image exceeds ${AVATAR_IMAGE_HINT.maxLabel} limit`);
  }
  const form = new FormData();
  form.append('file', file);
  const headers = new Headers();
  const token = getAccessToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  const res = await fetch(`/api/tenant/avatars/${encodeURIComponent(avatarId)}/image`, {
    method: 'POST',
    headers,
    credentials: 'include',
    body: form
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      // ignore
    }
    throw new ApiError(res.status, message);
  }
  return (await res.json()) as TenantAvatarImageUploadResponse;
}
