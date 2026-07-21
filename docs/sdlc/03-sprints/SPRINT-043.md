---
id: SPRINT-043
status: planned
start: 2026-07-20
end: 2026-07-24
updated: 2026-07-19
design_pack: review_pending
release_target: v2.17.0
roadmap_sprint: 43
feature: FEAT-0037
platform: Tenant / Platform
depends_on: [SPRINT-014, SPRINT-015, SPRINT-016, SPRINT-039]
---

# SPRINT-043 — Embed Auth, Config Groups & Tenant AI Extensibility

## Goal

Give each tenant safe control over embed authentication and the AI runtime
configuration used by its agents: grouped operational configuration, an
encrypted tenant Gemini key, bounded system prompts, allowlisted call tools,
and tenant-defined skills.

The current public embed remains the default when authentication is disabled;
all new AI configuration is tenant-scoped and must not weaken existing secret
or data-isolation boundaries.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S37, S39, S42) | 14, 14, 12 → **avg 13.3** |
| **Committed** | **14** |

The commitment is close to the trailing average. It includes the six cohesive
roadmap packages and excludes marketplace or multi-provider expansion.

## Commitment

Sprint planning found no unassigned `proposed`/`approved` task files. The
roadmap item is decomposed into six work packages; TASK tickets are created or
refined during the Sprint 43 technical-spec pass.

| Work package | Points | Owner | Outcome |
| --- | ---: | --- | --- |
| Embed auth mode | 3 | dev | Per-tenant/embed `auth=true|false` contract; authenticated mode gates chat/voice while public mode remains compatible |
| Configuration groups | 2 | devops | Core infrastructure env separated from named operational/product config groups with documented loading and restart behavior |
| Tenant Gemini key | 3 | dev | Encrypted-at-rest tenant key storage, masked API contract, tenant-key-first runtime resolution, and platform fallback |
| Tenant system prompt | 2 | dev | Bounded tenant/agent prompt CRUD and safe injection into text and voice orchestration |
| Tenant call tools | 2 | dev | Tenant-scoped allowlisted function definitions with enable/disable and invocation isolation |
| Tenant skills | 2 | dev | CRUD and agent assignment for prompt/tool bundles with tenant isolation and audit-safe validation |

**Committed:** 14 points · task decomposition pending technical-spec approval.

## Scope boundary

### In

- Embed auth mode with the existing customer authentication/session contract.
- Core infra env plus named grouped configuration for operational/product
  settings.
- Encrypted tenant Gemini keys with redacted read paths and runtime fallback.
- Tenant system prompts for chat and voice within length and safety bounds.
- Tenant-scoped allowlisted call tools and tenant-defined skills.
- Tenant isolation, authorization, validation, and regression coverage for each
  persisted configuration surface.

### Out

- Third-party skill marketplace or arbitrary code execution.
- Multi-provider LLM selection beyond Gemini.
- Replacing the platform Gemini default for all tenants.
- Full customer identity redesign, KYC, billing, or quota policy changes.

## Dependencies and design gates

1. Reuse Sprint 14 public embed resolve and iframe behavior; `auth=false` is a
   compatibility requirement.
2. Reuse Sprint 15 scope and Sprint 16 tenant settings/RBAC patterns; every
   persisted object carries tenant ownership and authorization checks.
3. Reuse Sprint 39 tenant/embed configuration surfaces without exposing secret
   values to browser clients.
4. Approve the Sprint 43 workflow, ER, API, UX, and secret-handling design
   before implementation of new storage or runtime resolution.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Feature | [FEAT-0037](../01-features/FEAT-0037-tenant-ai-config-extensibility.md) | `planned` |
| Deep spec | [DES-0039](../02-design/39-tenant-ai-config-extensibility-spec.md) | `review_pending` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §93–96 | `review_pending` |
| ER | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 43 | `review_pending` |
| API | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 43 | `review_pending` |
| UX | [05-ux-ui.md](../02-design/05-ux-ui.md) T22 | `review_pending` |

Implementation is gated on approval of the deep spec, API contract, and
secret-handling contract.

## Verification target

```bash
make test
make build
cd apps/tenant-web && npm run check && npm run build
cd apps/customer-web && npm run check && npm run build
git diff --check
# auth=true blocks unauthenticated embed chat/voice; auth=false preserves public behavior
# grouped config loads deterministically and make restart remains compatible
# tenant Gemini key is encrypted/masked and never returned in plaintext
# prompt, tool, and skill CRUD and runtime use are tenant-isolated
# at least one tool and one skill invoke successfully under tenant isolation tests
```

## Risks

| Risk | Mitigation |
| --- | --- |
| Tenant key leaks through API, logs, or browser state | Encrypt at rest, return metadata only, redact logs, and test secret absence |
| `auth=true` breaks existing public embeds | Default false, preserve resolve contract, and run both-mode regression tests |
| Tools or skills become arbitrary code execution | Persist definitions only, use an allowlist/dispatcher, and reject executable payloads |
| Prompt injection weakens platform safety | Length limits, reserved safety prefix, validation, and explicit precedence rules |
| Config group drift breaks local operations | Document precedence and keep `make restart` smoke coverage |

## Links

- Roadmap: [ROADMAP.md](../00-roadmap/ROADMAP.md) Sprint 43
- Feature: [FEAT-0037](../01-features/FEAT-0037-tenant-ai-config-extensibility.md)
- Depends: [SPRINT-014](SPRINT-014.md), [SPRINT-015](SPRINT-015.md), [SPRINT-016](SPRINT-016.md), [SPRINT-039](SPRINT-039.md)
