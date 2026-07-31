# Agent Instructions

## Project Notes

- Monti Jarvis is an inbound call center service: one Go server, Svelte customer portal, Gemini text + voice relay, optional shared local infra.
- First product surface is **no-auth**: callers select an AI avatar agent and ask questions by text or voice.
- Use Postgres database `monti_jarvis`, schema `callcenter`; do not write into Jarvis Chat's `jarvis_chat` schema.
- Use Redis DB index `4` with the `monti_jarvis:` prefix.
- Use MinIO bucket `monti-jarvis` with prefixes `calls/` and `km/`.
- ClickHouse database `monti_jarvis` for KM embeddings (`km_embeddings`).
- Prefer small, testable Go packages under `internal/`; customer UI in `apps/customer-web/`; platform admin in `apps/platform-admin-web/`.
- Workforce agents live in `internal/workforce/`; KM/RAG in `internal/km`, `internal/rag`, `internal/clickhouse`, `internal/scope`.
- Defer full auth/KYC/ticketing unless explicitly in sprint scope.

## Agents (`.claude/agents/`)

| Agent | Role |
| --- | --- |
| `pm` | Sprint planning, features, roadmap, release-cut |
| `dev` | Go + Svelte implementation |
| `tester` | AC verification, manual UAT |
| `devops` | Infra, compose, schema, Makefile |
| `doc-manager` | SDLC doc hygiene |

## Skills (`.claude/skills/`)

| Skill | Use when |
| --- | --- |
| `km-context` | Start of any agent task — load active sprint slice |
| `km-sync` | End of task — propagate status across SDLC docs |
| `sprint-plan` | Open or groom a sprint |
| `sprint-tech-specs` | Per-sprint design pack — workflow, ER, API, UX/UI ASCII + mapping |
| `sprint-status` | Standup / mid-sprint progress report |
| `feature-spec` | Scaffold a feature with ACs |
| `release-cut` | Sprint close — VERSION + tag |
| `manual-test-doc` | Generate UAT checklist |
| `debug` | Bugs, crashes, "not working" |
| `doc-audit` | Sprint-close doc consistency check |
| `doc-control` | Stamp sprint/task/feature metadata |
| `obsidian-graph` | Optional SDLC visual vault |

## SDLC

SDLC index: `docs/sdlc/README.md`

| Prefix | Path |
| --- | --- |
| `00-roadmap` | `docs/sdlc/00-roadmap/ROADMAP.md` |
| `01-features` | `docs/sdlc/01-features/` |
| `02-design` | `docs/sdlc/02-design/` (architecture, workflow, ER, API, UX/UI) |
| `03-sprints` | `docs/sdlc/03-sprints/` |
| `04-tasks` | `docs/sdlc/04-tasks/` |
| `05-test-scenarios` | `docs/sdlc/05-test-scenarios/` |
| `06-manual-tests` | `docs/sdlc/06-manual-tests/` |
| `07-deployment` | `docs/sdlc/07-deployment/` |
| `08-readiness` | `docs/sdlc/08-readiness/` |

**Current sprint:** residual S45 UAT; S49 held. Latest shipped: `SPRINT-063` **v2.35.1** · `SPRINT-060`–`SPRINT-062` **v2.35.0**.

**Shipped:** v2.35.1 `SPRINT-063` platform Leads inbox Book Demo rendering fix · v2.35.0 `SPRINT-060`–`SPRINT-062` tenant Gemini enforcement, top-bar status, and referral redemption · v2.32.0 `SPRINT-059` call conversation UX revamp · v2.31.0 `SPRINT-058` portal UI language selector EN/TH/JA · v2.30.0 `SPRINT-057` logo/social preview · v2.29.0 `SPRINT-056` caller desk branding + mic/speaker · v2.28.0 `SPRINT-055` call-center statistics group by topic · v2.27.0 `SPRINT-054` customer portal tenant list · v2.26.0 `SPRINT-053` conversation auto-register + app version · v2.25.0 `SPRINT-051` commercial · v2.24.0 `SPRINT-052` tenant avatar library · v2.23.0 `SPRINT-050` admin promotional package grant · v2.22.0 Sprint 41 security · v2.21.0 `SPRINT-048` product-web growth · v2.20.0 `SPRINT-046` referral bonus quota · v2.19.0 `SPRINT-046` referral foundation · v2.18.1/v2.18.0 `SPRINT-045` AiaaS packages · v2.16.0 `SPRINT-042` tenant UX bugfix · v2.15.0 `SPRINT-039` theme branding & colors · v2.14.0 `SPRINT-037` embed framework SDKs · v2.13.0 `SPRINT-032` platform billing usage readiness and reconciliation · v2.12.0 `SPRINT-031` platform billing/quota/AI usage · v2.11.0 `SPRINT-030` platform overall call-center statistics · v2.10.0 `SPRINT-029` platform system performance monitoring · v2.9.0 `SPRINT-028` cross-tenant audit log · v2.8.0 `SPRINT-027` mobile call API/SDK · v2.7.0 `SPRINT-026` tenant system performance monitoring · v2.6.0 `SPRINT-025` tenant call-center statistics/quota · v2.5.0 `SPRINT-024` customer satisfaction/statistics · v2.4.0 `SPRINT-023` tickets/human escalation · v2.3.0 `SPRINT-022` conversation records/knowledge gaps · v2.2.0 `SPRINT-021` authenticated workforce/quota · v2.1.0 `SPRINT-020` customer auth/domain enforcement · v2.0.0 `SPRINT-019` customer import/domain integration · v1.9.0 `SPRINT-018` customer tiers · v1.8.0 `SPRINT-017` test/preview · v1.7.0 `SPRINT-016` settings/locale/limits · v1.6.0 `SPRINT-015` tenant KM · v1.5.0 `SPRINT-014` embed · v1.4.0 `SPRINT-013` quota · v1.3.1 commerce harden · v1.3.0 `SPRINT-008`–`012` commerce · v0.8.0 `SPRINT-007` · v0.7.0 `SPRINT-006` · v0.6.0 `SPRINT-005` · v0.5.0 `SPRINT-004` · v0.4.0 `SPRINT-003` · v0.3.0 `SPRINT-002` · v0.2.0 `SPRINT-001` · v0.1.0 prototype

**Prod gate:** Before customer production launch (post tenant customer-user auth S19–20), verify **rate limit + quota management** under multi-user load (see SPRINT-016 / SPRINT-017 notes).

**KM ops:** `docs/KM_SETUP.md` · **API:** `docs/sdlc/02-design/04-api-spec.md` · **Deploy:** `docs/sdlc/07-deployment/LOCAL-DEV.md` · **UAT:** `docs/sdlc/06-manual-tests/`
