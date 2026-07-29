# SPRINT-054 — Customer portal tenant list manual UAT

**Feature:** FEAT-0045 · **Design:** DES-0049  
**Updated:** 2026-07-29

## Preconditions

- Go server on `:8091` with Postgres schema `callcenter`
- ≥2 tenants with brands `listed=true` and `platform_listed=true`, status active
- Customer web on `:5173` (or static build served by server)
- Optional: unlisted tenant for negative list checks

## Scenarios

### 1. Open portal without `tenant_id`

- [ ] Open `/` with no query string
- [ ] **Expected:** brand picker (or empty state), not a broken desk requiring tenant

### 2. Select tenant A → desk

- [ ] Search/select brand A
- [ ] URL becomes `/t/{slugA}` (no required `?tenant_id=`)
- [ ] Theme/brand chrome and workforce load for A only
- [ ] Sign in + chat/voice work under A

### 3. Switch A → B isolation

- [ ] On A desk, start a short chat (or leave welcome messages)
- [ ] Click **← Brands / เปลี่ยนแบรนด์**
- [ ] Select brand B
- [ ] **Expected:** workforce is B only; no A transcript; no A agents

### 4. Legacy deep link

- [ ] Open `/?tenant_id={id-or-slug-of-A}`
- [ ] **Expected:** redirects/normalizes to `/t/{slugA}` and desk for A

### 5. Path deep link

- [ ] Open `/t/{slugA}` directly
- [ ] **Expected:** desk for A without picker first

### 6. Unlisted / unknown

- [ ] Open `/t/not-a-real-tenant`
- [ ] **Expected:** error + back to brands; no desk
- [ ] Unlisted tenant never appears in picker list

### 7. Empty directory

- [ ] (Staging) with zero listed brands: picker shows empty state, no crash

### 8. API directory

```bash
curl -s 'http://127.0.0.1:8091/api/public/brands?limit=20' | jq .
curl -s 'http://127.0.0.1:8091/api/public/tenants?limit=20' | jq .
curl -s -o /dev/null -w '%{http_code}\n' 'http://127.0.0.1:8091/api/public/brands/not-a-real-tenant'
```

- [ ] Brands and tenants alias return same list shape
- [ ] Unknown brand returns 404
- [ ] Response has no secrets / admin email / platform_listed

### 9. Transport header preference

- [ ] After pick, network tab shows `X-Tenant-Id` on workforce/chat (query may still exist as fallback)

### 10. Embed unchanged

- [ ] `/embed` still uses embed key path (not brand picker)

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Dev | | |
| Tester | | |
