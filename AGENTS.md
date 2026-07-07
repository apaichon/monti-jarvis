# Agent Instructions

## Project Notes

- Monti Jarvis is an inbound call center service: one Go server, embedded static UI, Gemini text + voice relay, optional shared local infra.
- First product surface is **no-auth**: callers select an AI avatar agent and ask questions by text or voice.
- Use Postgres database `monti_jarvis`, schema `callcenter`; do not write into Jarvis Chat's `jarvis_chat` schema.
- Use Redis DB index `4` with the `monti_jarvis:` prefix.
- Use MinIO bucket/prefix `monti-jarvis/calls/` for future call recordings and artifacts.
- Prefer small, testable Go packages under `internal/`; UI is plain HTML/CSS/JS under `internal/web/public`.
- Workforce agents live in `internal/workforce/`; extend there when adding agents or role prompts.
- Defer ticketing, CRM, auth/KYC, and knowledge-base features unless explicitly requested.

## Skills

Project skills are under `.claude/skills/` (adapted from Jarvis Chat):

- `feature-spec` — scaffold a buildable feature with acceptance criteria
- `sprint-plan` — open or groom a sprint with tasks and ACs
- `debug` — diagnose runtime and test failures
- `doc-audit` — check SDLC doc consistency

## SDLC

Sprint and task docs live in `docs/sdlc/`. Current sprint: `docs/sdlc/sprints/SPRINT-001.md`.