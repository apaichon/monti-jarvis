---
id: FEAT-0050
title: "Portal UI language selector: EN, TH, and Japanese labels"
status: shipped
release: v2.31.0
roadmap_sprint: 58
priority: I18N
depends_on: [SPRINT-016, SPRINT-020, SPRINT-039, SPRINT-054, SPRINT-056]
design: DES-0054
design_spec: ../02-design/54-portal-ui-language-selector-spec.md
updated: 2026-07-30
---

# FEAT-0050: Portal UI Language Selector (EN · TH · JA)

## Purpose

Let callers, tenant operators, and platform admins switch portal **UI chrome
labels** between English (`en`), Thai (`th`), and Japanese (`ja`) without
changing AI reply language, KM content, or tenant brand copy.

## Problem

Sprint 16 stores tenant `locale` / `ai_reply_locale` for AI behavior, and
product-web already ships EN/TH/JA marketing i18n. Customer call desk, tenant
console, and platform admin remain English-first with ad-hoc bilingual strings.
Operators and callers in TH/JA markets need a consistent language selector and
complete primary chrome labels.

## Scope

### In

- Display-language selector (EN / TH / JA) on:
  - Customer portal: brand directory + call desk
  - Tenant portal shell (nav, common chrome, primary pages)
  - Platform admin shell (nav, common chrome, primary pages)
- Per-app message catalogs (mirror product-web `lib/i18n` pattern)
- Persist selection in browser `localStorage` (shared key namespace)
- Missing key fallback: selected lang → `en` → raw key never blank for controls
- Keep S16 AI reply locale and tenant settings locale **separate** from UI lang

### Out

- Machine translation / auto-detect beyond optional browser `navigator.language` default
- Translating KM documents, transcripts, invoices/legal, payment provider UIs
- Server-side user preference API (optional later when profile fields exist)
- Product-web marketing site (already i18n’d — reference only)
- RTL languages
- Changing voice model spoken language as a product of the selector

## Acceptance criteria

1. Customer call page, tenant portal, and platform admin each expose EN, TH, and
   Japanese display-language selection.
2. Selecting a language updates visible UI labels without logout or full route
   reset.
3. Reloading the page preserves the selected display language for that browser.
4. Missing TH/JA strings fall back to EN and never leave primary controls blank.
5. Tenant settings **AI reply locale** / workspace locale from S16 remain
   independent of the UI language selector.
6. Primary navigation, primary actions, form labels, statuses, and common
   empty/error states are localized for the three surfaces in scope.

## Test notes

- Switch EN → TH → JA on each surface; confirm labels and `document.documentElement.lang`
- Hard reload preserves language
- Delete one JA key in catalog → falls back to EN for that key only
- Tenant AI reply locale stays `th` while UI is `ja` (no accidental coupling)
- Compare pattern with `apps/product-web/src/lib/i18n/`

## Dependencies

- packages: `apps/customer-web`, `apps/tenant-web`, `apps/platform-admin-web`
- reference: `apps/product-web/src/lib/i18n/` (EN/TH/JA runtime already shipped)
- prior: S16 settings locale, S54 tenant list, S56 caller desk chrome
- design: [DES-0054](../02-design/54-portal-ui-language-selector-spec.md)

## Soft note on S57

Roadmap lists a soft depend on S57 (logo/social preview). **UI i18n does not
block on S57** — implement against current branding assets.
