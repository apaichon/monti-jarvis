# Monti Jarvis — SDLC Documentation

Numbered folders keep artifacts ordered and easy to navigate.

| Prefix | Folder | Contents |
| --- | --- | --- |
| `00` | [00-roadmap/](00-roadmap/) | Roadmap (36 core + S37 embed SDKs + S38 central brand portal) |
| `01` | [01-features/](01-features/) | Feature specs (`FEAT-NNNN`) |
| `02` | [02-design/](02-design/) | Architecture, workflow, ER, API, UX/UI — use **`sprint-tech-specs`** skill per sprint |
| `03` | [03-sprints/](03-sprints/) | Sprint plans (`SPRINT-NNN`) |
| `04` | [04-tasks/](04-tasks/) | Task tickets (`TASK-NNNN`) |
| `05` | [05-test-scenarios/](05-test-scenarios/) | AC → scenario matrix (auto vs manual) |
| `06` | [06-manual-tests/](06-manual-tests/) | Sprint UAT checklists |
| `07` | [07-deployment/](07-deployment/) | Local dev + production deploy guides |
| `08` | [08-readiness/](08-readiness/) | Pre-release / pre-demo gates |

**Blueprint (business):** [`docs/monti_multi_tenant_ai_call_center_blueprint.md`](../monti_multi_tenant_ai_call_center_blueprint.md)

**Current sprint:** Sprint 43 — Embed Auth, Config Groups, and Tenant AI Extensibility (**planned**)

**Parallel build:** _none_

**Latest shipped:** [`SPRINT-042`](03-sprints/SPRINT-042.md) — Tenant UX bugfix · **v2.16.0**

**Planned next:** Sprint 43 — Embed Auth, Config Groups, and Tenant AI Extensibility

**Shipped:** v2.13.0 [`SPRINT-032`](03-sprints/SPRINT-032.md) · v2.12.0 [`SPRINT-031`](03-sprints/SPRINT-031.md) · v2.11.0 [`SPRINT-030`](03-sprints/SPRINT-030.md) · v2.10.0 [`SPRINT-029`](03-sprints/SPRINT-029.md) · v2.9.0 [`SPRINT-028`](03-sprints/SPRINT-028.md) · v2.8.0 [`SPRINT-027`](03-sprints/SPRINT-027.md) · v2.7.0 [`SPRINT-026`](03-sprints/SPRINT-026.md) · v2.6.0 [`SPRINT-025`](03-sprints/SPRINT-025.md) · v2.5.0 [`SPRINT-024`](03-sprints/SPRINT-024.md) · v2.4.0 [`SPRINT-023`](03-sprints/SPRINT-023.md) · v2.3.0 [`SPRINT-022`](03-sprints/SPRINT-022.md) · v2.2.0 [`SPRINT-021`](03-sprints/SPRINT-021.md) · v2.1.0 [`SPRINT-020`](03-sprints/SPRINT-020.md) · v2.0.0 [`SPRINT-019`](03-sprints/SPRINT-019.md) · v1.9.0 [`SPRINT-018`](03-sprints/SPRINT-018.md) · v1.8.0 [`SPRINT-017`](03-sprints/SPRINT-017.md) · v1.7.0 [`SPRINT-016`](03-sprints/SPRINT-016.md) · v1.6.0 [`SPRINT-015`](03-sprints/SPRINT-015.md) · v1.5.0 [`SPRINT-014`](03-sprints/SPRINT-014.md) · v1.4.0 [`SPRINT-013`](03-sprints/SPRINT-013.md) · v1.3.0 SPRINT-008–012 · v0.8.0 [`SPRINT-007`](03-sprints/SPRINT-007.md) · v0.2.0 [`SPRINT-001`](03-sprints/SPRINT-001.md)

**Production gate:** After SPRINT-019–020 customer identity, verify quota, rate-limit, tier overrides, and tenant isolation under multi-user load before opening customer traffic.

**Ops:** [`docs/KM_SETUP.md`](../KM_SETUP.md) · **Deploy:** [`07-deployment/LOCAL-DEV.md`](07-deployment/LOCAL-DEV.md) · **UAT:** [`06-manual-tests/`](06-manual-tests/)
