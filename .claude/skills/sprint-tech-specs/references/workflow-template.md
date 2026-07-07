# Workflow section template — append to 02-workflow.md

```markdown
## N. <Flow title> (Sprint N)

```mermaid
sequenceDiagram
  participant B as Browser
  participant G as Go :8091
  participant X as internal/<pkg>
  participant DB as Postgres

  B->>G: <METHOD> <path> {<payload>}
  G->>X: <operation>
  alt <failure case>
    G-->>B: <status> {<error>}
  else ok
    X->>DB: <SQL action>
    G-->>B: <status> {<body>}
  end
```

### State: <entity> (if applicable)

| Status | Meaning |
| --- | --- |
| `active` | … |
| `ended` | … |
```