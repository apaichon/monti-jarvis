---
id: UAT-SPRINT-044
status: backlog
hold_reason: security_review_required
updated: 2026-07-25
sprint: SPRINT-044
owner: Tester
---

# SPRINT-044 — Manual Test Checklist (ON HOLD)

> **ON HOLD:** This checklist is retained for a future security review only.
> The customer workspace and provider/CLI runtime are not present in the build.

## 0. Preconditions

- [ ] `TENANT_SECRET_ENCRYPTION_KEY` is a valid base64-encoded 32-byte value
  and remains stable across the server restart.
- [ ] `CLAUDE_BASE_URL` points to the configured Claude Messages API, or to a
  local Claude-compatible test double for deterministic smoke testing.
- [ ] Customer auth is enabled for the test tenant and a customer has an
  active OTP/session account.
- [ ] Postgres `callcenter`, Redis DB 4, and MinIO bucket `monti-jarvis` are
  available.

## 1. Initialize infrastructure

```bash
make infra-up
make infra-check
make db-migrate
make restart
```

- [ ] `/healthz` returns `ok: true`.
- [ ] `/api/infra` reports Postgres, Redis, and MinIO as available.

## 2. Customer workspace and authentication

1. Open the customer portal and sign in with OTP.
2. Click **Workspace**, or open `/workspace?tenant_id=<tenant>`.

| Step | Expected | Result |
| --- | --- | --- |
| 2.1 | Workspace without a customer token | 401 / sign-in message; no job can be created | [ ] |
| 2.2 | Signed-in workspace loads | Claude, Codex, Antigravity, and Grok states are visible | [ ] |
| 2.3 | Tenant context omitted from request | Authenticated token tenant is used | [ ] |

## 3. Credential vault and provider states

1. Save a valid Claude API key.
2. Refresh the page and inspect the connection response/network payload.
3. Revoke the connection.

| Step | Expected | Result |
| --- | --- | --- |
| 3.1 | Save key | Response shows only configured status, version, expiry, and last four characters | [ ] |
| 3.2 | Browser/network inspection | Full key, ciphertext, and nonce are absent | [ ] |
| 3.3 | Revoke | Claude returns to `not_configured`; generation is disabled | [ ] |
| 3.4 | Try Codex/Antigravity/Grok connection | Bounded `409 unsupported_provider`; no secret row is created | [ ] |
| 3.5 | Expired credential | Generation fails closed with `credential_expired` | [ ] |

## 4. End-to-end Claude generation

1. Connect Claude.
2. Select `html` and submit a short prompt.
3. Poll history until terminal.
4. Preview the artifact.

| Step | Expected | Result |
| --- | --- | --- |
| 4.1 | Create request | `202` queued response with a job id | [ ] |
| 4.2 | Progress polling | State moves `queued` → `running` → `completed` or safe `failed` | [ ] |
| 4.3 | Completion | One HTML artifact has MIME, filename, size, SHA-256, and authenticated URL | [ ] |
| 4.4 | Preview | HTML renders only in the sandboxed preview; direct response has private/no-store/nosniff headers | [ ] |
| 4.5 | Other output types | image, canvas, link, report, and doc retain explicit type/file metadata | [ ] |

## 5. Idempotency, rate, and tenant isolation

| Step | Expected | Result |
| --- | --- | --- |
| 5.1 | Repeat same create request/idempotency key | Existing job returned; no second provider call, artifact, or usage event | [ ] |
| 5.2 | Burst beyond `RATE_LIMIT_GENERATION_PER_MIN` | Structured 429 quota/rate failure | [ ] |
| 5.3 | Read job/artifact with another tenant token | 403 context mismatch or 404; no data leak | [ ] |
| 5.4 | Audit/usage inspection | Only redacted resource/provider/status metadata; no key or raw provider body | [ ] |

## 6. Teardown

```bash
make stop
# Keep shared infra running for other work, or use make infra-down when safe.
```

## 7. Sign-off

| Tester | Date | Result | Defects |
| --- | --- | --- | --- |
| | | PASS / FAIL | |
