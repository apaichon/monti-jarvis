import {
  getAttribution,
  getOrCreateSessionKey,
  type AttributionMap
} from '$lib/attribution';

export class ApiError extends Error {
  status: number;
  code?: string;

  constructor(status: number, message: string, code?: string) {
    super(message);
    this.status = status;
    this.code = code;
  }
}

async function parseError(res: Response): Promise<ApiError> {
  let message = res.statusText;
  let code: string | undefined;
  try {
    const body = await res.json();
    if (body?.error) message = typeof body.error === 'string' ? body.error : message;
    if (body?.code) code = String(body.code);
    if (body?.message && !body?.error) message = String(body.message);
  } catch {
    // ignore
  }
  return new ApiError(res.status, message, code);
}

export type LeadKind = 'contact' | 'book_demo' | 'newsletter';

export type LeadRequest = {
  kind: LeadKind;
  email: string;
  full_name?: string;
  company_name?: string;
  phone?: string;
  use_case?: string;
  preferred_channel?: 'email' | 'phone' | 'line' | 'other' | '';
  language?: 'en' | 'th' | 'ja';
  consent_contact?: boolean;
  consent_marketing?: boolean;
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
  referral_code?: string;
  landing_path?: string;
  package_interest_id?: string;
  /** Honeypot — must be empty. */
  website?: string;
};

export type LeadResponse = {
  lead_id: string;
  status: string;
  deduped: boolean;
};

export type PublicPackage = {
  id: string;
  slug?: string;
  name: string;
  description?: string;
  price_amount: number;
  price_currency: string;
  billing_period: string;
  /** self_serve = payment gateway; quote = dedicated capacity request */
  purchase_mode?: 'self_serve' | 'quote' | string;
  deployment?: 'shared_cloud' | 'dedicated_vm' | string;
  highlights?: string[];
  rules_summary?: Record<string, number | string>;
};

export type PublicPackagesResponse = {
  packages: PublicPackage[];
};

export type FunnelEventName =
  | 'page_view'
  | 'cta_click'
  | 'demo_start'
  | 'lead_submit'
  | 'register_start';

export type FunnelEventRequest = {
  event_name: FunnelEventName;
  page_path: string;
  cta_id?: string;
  session_key?: string;
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
  referral_code?: string;
};

function attributionFields(attrs: AttributionMap = getAttribution()) {
  return {
    utm_source: attrs.utm_source,
    utm_medium: attrs.utm_medium,
    utm_campaign: attrs.utm_campaign,
    utm_content: attrs.utm_content,
    utm_term: attrs.utm_term,
    referral_code: attrs.ref
  };
}

export async function postLead(body: LeadRequest): Promise<LeadResponse> {
  const attrs = getAttribution();
  const payload: LeadRequest = {
    ...body,
    website: body.website ?? '',
    utm_source: body.utm_source ?? attrs.utm_source,
    utm_medium: body.utm_medium ?? attrs.utm_medium,
    utm_campaign: body.utm_campaign ?? attrs.utm_campaign,
    utm_content: body.utm_content ?? attrs.utm_content,
    utm_term: body.utm_term ?? attrs.utm_term,
    referral_code: body.referral_code ?? attrs.ref,
    package_interest_id: body.package_interest_id ?? attrs.package_id,
    language:
      body.language ??
      (attrs.lang === 'th' || attrs.lang === 'ja' || attrs.lang === 'en' ? attrs.lang : 'en')
  };

  const res = await fetch('/api/public/leads', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  if (!res.ok) throw await parseError(res);
  return (await res.json()) as LeadResponse;
}

export async function getPublicPackages(): Promise<PublicPackagesResponse> {
  const res = await fetch('/api/public/packages');
  if (!res.ok) throw await parseError(res);
  return (await res.json()) as PublicPackagesResponse;
}

export async function postFunnelEvent(
  partial: Omit<FunnelEventRequest, 'session_key'> & { session_key?: string }
): Promise<void> {
  const attrs = getAttribution();
  const payload: FunnelEventRequest = {
    ...partial,
    session_key: partial.session_key ?? getOrCreateSessionKey(),
    ...attributionFields(attrs),
    // prefer explicit over attribution defaults
    utm_source: partial.utm_source ?? attrs.utm_source,
    utm_medium: partial.utm_medium ?? attrs.utm_medium,
    utm_campaign: partial.utm_campaign ?? attrs.utm_campaign,
    utm_content: partial.utm_content ?? attrs.utm_content,
    utm_term: partial.utm_term ?? attrs.utm_term,
    referral_code: partial.referral_code ?? attrs.ref
  };

  const res = await fetch('/api/public/funnel/events', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });
  if (!res.ok) {
    // Callers may ignore funnel errors; still throw for optional handling
    throw await parseError(res);
  }
}
