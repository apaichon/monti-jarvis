---
id: DES-0051
title: Caller Desk Branding and Audio Device Settings Specification
status: approved
updated: 2026-07-29
sprint: SPRINT-056
owner: SA
feature: FEAT-0047
release_target: v2.28.0
---

# DES-0051 — Caller Desk Branding + Mic/Speaker Settings

**Sprint:** SPRINT-056 · **Release target:** v2.28.0  
**Feature:** [FEAT-0047](../01-features/FEAT-0047-caller-desk-branding-audio-devices.md)  
**Tasks:** [TASK-0199](../04-tasks/TASK-0199.md), [TASK-0200](../04-tasks/TASK-0200.md),
[TASK-0201](../04-tasks/TASK-0201.md)  
**Mockups:** [call-page.png](mockups/s56-caller-desk-branding/call-page.png) ·
[tenant-list.png](mockups/s56-caller-desk-branding/tenant-list.png) ·
[composite](mockups/s56-caller-desk-branding/new-call-design-composite.png)  
**Related:** [DES-0049](49-customer-portal-tenant-list-spec.md) (S54 list→desk),
[DES-0037](37-theme-color-customization-spec.md) (public theme branding)

## 1. Goals

1. Tenant directory and call desk present **large Monti** and **large company
   logos** consistent with product mockups.
2. Callers can select **microphone** and **speaker** devices for voice.
3. No new backend audio registry; client-side media APIs only.
4. Preserve S54 routing and non-technical **Live · OK** system status.

## 2. Non-goals

- Theme/CMS redesign for logo upload (use existing brand/theme fields)
- Server-side device inventory
- S55 topic statistics
- Noise cancellation product UI

## 3. Branding policy

| Surface | Monti mark | Company mark |
| --- | --- | --- |
| Tenant list `/` | Large hero (product logo + wordmark) | Large logo on each card |
| Call desk `/t/{slug}` | Large hero in control rail | Selected brand card with logo + name |

**Sources**

- Monti: static product asset (e.g. `/images/monti-logo.png` or approved hero art)
- Company: public brand `logo_url` and/or published theme `branding.logo_url`
- Fallback: monogram from brand name (1–2 letters) on colored disk

**Isolation:** desk company logo always follows selected tenant only.

## 4. Audio device policy

| Concern | Rule |
| --- | --- |
| Enumerate | After permission (or when labels available), list `audioinput` / `audiooutput` |
| Default | Browser default if no preference or device gone |
| Persist | `localStorage` keys `monti_jarvis:audio_input_id`, `monti_jarvis:audio_output_id` (+ session mirror optional) |
| Mic apply | `getUserMedia({ audio: { deviceId: { exact|ideal } } })` into voice pipeline |
| Speaker apply | `HTMLMediaElement.setSinkId` / AudioContext sink when supported |
| Denied | Non-technical copy; voice start may re-prompt once |

## 5. Data model

**No Postgres/ClickHouse migration.** Client preference only.

| Store | Key | Value |
| --- | --- | --- |
| localStorage | `monti_jarvis:audio_input_id` | deviceId string |
| localStorage | `monti_jarvis:audio_output_id` | deviceId string |
| (existing) | `monti_jarvis:selected_tenant` | S54 selection |

## 6. API summary

No new REST endpoints required. Reuse:

| Method | Path | Use |
| --- | --- | --- |
| GET | `/api/public/brands` | List logos/names |
| GET | `/api/public/brands/{slug}` | Selected brand |
| GET | `/api/public/theme/{tenant_id}` | Published branding tokens/logo |

Optional later (out of sprint): none.

## 7. Client modules

| Module | Responsibility |
| --- | --- |
| `TenantPicker.svelte` | Directory branding layout |
| `CustomerDesk.svelte` | Desk hero + company card + audio settings entry |
| `lib/audio/devices.ts` (new) | enumerate, prefer, persist, apply |
| `lib/voice/gemini.ts` | Accept deviceId options for capture/playback |

## 8. Verification

```bash
cd apps/customer-web && npm run check
# Manual per TASK-0201 checklist
```

## 9. Implementation sequence

1. **TASK-0199** — branding layouts from mockups  
2. **TASK-0200** — device settings + voice wiring  
3. **TASK-0201** — UAT checklist + verification  

## See also

- Workflow §127–129 · ER Sprint 56 · API Sprint 56 · UX C56
- [SPRINT-056](../03-sprints/SPRINT-056.md)
