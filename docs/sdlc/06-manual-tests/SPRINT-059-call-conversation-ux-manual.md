# SPRINT-059 — Call Conversation UX Revamp Manual UAT

**Feature:** FEAT-0051 · **Design:** DES-0055 · **Tasks:** TASK-0212–0215

## Preconditions

- Customer portal available at `/` or `/t/{tenant}`.
- Tenant has an active avatar and customer call access.
- Desktop and mobile-sized browsers are available.
- A target device with microphone and speaker is available for media smoke.

## Recorded cases

| ID | Scenario | Expected | Result |
| --- | --- | --- | --- |
| U59-01 | Desktop conversation layout | Avatar remains prominent; context rails, transcript, and composer do not overlap | Pass — user screenshot review and follow-up layout fixes |
| U59-02 | Left context sections | Tenant, customer, and device sections start collapsed and expand independently | Pass — user screenshot review |
| U59-03 | Center workspace scroll | Subtitle, call controls, transcript, finish action, and bottom composer remain reachable | Pass — structural review and customer-web build |
| U59-04 | Light/dark mode | Circular sun/moon icon switches mode and persists preference | Pass — code review and customer-web build |
| U59-05 | Avatar variants | Dark/light portrait is selected for its mode; default portrait remains fallback | Pass — customer/tenant checks and builds |
| U59-06 | Mobile layout | Avatar, controls, transcript, composer, and compact context rows remain usable without horizontal overflow | Pass — responsive implementation and prior visual capture |
| U59-07 | EN/TH/JA labels | Call controls and context labels fit and preserve tenant/avatar names | Pass — catalog/build verification |
| U59-08 | Keypad | Pointer or keyboard number entry appends to the text composer; close/delete work | Pass — code review and customer-web build |
| U59-09 | Live speaker/microphone | Speaker mute changes playback gain; microphone mute disables active input tracks | Follow-up — run on production target device after deployment |

## Automated release gate

```bash
go test ./...
cd apps/customer-web && npm run check && npm run build
cd apps/tenant-web && npm run check && npm run build
git diff --check
```

## Sign-off

| Role | Date | Result | Notes |
| --- | --- | --- | --- |
| Tester | 2026-07-30 | Pass with target-device follow-up | User-directed screenshot UAT plus automated release gate; U59-09 remains a deployment smoke item |
| PM | 2026-07-30 | Accepted | User explicitly requested Sprint 59 close, merge, ROADMAP update, and version tag |
