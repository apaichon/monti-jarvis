-- Sprint 53: conversation email OTP may auto-create customer (default off).
\set ON_ERROR_STOP on

ALTER TABLE :POSTGRES_SCHEMA.tenant_customer_auth_settings
  ADD COLUMN IF NOT EXISTS auto_register_on_conversation_otp boolean NOT NULL DEFAULT false;
