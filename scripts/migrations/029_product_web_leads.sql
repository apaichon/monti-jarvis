-- Sprint 48: product web marketing leads and funnel events.
\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.marketing_leads (
  id text PRIMARY KEY,
  kind text NOT NULL CHECK (kind IN ('contact', 'book_demo', 'newsletter')),
  status text NOT NULL DEFAULT 'new' CHECK (status IN (
    'new', 'contacted', 'demo_scheduled', 'demo_completed', 'qualified',
    'registered', 'kyc_pending', 'package_selected', 'paid', 'converted',
    'lost', 'unsubscribed'
  )),
  email text NOT NULL,
  full_name text NOT NULL DEFAULT '',
  company_name text NOT NULL DEFAULT '',
  phone text NOT NULL DEFAULT '',
  use_case text NOT NULL DEFAULT '',
  preferred_channel text NOT NULL DEFAULT '' CHECK (preferred_channel IN ('', 'email', 'phone', 'line', 'other')),
  language text NOT NULL DEFAULT 'en' CHECK (language IN ('en', 'th')),
  consent_marketing boolean NOT NULL DEFAULT false,
  consent_contact boolean NOT NULL DEFAULT false,
  consent_at timestamptz,
  utm_source text NOT NULL DEFAULT '',
  utm_medium text NOT NULL DEFAULT '',
  utm_campaign text NOT NULL DEFAULT '',
  utm_content text NOT NULL DEFAULT '',
  utm_term text NOT NULL DEFAULT '',
  referral_code text NOT NULL DEFAULT '',
  landing_path text NOT NULL DEFAULT '',
  package_interest_id text NOT NULL DEFAULT '',
  dedupe_key text NOT NULL DEFAULT '',
  assigned_to text NOT NULL DEFAULT '',
  converted_tenant_id text REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (kind, email)
);

CREATE INDEX IF NOT EXISTS marketing_leads_status_created_idx
  ON :POSTGRES_SCHEMA.marketing_leads (status, created_at DESC);

CREATE INDEX IF NOT EXISTS marketing_leads_kind_created_idx
  ON :POSTGRES_SCHEMA.marketing_leads (kind, created_at DESC);

CREATE INDEX IF NOT EXISTS marketing_leads_utm_source_idx
  ON :POSTGRES_SCHEMA.marketing_leads (utm_source) WHERE utm_source <> '';

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.marketing_lead_notes (
  id text PRIMARY KEY,
  lead_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.marketing_leads(id) ON DELETE CASCADE,
  body text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS marketing_lead_notes_lead_idx
  ON :POSTGRES_SCHEMA.marketing_lead_notes (lead_id, created_at ASC);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.marketing_lead_events (
  id text PRIMARY KEY,
  lead_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.marketing_leads(id) ON DELETE CASCADE,
  from_status text,
  to_status text NOT NULL,
  actor text NOT NULL DEFAULT 'system',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS marketing_lead_events_lead_idx
  ON :POSTGRES_SCHEMA.marketing_lead_events (lead_id, created_at ASC);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.funnel_events (
  id text PRIMARY KEY,
  event_name text NOT NULL,
  page_path text NOT NULL DEFAULT '',
  cta_id text NOT NULL DEFAULT '',
  utm_source text NOT NULL DEFAULT '',
  utm_medium text NOT NULL DEFAULT '',
  utm_campaign text NOT NULL DEFAULT '',
  utm_content text NOT NULL DEFAULT '',
  utm_term text NOT NULL DEFAULT '',
  referral_code text NOT NULL DEFAULT '',
  session_key text NOT NULL DEFAULT '',
  client_ip_hash text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS funnel_events_name_created_idx
  ON :POSTGRES_SCHEMA.funnel_events (event_name, created_at DESC);

CREATE INDEX IF NOT EXISTS funnel_events_created_idx
  ON :POSTGRES_SCHEMA.funnel_events (created_at DESC);
