export type InfraStatus = {
  postgres: string;
  redis: string;
  minio: string;
  nats?: string;
  livekit?: string;
};

export async function loadInfra(): Promise<InfraStatus | null> {
  try {
    const res = await fetch('/api/infra');
    const data = await res.json();
    if (!res.ok) return null;
    return data;
  } catch {
    return null;
  }
}

export function formatInfra(status: InfraStatus | null): string {
  if (!status) return 'Infra unavailable';
  return `Postgres ${status.postgres} · Redis ${status.redis} · MinIO ${status.minio}`;
}