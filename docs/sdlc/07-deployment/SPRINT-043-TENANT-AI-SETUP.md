---
id: DEPLOY-SPRINT-043-TENANT-AI
status: active
updated: 2026-07-22
environment: local-dev-and-production
sprint: SPRINT-043
---

# Sprint 43 Tenant AI Setup

This guide explains how to configure the **AI configuration / การตั้งค่า AI**
page in the tenant portal and the server parameters required by Sprint 43.

It covers:

- Embed customer authentication
- Tenant Gemini API keys
- Agent system prompts
- Tenant tools and skills
- Environment groups and database setup

## 1. Configure the server first

The tenant Gemini key is encrypted before it is stored. Set the deployment
encryption key before saving a tenant Gemini key.

For local development:

```bash
cd /path/to/monti-jarvis
cp infra/.env.dev.example infra/.env.dev
openssl rand -base64 32
```

Copy the generated value into `infra/.env.dev`:

```dotenv
TENANT_SECRET_ENCRYPTION_KEY=<base64-encoded-32-byte-value>
TENANT_SECRET_KEY_VERSION=v1
```

`TENANT_SECRET_ENCRYPTION_KEY` is a deployment secret, not a tenant Gemini
key. Keep the same value across restarts. If it is changed, previously saved
tenant keys cannot be decrypted; tenants must remove and save their provider
key again unless a key-rotation process has been prepared.

Set the platform Gemini fallback in the same file:

```dotenv
GEMINI_API_KEY=<platform-gemini-api-key>
GEMINI_MODEL=gemini-flash-latest
GEMINI_LIVE_MODEL=gemini-2.5-flash-native-audio-latest
GEMINI_EMBED_MODEL=gemini-embedding-001
```

The platform key is used when a tenant has not configured its own key. It is
also the normal fallback for platform-level chat, voice, and embeddings.

### Environment parameter reference

| Parameter | Required | How to set it | Purpose |
| --- | --- | --- | --- |
| `APP_ENV` | No | `dev` or `prod` | Selects `infra/.env.<APP_ENV>`. |
| `CONFIG_GROUPS` | No | Empty, or `ai,ops,email,features` | Loads optional grouped environment files. |
| `TENANT_SECRET_ENCRYPTION_KEY` | For tenant keys | `openssl rand -base64 32` | Base64-encoded 32-byte AES-256-GCM deployment key. |
| `TENANT_SECRET_KEY_VERSION` | No | `v1` | Metadata identifying the encryption-key version. |
| `GEMINI_API_KEY` | Recommended | Gemini API key | Platform provider fallback. |
| `GEMINI_MODEL` | No | Model name | Platform text model; default `gemini-flash-latest`. |
| `GEMINI_LIVE_MODEL` | No | Model name | Platform voice model. |
| `GEMINI_EMBED_MODEL` | No | Model name | Knowledge-base embedding model. |
| `POSTGRES_URL` | Yes with DB | Postgres connection URL | Stores tenant AI configuration in `callcenter`. |
| `POSTGRES_SCHEMA` | No | `callcenter` | Must remain `callcenter`; do not use Jarvis Chat’s schema. |
| `AUTH_DISABLED` | Local only | `true` or `false` | Keep `true` for the no-auth local demo; use authenticated tenant admin access when false. |

The simplest setup is to leave `CONFIG_GROUPS=` empty and put the values in
`infra/.env.dev`. If using `CONFIG_GROUPS=ai`, put the `GEMINI_*` values in
`infra/.env.dev.ai`; do not rely on a blank `GEMINI_*` value already loaded
from the core file, because process environment values take precedence.

Apply the schema and restart the server:

```bash
make infra-up       # first time only, if local infrastructure is stopped
make db-migrate
make restart
```

The migration is safe to rerun. It upgrades existing Sprint 43 tables with
missing columns as well as creating new installations.

## 2. Embed access

Page section: **Embed access / การเข้าถึง Embed**

| UI parameter | API field | Values | Effect |
| --- | --- | --- | --- |
| Require customer login | `auth_required` | Off/`false`, On/`true` | When on, customers must complete the existing OTP/session login before chat or voice. |

Steps:

1. Turn **Require customer login** on or off.
2. Click **Save embed settings**.
3. Test the embed in both the public and authenticated flows as appropriate.

The default is off to preserve the existing public embed behavior.

## 3. Gemini provider

Page section: **Gemini provider / ผู้ให้บริการ Gemini**

| UI parameter | API field | Rule |
| --- | --- | --- |
| Replace key | `api_key` | Paste the tenant’s Gemini API key. The server trims it and validates its length. |
| Configured status | `configured` | Only metadata is returned after saving. |
| Last four characters | `last4` | Display-only confirmation; the full key is never returned. |
| Remove key | `DELETE /api/tenant/ai/gemini-key` | Removes the tenant key and restores the platform fallback. |

Steps:

1. Confirm `TENANT_SECRET_ENCRYPTION_KEY` is set on the server.
2. Paste the tenant Gemini API key into **Replace key**.
3. Click **Encrypt and save key**.
4. Confirm the page shows `Configured · ••••1234` (the final four characters will differ).

The key is encrypted at rest with the deployment secret. It is not stored in
browser local storage, Redis, or the API response. A tenant key is preferred
over the platform key; if the tenant key is configured but unusable, the
request fails closed instead of silently switching providers.

### Error shown in the screenshot

`tenant secret encryption is not configured` means the server process did not
load `TENANT_SECRET_ENCRYPTION_KEY`. It does not mean the Gemini API key is
wrong.

Fix it by setting the variable in the environment file actually used by the
server, then restarting:

```bash
make restart
```

If the error remains, check the startup log for configuration errors:

```bash
make logs
```

Do not print the secret or paste it into logs, tickets, screenshots, or chat.

## 4. Agent system prompt

Page section: **Agent system prompt / System prompt**

The current tenant portal lists these agents:

| Agent selector | Agent ID |
| --- | --- |
| Ava | `ava` |
| Max | `max` |
| Luna | `luna` |
| Neo | `neo` |

| UI parameter | API field | Rule |
| --- | --- | --- |
| Agent selector | URL `{agent_id}` | Select one assigned AI employee. |
| Prompt text | `system_prompt` | Maximum 8,000 Unicode characters. |
| Enabled | `enabled` | Include or exclude this tenant prompt for the selected agent. |

Steps:

1. Select an agent.
2. Enter the tenant-specific behavior, tone, and business instructions.
3. Leave **Enabled** on unless the prompt is being tested or temporarily disabled.
4. Click **Save prompt**.

Platform safety rules remain locked. The tenant prompt is additional context;
it cannot replace the platform safety policy or authorize arbitrary actions.

Runtime prompt precedence is:

```text
platform safety policy
  → built-in workforce prompt
  → tenant agent prompt
  → assigned skill prompts
  → locale and RAG context
  → closing safety reminder
```

Good prompt example:

```text
Answer as the Thailand Post customer-support specialist. Prefer Thai when the
caller writes Thai. Ask for the tracking number before checking delivery
status. Never invent a tracking result; explain how the customer can contact a
human agent when the information is unavailable.
```

## 5. Tools

Page section: **Tools / เครื่องมือ**

Tools are declarative definitions. Tenants cannot upload code, shell commands,
SQL, arbitrary URLs, or webhooks. The server executes only compiled handlers
that are explicitly allowlisted.

The current UI’s **+ Add tool** action creates the supported ticket tool with
these values:

| Parameter | Example | Rule |
| --- | --- | --- |
| `tool_key` | `create_support_ticket` | Lowercase; starts with a letter; max 64 characters; underscores allowed. |
| `display_name` | `Create support ticket` | Human-readable label. |
| `description` | `Create a ticket after the caller confirms human follow-up.` | Sent to Gemini; keep it precise. |
| `handler_key` | `create_ticket` | Must match a registered server handler. |
| `input_schema` | JSON Schema object | Defines and validates function arguments. |
| `enabled` | `false` | Disabled by default; enable only after testing. |

Recommended workflow:

1. Add the tool.
2. Verify its handler, schema, and tenant ownership.
3. Keep it disabled while testing.
4. Assign it to a skill only when the arguments and confirmation behavior are correct.
5. Enable it and test with a customer-safe request.

Tool calls are tenant-scoped, argument-validated, and limited per turn. Never
put API keys, passwords, or unrestricted execution instructions in a tool
schema.

## 6. Skills

Page section: **Skills / ทักษะ**

A skill combines bounded prompt guidance with tenant-owned tools and agent
assignments.

| Parameter | Example | Rule |
| --- | --- | --- |
| `slug` | `support-follow-up` | Lowercase slug; starts with a letter; max 64 characters; hyphens allowed. |
| `name` | `Support follow-up` | Human-readable name. |
| `prompt` | Approved support checklist | Maximum 8,000 Unicode characters. |
| `tool_ids` | A tenant tool ID list | Every tool must belong to the same tenant. |
| `agent_ids` | `ava` | Every agent must be assigned/known for the tenant. |
| `enabled` | `true` | Only enabled skills are included at runtime. |

Example skill prompt:

```text
Use the approved support follow-up checklist. Ask one clarifying question at
a time. Before creating a ticket, summarize the issue and ask the caller to
confirm human follow-up.
```

## 7. Verification checklist

After configuration:

- [ ] `TENANT_SECRET_ENCRYPTION_KEY` is set and the server has been restarted.
- [ ] `make db-migrate` completed successfully.
- [ ] Tenant Gemini key shows only masked metadata after saving.
- [ ] Platform Gemini fallback is configured if tenants without their own key must use AI.
- [ ] Embed auth behavior matches the tenant’s intended customer flow.
- [ ] The selected agent prompt is enabled and under 8,000 characters.
- [ ] Tools are disabled until their schemas and handlers are tested.
- [ ] Skills reference only same-tenant tools and agents.
- [ ] No secret appears in browser storage, logs, audit records, or screenshots.

## 8. Troubleshooting

| Symptom | Cause | Fix |
| --- | --- | --- |
| `tenant secret encryption is not configured` | Missing `TENANT_SECRET_ENCRYPTION_KEY` | Set a base64 32-byte key and run `make restart`. |
| `invalid provider key` | Empty, malformed, or too-short tenant key | Paste the actual Gemini API key and save again. |
| `column ... does not exist` | Existing Sprint 43 table was created before a column was added | Run `make db-migrate`, then restart. |
| Tenant key shows configured but requests fail | Deployment secret changed or key cannot decrypt | Restore the original deployment secret, or remove and save the tenant key again. |
| Requests use platform key | Tenant key was removed or never configured | Save a tenant key after configuring encryption, or keep using the platform fallback intentionally. |
| Agent prompt not applied | Prompt disabled, agent not assigned, or wrong agent selected | Select the assigned agent, enable the prompt, and save it again. |
