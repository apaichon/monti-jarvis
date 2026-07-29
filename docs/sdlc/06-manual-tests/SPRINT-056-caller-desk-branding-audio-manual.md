# SPRINT-056 — Caller desk branding + mic/speaker manual UAT

**Feature:** FEAT-0047 · **Design:** DES-0051  
**Updated:** 2026-07-29

## Preconditions

- Customer web on `:5173` or server static build
- ≥2 listed public brands with different logos when possible
- Chrome recommended for speaker `setSinkId`

## Scenarios

### 1. Tenant list large branding

- [ ] Open `/`
- [ ] **Expected:** large Monti hero + wordmark; brand cards with large logos
- [ ] Card without logo shows monogram

### 2. Select brand → desk branding

- [ ] Call brand A → `/t/{slugA}`
- [ ] **Expected:** large Monti hero; company card with A logo/name
- [ ] Change brand → B; company card updates to B only

### 3. Mic selection

- [ ] Open Audio settings → Refresh devices (allow mic if prompted)
- [ ] Select a non-default microphone if available
- [ ] Start call
- [ ] **Expected:** voice uses selected mic (or clear error if denied)

### 4. Speaker selection

- [ ] Select a speaker when browser supports it
- [ ] Start call / play agent audio
- [ ] **Expected:** output on selected device when supported; otherwise system default note

### 5. Permission denied

- [ ] Block microphone for the site
- [ ] Refresh devices or Start call
- [ ] **Expected:** friendly copy (can still chat); no raw exception dump

### 6. Live · OK status

- [ ] Desk header shows **Live · OK** (or Limited/Unavailable), not Postgres/Redis names

### 7. Preference persist

- [ ] Select mic/speaker, reload desk
- [ ] **Expected:** same device ids restored from localStorage

## Sign-off

| Role | Date | Result |
| --- | --- | --- |
| Dev | | |
| Tester | | |
