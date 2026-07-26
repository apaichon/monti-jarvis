---
id: SPRINT-048
status: shipped
start: 2026-08-15
end: 2026-08-21
updated: 2026-07-26
design_pack: approved
release: v2.21.0
release_target: v2.21.0
roadmap_sprint: 48
feature: FEAT-0040
platform: Customer / Growth / Tenant
depends_on: [SPRINT-004, SPRINT-006, SPRINT-009, SPRINT-017, SPRINT-020, SPRINT-031, SPRINT-039, SPRINT-046]
goal: "Ship product web marketing shell, lead capture, demo/register conversion, and sales lead ops without bypassing existing commerce authorities."
parallel_track: true
---

# SPRINT-048 — Product Web Sales, Marketing, Demo, and Tenant Conversion

## Goal

Build a public Monti product website that turns marketing traffic into
qualified leads, live demos, tenant signups, and paid package customers using
existing registration, KYC, checkout, and referral attribution.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S37, S39, S42) | 14, 14, 12 → **avg 13.3** |
| **Proposed capacity** | **13** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0170](../04-tasks/TASK-0170.md) Product-web shell + marketing pages | 4 | completed | dev | `apps/product-web` dark brand shell, routes, SEO, static serve at `/product/` |
| [TASK-0171](../04-tasks/TASK-0171.md) Lead capture API + store + consent | 3 | completed | dev | Public lead endpoints, dedupe, rate limit, consent, lifecycle |
| [TASK-0172](../04-tasks/TASK-0172.md) Pricing catalog + conversion redirects | 3 | completed | dev | Public packages API, attribution helper, allowlisted CTAs |
| [TASK-0173](../04-tasks/TASK-0173.md) Sales lead console + lifecycle notes | 2 | completed | dev | Platform `/admin/leads` + API |
| [TASK-0174](../04-tasks/TASK-0174.md) Funnel events + verification | 1 | completed | tester/dev | Funnel API + unit tests; merged/tagged **v2.21.0** |
| **Total** | **13** | | | **13/13 shipped on `main` as v2.21.0** |

## Scope boundary

### In

- Product-web Svelte SPA served at `/product/` (apex mapping is deploy config)
- Marketing pages: Home, Product, Solutions, Resources, Pricing, About, Contact, Demo entry
- Lead capture forms + backend store + sales ops surface
- Catalog-driven public pricing (read-only)
- Safe redirects into customer demo, tenant register, and authenticated billing
- Attribution persistence for `utm_*` + referral code + landing path
- Funnel measurement without raw conversation content

### Out

- CRM replacement, bulk email campaigns, paid ad creative production
- Public payment / entitlement grant from pricing page
- Hard-coded package prices bypassing catalog
- S47 Langfuse, S44 generative AI, S38 multi-brand hub
- Changing S46 referral qualification rules (only consume attribution when present)

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0040](../01-features/FEAT-0040-product-web-growth.md) | `in_progress` |
| Deep spec | [DES-0043](../02-design/43-product-web-growth-spec.md) | `approved` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §105–108 | `approved` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 48 | `approved` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 48 | `approved` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) P48 / A48 | `approved` |

## Conversion funnel

```text
Advertising / SEO / referral
        ↓
Product web (/product) + attribution
  ├─ Try live demo → / (customer no-auth) → CTA back to contact/register
  ├─ Book demo / Contact → lead form → sales lifecycle
  └─ Pricing / Start now → /tenant/register → KYC → /tenant/billing checkout
```

## Verification target

```bash
make product-web
make build
go test ./internal/leads ./internal/store ./cmd/server -count=1
make test
git diff --check
# Manual: /product pages, lead form, public pricing, register redirect, demo CTA
```

Manual checklist: [SPRINT-048-manual.md](../06-manual-tests/SPRINT-048-manual.md) (generated at VERIFY)

## Risks

| Risk | Mitigation |
| --- | --- |
| Pricing content drifts from catalog | Public pricing always reads active packages API |
| Open redirects / attribute injection | Allowlist destinations; sanitize utm/referral keys |
| Lead spam | Redis rate limit + honeypot + dedupe key |
| Parallel with S45 residual UAT | Worktree isolation; no commerce authority changes |
| Brand claims vs legal | Ship placeholder metrics only from approved content flags |

## Worktree / parallel build

Implementation runs in git worktree `SPRINT-048` so residual S45/S46 docs and
main-branch churn stay isolated. Only one release cut ships S48 at close.
