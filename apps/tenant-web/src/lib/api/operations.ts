import { apiFetch, handleUnauthorized } from './http';
import { getAccessToken } from '$lib/auth/session';

export type ConversationRecord = {
  id: string;
  call_id: string;
  customer_id?: string;
  avatar_id?: string;
  avatar_name?: string;
  channel: 'chat' | 'voice';
  status: 'recording' | 'archived' | 'archive_failed';
  started_at: string;
  ended_at?: string;
  duration_seconds: number;
  summary?: Record<string, unknown>;
  archive_object_count?: number;
  knowledge_gap_count?: number;
};

export type ConversationTranscriptLine = {
  id: string;
  role: string;
  content: string;
  created_at: string;
};

export type ConversationArchiveObject = {
  id: string;
  conversation_record_id: string;
  object_key?: string;
  object_type: 'transcript' | 'audio' | 'metadata';
  content_type: string;
  size_bytes: number;
  checksum_sha256: string;
  protection_mode: string;
  status: 'stored' | 'failed' | 'deleted';
  error_code?: string;
  stored_at?: string;
};

export type ConversationRecordDetail = {
  record: ConversationRecord;
  transcript: ConversationTranscriptLine[];
  archive_objects: ConversationArchiveObject[];
};

export type KnowledgeGap = {
  id: string;
  conversation_record_id?: string;
  avatar_id?: string;
  customer_id?: string;
  question: string;
  answer_excerpt?: string;
  gap_reason: string;
  confidence: number;
  status: 'open' | 'snoozed' | 'resolved' | 'ignored';
  reviewer_note?: string;
  created_at?: string;
};

export async function listConversationRecords(status = '', startDate = '', endDate = '') {
  const params = new URLSearchParams();
  if (status) params.set('status', status);
  if (startDate) params.set('start_date', startDate);
  if (endDate) params.set('end_date', endDate);
  const qs = params.toString() ? `?${params.toString()}` : '';
  const data = await apiFetch<{ records: ConversationRecord[] }>(`/api/tenant/conversation-records${qs}`);
  return data.records ?? [];
}

export async function getConversationRecord(id: string) {
  return apiFetch<ConversationRecordDetail>(`/api/tenant/conversation-records/${id}`);
}

export function retryConversationArchive(id: string) {
  return apiFetch<{ status: string }>(`/api/tenant/conversation-records/${id}/archive/retry`, {
    method: 'POST'
  });
}

export async function getConversationArchiveObjectBlob(recordId: string, objectId: string) {
  const headers = new Headers();
  const token = getAccessToken();
  if (token) headers.set('Authorization', `Bearer ${token}`);
  const res = await fetch(`/api/tenant/conversation-records/${recordId}/archive-objects/${objectId}/content`, {
    headers
  });
  if (res.status === 401 && !!getAccessToken()) handleUnauthorized(true);
  if (!res.ok) throw new Error('Failed to load archive audio');
  return res.blob();
}

export async function listKnowledgeGaps(status = 'open') {
  const qs = status ? `?status=${encodeURIComponent(status)}` : '';
  const data = await apiFetch<{ gaps: KnowledgeGap[] }>(`/api/tenant/knowledge-gaps${qs}`);
  return data.gaps ?? [];
}

export async function patchKnowledgeGap(id: string, body: { status: KnowledgeGap['status']; reviewer_note?: string }) {
  const data = await apiFetch<{ gap: KnowledgeGap }>(`/api/tenant/knowledge-gaps/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(body)
  });
  return data.gap;
}

export type SatisfactionBucket = {
  id?: string;
  name?: string;
  channel?: 'chat' | 'voice';
  completed: number;
  reviewed: number;
  average_score: number;
};

export type SatisfactionStatistics = {
  range: { start_date: string; end_date: string };
  total_completed_conversations: number;
  reviewed_conversations: number;
  unrated_conversations: number;
  review_completion_rate: number;
  average_score: number;
  distribution: Record<string, number>;
  by_avatar: SatisfactionBucket[];
  by_channel: SatisfactionBucket[];
};

export async function getSatisfactionStatistics(filters: {
  startDate?: string;
  endDate?: string;
  avatarId?: string;
  channel?: '' | 'chat' | 'voice';
} = {}) {
  const params = new URLSearchParams();
  if (filters.startDate) params.set('start_date', filters.startDate);
  if (filters.endDate) params.set('end_date', filters.endDate);
  if (filters.avatarId) params.set('avatar_id', filters.avatarId);
  if (filters.channel) params.set('channel', filters.channel);
  const qs = params.toString() ? `?${params.toString()}` : '';
  return apiFetch<SatisfactionStatistics>(`/api/tenant/satisfaction/statistics${qs}`);
}
