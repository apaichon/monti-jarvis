import { apiFetch } from './http';

export type AvatarVoice = {
  id?: string;
  voice_provider_id: string;
  voice_id: string;
  voice: string;
  priority: number;
  status: string;
};

export type AvatarFlags = {
  popular?: boolean;
  robot?: boolean;
  skin?: string;
  hair?: string;
};

export type Avatar = {
  id: string;
  slug: string;
  name: string;
  role: string;
  trait: string;
  color: string;
  image_url: string;
  greeting: string;
  status: string;
  flags: AvatarFlags;
  voices: AvatarVoice[];
};

export type TenantAvatarAssignment = {
  avatar_id: string;
  status: string;
  avatar: Avatar;
};

export type TenantAssignmentsResponse = {
  tenant_id: string;
  assignments: TenantAvatarAssignment[];
  cap: {
    max_ai_employees: number;
    active_count: number;
  };
};

export type AvatarInput = {
  slug: string;
  name: string;
  role: string;
  trait?: string;
  color?: string;
  image_url?: string;
  greeting: string;
  status?: string;
  flags?: AvatarFlags;
  voices: AvatarVoice[];
};

export function listAvatars(status = '') {
  const q = status ? `?status=${encodeURIComponent(status)}` : '';
  return apiFetch<{ avatars: Avatar[] }>(`/api/platform/avatars${q}`);
}

export function getAvatar(id: string) {
  return apiFetch<Avatar>(`/api/platform/avatars/${id}`);
}

export function createAvatar(body: AvatarInput) {
  return apiFetch<Avatar>('/api/platform/avatars', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function updateAvatar(id: string, body: Partial<AvatarInput>) {
  return apiFetch<Avatar>(`/api/platform/avatars/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function archiveAvatar(id: string) {
  return apiFetch<{ status: string }>(`/api/platform/avatars/${id}`, { method: 'DELETE' });
}

export function listTenantAvatars(tenantId: string) {
  return apiFetch<TenantAssignmentsResponse>(`/api/platform/tenants/${tenantId}/avatars`);
}

export function assignTenantAvatar(tenantId: string, avatarId: string) {
  return apiFetch<TenantAvatarAssignment>(`/api/platform/tenants/${tenantId}/avatars`, {
    method: 'POST',
    body: JSON.stringify({ avatar_id: avatarId })
  });
}

export function revokeTenantAvatar(tenantId: string, avatarId: string) {
  return apiFetch<{ status: string }>(
    `/api/platform/tenants/${tenantId}/avatars/${avatarId}`,
    { method: 'DELETE' }
  );
}

export function defaultVoiceRow(priority = 1): AvatarVoice {
  return {
    voice_provider_id: 'voice-gemini-live',
    voice_id: 'gemini-2.5-flash-native-audio-latest',
    voice: 'Aoede',
    priority,
    status: 'active'
  };
}