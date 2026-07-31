# SPRINT-060-062 Manual UAT

**Status:** release accepted; credentialed browser UAT follow-up  
**Automated verification:** passed 2026-07-31

## Automated evidence

- [x] `go test ./internal/store ./internal/gemini ./internal/env ./cmd/server -count=1`
- [x] `go test ./... -count=1` (303 tests in 38 packages)
- [x] `go vet ./...`
- [x] Postgres integration: apply, idempotent retry, expired/self rejection, tenant isolation, platform list, reverse, reverse retry
- [x] `apps/customer-web`: `npm run check` and `npm run build`
- [x] `apps/tenant-web`: `npm run check` and `npm run build`
- [x] `apps/platform-admin-web`: `npm run check` and `npm run build`
- [x] `git diff --check`

The Svelte checks report only pre-existing accessibility warnings in unrelated
avatar, package, lead, profile, entitlement, and KYC views.

## SPRINT-060 - Tenant Gemini key

- [ ] Sign in as an active tenant admin and open AI Settings.
- [ ] Save a proposed key and confirm status is `Saved - untested`.
- [ ] Confirm text and voice runtime reject the untested key.
- [ ] Test a valid key and confirm masked last four digits and validation time.
- [ ] Confirm text and voice runtime use the validated tenant key.
- [ ] Test an invalid key and confirm no provider secret or raw provider response is shown.
- [ ] Delete the key and confirm production runtime returns `tenant_gemini_key_required`.
- [ ] Confirm create/replace/delete/test mutations appear in the audit log without request bodies.
- [ ] Confirm `/readyz` reports `tenant_gemini_env_fallback_disabled=true` in production.

## SPRINT-061 - Gemini status top bar

- [ ] Confirm tenant navigation has no Monitoring/System Performance link.
- [ ] Confirm top-bar labels for ready, missing, invalid, degraded/unavailable.
- [ ] Confirm labels render in English, Thai, and Japanese.
- [ ] Confirm save/test/delete in AI Settings refreshes the top-bar state immediately.
- [ ] Confirm the status chip opens AI Settings.
- [ ] Confirm platform-admin Monitoring still loads.

## SPRINT-062 - Referral redemption

- [ ] As a tenant admin, validate and apply another active tenant's code.
- [ ] Confirm the redemption history and bonus quota/expiry refresh after apply.
- [ ] Retry the same code and confirm no second grant.
- [ ] Confirm own, expired, disabled, and inactive-owner codes are rejected safely.
- [ ] As platform admin, filter redemptions by tenant, code, and status.
- [ ] Reverse an applied redemption and confirm tenant remaining bonus updates.
- [ ] Repeat reverse and confirm it is idempotent.
- [ ] Confirm apply/reverse mutations appear in the audit log without request bodies.

## Deferred requirements

- A real Gemini credential is required for provider connectivity UAT.
- Authenticated browser sessions are required for the tenant and platform UI checks.
