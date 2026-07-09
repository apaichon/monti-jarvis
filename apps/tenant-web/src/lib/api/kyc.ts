import { apiFetch } from '$lib/api/http';
import { getAccessToken } from '$lib/auth/session';

export type KYCDocument = {
  object_key: string;
  url: string;
};

export type KYCProfile = {
  tenant_id: string;
  contact_name: string;
  contact_phone: string;
  contact_address: string;
  photo_url: string;
  photo_object_key: string;
  documents: KYCDocument[];
  status: 'draft' | 'submitted';
  submitted_at?: string;
  updated_at: string;
};

export function getKYCProfile() {
  return apiFetch<KYCProfile>('/api/tenant/kyc');
}

export function updateKYCProfile(input: {
  contact_name: string;
  contact_phone: string;
  contact_address: string;
}) {
  return apiFetch<KYCProfile>('/api/tenant/kyc', {
    method: 'PUT',
    body: JSON.stringify(input)
  });
}

export async function uploadKYCPhoto(file: File) {
  const token = getAccessToken();
  const body = new FormData();
  body.append('photo', file);
  const res = await fetch('/api/tenant/kyc/photo', {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    body
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const data = await res.json();
      if (data?.error) message = data.error;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
  return (await res.json()) as { photo_url: string };
}

export async function uploadKYCDocument(file: File) {
  const token = getAccessToken();
  const body = new FormData();
  body.append('document', file);
  const res = await fetch('/api/tenant/kyc/documents', {
    method: 'POST',
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    body
  });
  if (!res.ok) {
    let message = res.statusText;
    try {
      const data = await res.json();
      if (data?.error) message = data.error;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
  return (await res.json()) as { document_url: string };
}

export function submitKYC() {
  return apiFetch<KYCProfile>('/api/tenant/kyc/submit', { method: 'POST' });
}