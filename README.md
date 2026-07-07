# Monti Jarvis — Inbound Call Center

Voice-first inbound call center service built on the same lean stack as [Jarvis Chat](https://github.com/libra/jarvis): one Go server, embedded browser UI, Gemini text + voice relay, and optional shared local infra.

## First release scope

- **AI avatar workforce** — select Ava, Max, Luna, or Neo; each agent has role-specific prompts and voice
- **Inbound Q&A** — text chat and voice-to-voice conversation to answer caller questions
- **No auth / no ticketing yet** — login, KYC, CRM, and ticket flows are planned later

## Run

```bash
cp infra/.env.dev.example infra/.env.dev
# set GEMINI_API_KEY in infra/.env.dev
make infra-check
make infra-init
make start
```

Open http://localhost:8091.

## Commands

```bash
make run          # foreground server
make start        # background server, logs in .run/server.log
make stop
make status
make logs
make infra-check  # verify shared Postgres, Redis, MinIO containers/ports
make infra-init   # create monti_jarvis DB/schema and MinIO bucket
make test
```

## API

| Method | Path | Purpose |
| --- | --- | --- |
| GET | `/healthz` | Liveness + Gemini/voice status |
| GET | `/api/workforce` | List AI avatar agents |
| POST | `/api/chat` | Text Q&A with `agent_id`, `topic`, `message` |
| GET | `/ws/voice?agent=ava` | Voice call with selected agent |
| GET | `/api/infra` | Postgres / Redis / MinIO health |

## Infra isolation

Reuses the shared local stack but keeps Monti data separate from Jarvis Chat:

| Resource | Monti Jarvis | Jarvis Chat (reference) |
| --- | --- | --- |
| Postgres DB | `monti_jarvis` | `jarvis_chat` |
| Postgres schema | `callcenter` | `chat` |
| Redis DB index | `4` | `3` |
| Redis prefix | `monti_jarvis:` | `jarvis_chat:` |
| MinIO bucket | `monti-jarvis` | `jarvis-chat` |

## Workforce

| Agent | Role | Voice |
| --- | --- | --- |
| Ava | General Support | Aoede |
| Max | Billing Specialist | Charon |
| Luna | Technical Support | Kore |
| Neo | Triage Bot | Puck |

## Planned later

- Caller authentication and account lookup
- Ticket creation and CRM integration
- Knowledge-base search over pgvector
- Outbound campaigns and supervisor dashboard
- Call recording and compliance retention