---
id: FEAT-0048
title: "Caller desk branding revamp + mic/speaker device settings"
status: shipped
release: v2.29.0
roadmap_sprint: 56
priority: A+
depends_on: [SPRINT-001, SPRINT-005, SPRINT-039, SPRINT-054]
design: DES-0052
design_spec: ../02-design/52-caller-desk-branding-audio-devices-spec.md
updated: 2026-07-29
---

# FEAT-0048: Caller Desk Branding + Mic/Speaker Settings

## Purpose

Make the customer portal feel like a branded multi-company hub: large Monti
and per-company logos on the tenant list and call desk, and let callers choose
microphone and speaker devices for voice.

## Problem

After S54, callers pick a brand and enter the desk, but brand identity is small
and voice always uses the browser default audio devices. Mockups show a large
Monti hero, large company logos on directory cards, a selected-company card on
the desk, and an expectation of controllable audio I/O.

## Scope

### In

- Call desk visual revamp: large Monti product mark + selected company logo card
- Tenant list visual revamp: large Monti hero + large brand logos on cards
- Mic and speaker device selection (browser media device APIs)
- Apply selected devices to existing Gemini voice / LiveKit capture & playback
- Session and/or localStorage preference for last device ids
- Fallbacks when logo missing or permission denied (non-technical copy)
- Keep S54 routes (`/`, `/t/{slug}`) and Live · OK system status

### Out

- Tenant theme publish CMS changes beyond consuming public theme/logo
- Native OS-only audio routing outside browser APIs
- Full OTP flow redesign
- S55 topic statistics
- Device test-tone / noise-suppression product

## Acceptance criteria

1. Tenant list shows a large Monti mark and each company card shows a large
   brand logo (or monogram fallback).
2. Call desk shows a large Monti product logo and a clear selected-company logo
   block (not only a tiny header chip).
3. Caller can open audio settings and choose **microphone** and **speaker**
   when multiple devices exist.
4. Voice start uses the selected mic/speaker (or safe default).
5. Switching brand updates company logo to the newly selected tenant only.
6. Permission denied / no devices shows clear non-technical messaging.

## Test notes

- Two listed brands with different `logo_url` values
- Missing logo → monogram fallback
- Device switch mid-session if supported by browser; otherwise on next call
- Denied mic permission path
- Thai + English labels for audio settings chrome

## Dependencies

- packages: `apps/customer-web` (TenantPicker, CustomerDesk, GeminiVoice),
  public theme/brand APIs, S54 routing
- mockups: [call-page](../02-design/mockups/s56-caller-desk-branding/call-page.png),
  [tenant-list](../02-design/mockups/s56-caller-desk-branding/tenant-list.png)
- design: [DES-0052](../02-design/52-caller-desk-branding-audio-devices-spec.md)
