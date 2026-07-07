# ER diagram delta template — extend 03-er-diagram.md

## New Postgres entities (mermaid block)

```mermaid
erDiagram
  tenants ||--o{ tenant_entitlements : has
  packages ||--o{ tenant_entitlements : grants
  package_rule_schemas ||--o{ package_limits : shapes
  packages ||--|| package_limits : defines

  package_rule_schemas {
    text id PK
    int version UK
    jsonb fields
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  packages {
    text id PK
    text slug UK
    text name
    text status
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }

  package_limits {
    text package_id PK_FK
    text rules_schema_id FK
    jsonb rules
    timestamptz created_at
    timestamptz updated_at
    text created_by
    text updated_by
  }
```

## Future entities table row

```markdown
| N (in progress) | `table_a`, `table_b` (all with audit columns) |
```

## Audit reminder (if new table)

Every new `callcenter` table includes:
`created_at`, `updated_at`, `created_by`, `updated_by` + `touch_updated_at()` trigger.