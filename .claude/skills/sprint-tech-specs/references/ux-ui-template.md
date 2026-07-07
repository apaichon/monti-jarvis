# UX/UI section template — append to 05-ux-ui.md

```markdown
## Sprint N — <title>

<One line: what UX changes — e.g. "API-only; no customer portal change.">

### Screen map → API

| UI zone | User action | API / WS |
| --- | --- | --- |
| T1 Terminal | Login as platform admin | `POST /api/auth/login` |
| T2 Terminal | List packages | `GET /api/platform/packages` |
| T3 Terminal | Assign package | `POST /api/platform/tenants/{id}/entitlement` |

### Operator layout (API-only sprint)

```
┌─────────────────────────────────────────────────────────────┐
│  Terminal / REST Client (platform@monti.local)              │
├─────────────────────────────────────────────────────────────┤
│  1. POST /api/auth/login  →  access_token                   │
│  2. GET  /api/platform/packages  →  catalog[]             │
│  3. POST /api/platform/tenants/demo/entitlement           │
│         { "package_id": "pkg-pro" }                         │
│  4. GET  /api/entitlements/me  (tenant_admin token)        │
└─────────────────────────────────────────────────────────────┘
```

### Flow A — <primary interaction>

```
User action (zone T1)
    │
    ├─► POST /api/...
    │       │
    │       ▼
    └─► Display result / next step
```

### Component → file (when UI ships)

| Component | Path |
| --- | --- |
| … | `apps/customer-web/src/...` |
```

## Customer portal sprint (full layout)

Reuse zone labels from existing `05-ux-ui.md`:
- **A*** = control panel (avatar, agents, call controls)
- **B*** = workspace (topics, chat, composer, infra line)

Always add new rows to the master **Screen map → API** table at the top when customer UI changes.