# API section template — append to 04-api-spec.md

```markdown
## <Domain name>

**Auth:** `platform_admin` | `tenant_admin` | public (when `AUTH_DISABLED=true`)

### `<METHOD> <path>`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `field` | string | yes | … |

**Request**

```json
{ "field": "value" }
```

**Response `200`**

```json
{ "id": "...", "status": "active" }
```

**Errors**

| Code | When |
| --- | --- |
| `401` | Missing/invalid Bearer |
| `403` | Wrong role |
| `404` | Resource not found |
| `409` | Conflict (duplicate active entitlement) |
```