---
id: FEAT-0040
title: "Product web for marketing, lead capture, demos, tenant registration, and package conversion"
status: in_progress
roadmap_sprint: 48
priority: O
depends_on: [SPRINT-004, SPRINT-006, SPRINT-009, SPRINT-017, SPRINT-020, SPRINT-031, SPRINT-039, SPRINT-046]
updated: 2026-07-25
---

# FEAT-0040: Product Web Growth and Tenant Conversion

## Purpose

Ship a public Monti product website that turns advertising, SEO, and referral
traffic into qualified leads, live demos, tenant registrations, and paid
package customers — without replacing existing registration, KYC, checkout, or
billing authorities.

## Problem

Monti has tenant registration, package catalog, checkout, demo/preview, and
referral attribution, but no cohesive public marketing surface that:

- explains product value in Thai/English-ready structure
- captures consent-aware leads for sales follow-up
- preserves campaign/referral attribution through conversion
- routes safely into live demo, register, and authenticated purchase

## Scope

### In

- Responsive Svelte product-web shell (`apps/product-web`) with Monti dark brand
- Marketing pages: Home, Product, Solutions, Resources, Pricing, About, Contact, Demo entry
- Lead capture (contact / book demo / newsletter) with consent, rate limit, dedupe
- Safe attribution (`utm_*`, referral code, landing path) through CTA redirects
- Public catalog-driven pricing (read-only; no entitlement grant)
- Conversion links into existing demo, tenant register, and authenticated checkout
- Platform sales lead list with lifecycle status and notes
- Funnel measurement events (visit, CTA, demo, lead, register, purchase) without PII abuse

### Out

- Full external CRM replacement
- Unapproved advertising claims or hard-coded package prices
- Public payment handling or entitlement creation from pricing pages
- Automatic marketing email blasts without consent
- Arbitrary third-party redirects
- Replacing S46 referral qualification rules
- S44 generative workspace, Langfuse (S47), or central multi-brand hub (S38)

## Acceptance criteria

1. A visitor can open product-web primary pages, understand Monti value, view
   catalog-driven package options, and reach live demo, contact, or registration
   with one clear CTA per primary page.
2. A visitor can start the approved no-auth live demo without tenant signup; a
   later contact/registration can retain originating campaign/referral context
   when consent and attribution policy allow it.
3. Book-demo and contact forms validate, rate-limit, record consent/source,
   show confirmation, and create exactly one deduplicated lead for safe retries.
4. Register/pricing CTAs redirect only to approved Monti routes, preserve safe
   context query params, and never grant package entitlement from public content.
5. Authenticated package purchase after registration still uses existing checkout
   and billing authority; success surfaces receipt/tax state and tenant workspace.
6. Platform sales roles can list leads, update lifecycle status, and add notes
   without seeing other tenants’ private customer conversations, credentials, or
   payment secrets.
7. Funnel analytics report visits, CTA clicks, demo starts, leads, registrations,
   and purchase conversions by campaign without exposing raw customer content.

## Test notes

- Functional: form validation, rate limit, consent, dedupe, allowlisted redirects
- Isolation: two concurrent lead submissions and two-tenant sales views
- Conversion smoke: product web → register → verify/KYC → package buy (existing chain)
- Languages: Thai + English labels/structure ready (content may ship EN-first)
- Responsive/a11y: keyboard nav, focus, contrast on dark brand chrome

## Dependencies

- packages: `apps/product-web`, `internal/leads` (new), `internal/productweb`,
  `internal/store`, `cmd/server`, existing tenant register/checkout/packages
- design: [DES-0043](../02-design/43-product-web-growth-spec.md)
- blueprint: `docs/monti_multi_tenant_ai_call_center_blueprint.md`
- reference art: `/Users/apaichon/Downloads/monti/product-web/` (inspiration only)
