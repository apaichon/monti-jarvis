# SPRINT-052 — Tenant avatar library manual UAT

**Feature:** FEAT-0043 · **Design:** DES-0047  
**Updated:** 2026-07-28

## Preconditions

- Tenant admin login with active package (`max_ai_employees` known, e.g. 2)
- Server restarted so `owner_tenant_id` schema is applied

## Scenarios

### 1. Create beyond package active limit

- [ ] Open `/tenant/avatars`
- [ ] Create 3 avatars when package limit is 2
- [ ] **Expected:** all three succeed as **Library** (inactive); cap meter still 0/2 until activate

### 2. Activate to cap then block

- [ ] Activate avatar A → 1/2
- [ ] Activate avatar B → 2/2
- [ ] Activate avatar C → error / blocked (`quota_exceeded` or disabled Activate)
- [ ] Deactivate A → Activate C succeeds

### 3. Workforce only shows active

- [ ] Open customer workforce / demo agent picker
- [ ] **Expected:** only active avatars appear; library drafts hidden

### 4. Isolation

- [ ] Tenant A library not visible to tenant B
- [ ] Tenant B cannot activate A’s avatar ids

### 5. Platform catalog assign still works

- [ ] Platform assign of catalog avatar still enforces active cap (existing path)

### 6. Portrait upload

- [ ] Create avatar with portrait selected in modal
- [ ] UI shows recommended size **512×512**, types **JPEG/PNG/WebP/GIF**, max **4 MB**
- [ ] After create, library row shows portrait circle
- [ ] Replace portrait from library row Upload/Replace
- [ ] Reject oversized file (>4MB) with clear error
- [ ] Platform-owned assigned avatars cannot upload from tenant UI

### 7. Speaker voice (AI Studio generate-speech)

- [ ] Create form shows voice dropdown with labels like **Aoede — Breezy**, **Puck — Upbeat**
- [ ] List includes ~30 Gemini speaker setting names
- [ ] Create with selected voice; row shows that voice
- [ ] Change voice from library row dropdown; persists after reload
- [ ] Link/hint references https://aistudio.google.com/generate-speech

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Tester | | |
| Dev | | |
