export type InfraStatus = {
  postgres: string;
  redis: string;
  minio: string;
  clickhouse?: string;
  nats?: string;
  livekit?: string;
};

/** Customer-facing overall health — no service names or ops jargon. */
export type SystemLiveState = 'ok' | 'issues' | 'offline' | 'checking';

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

function isServiceOk(value: string | undefined): boolean {
  if (!value) return true; // optional component omitted
  const v = value.trim().toLowerCase();
  return v === 'ok' || v === 'up' || v === 'healthy' || v === 'ready';
}

export function systemLiveState(status: InfraStatus | null): SystemLiveState {
  if (!status) return 'offline';
  const core = [status.postgres, status.redis, status.minio];
  const optional = [status.clickhouse, status.nats, status.livekit].filter(
    (v): v is string => typeof v === 'string' && v.trim() !== ''
  );
  const coreOk = core.every(isServiceOk);
  const optionalOk = optional.every(isServiceOk);
  if (coreOk && optionalOk) return 'ok';
  if (coreOk) return 'issues';
  return 'offline';
}

/** Friendly label for callers — e.g. "Live · OK", never Postgres/Redis/… */
export function formatSystemLive(status: InfraStatus | null): string {
  switch (systemLiveState(status)) {
    case 'ok':
      return 'Live · OK';
    case 'issues':
      return 'Live · Limited';
    case 'offline':
      return 'Unavailable';
    default:
      return 'Checking…';
  }
}

/** @deprecated Use formatSystemLive for customer UI. */
export function formatInfra(status: InfraStatus | null): string {
  return formatSystemLive(status);
}