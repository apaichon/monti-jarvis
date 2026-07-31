---
id: RELEASE-v2.35.0
status: ready
release: v2.35.0
date: 2026-07-31
sprints: [SPRINT-060, SPRINT-061, SPRINT-062]
---

# Release v2.35.0

Combined release train for tenant-owned Gemini key enforcement, tenant Gemini
status simplification, and referral-code bonus quota redemption.

## Included

- Production tenant AI uses only validated tenant-owned Gemini keys.
- AI Settings supports bounded key testing and safe readiness metadata.
- Tenant top bar shows localized Gemini state; tenant performance navigation is
  removed while platform monitoring remains.
- Referral codes can be validated and applied idempotently to separate bonus
  quota grants.
- Tenant redemption history and platform inspection/reversal are available.

## Database

Apply the migrations added by Sprint 60-62 through the normal startup migration
path before serving the new APIs.

## Configuration

- Production must not enable `ALLOW_PLATFORM_GEMINI_FALLBACK`.
- Redis DB index remains `4` and limiter keys use the `monti_jarvis:` prefix.
- A tenant Gemini key is usable only after successful validation.

## Verification

- `go test ./... -count=1`: 303 passed in 38 packages.
- `go vet ./...`: passed.
- Postgres referral lifecycle integration: passed; zero test fixtures remain.
- Customer web check/build: passed.
- Tenant web check/build: passed.
- Platform-admin web check/build: passed with pre-existing accessibility
  warnings only.
- Version synchronization test: passed.

## Deferred manual check

Run the credentialed Gemini browser scenarios in
`docs/sdlc/06-manual-tests/SPRINT-060-062-manual.md` in the target environment.
