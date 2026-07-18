# Feature: Embed Framework SDKs   (FEAT-0017)
**Sprint:** SPRINT-037 (roadmap #37)   **Owner:** DEV   **Status:** shipped · **Release:** v2.14.0

## Problem

SPRINT-014 shipped a **vanilla** loader (`monti-embed.js`) and iframe embed surface. Integrators on modern stacks need **first-class packages** for:

| Target | Typical use |
| --- | --- |
| **Vue 3** | Composition API component + plugin |
| **React** | Component + hook(s) |
| **Svelte** | Svelte 4/5 component (align with Monti portal stack) |
| **Web Component** | Framework-agnostic custom element (`<monti-embed>`) for plain HTML, Angular, etc. |

Without SDKs, each host re-implements script injection, lifecycle, and props; type safety and SSR edge cases stay undocumented.

## Scope

**In:**
- Monorepo or `packages/` publishables (npm-ready; private or public registry TBD):
  - `@monti/embed-core` — shared resolve, origin, open/close, postMessage bridge (thin wrapper over existing public embed APIs + iframe)
  - `@monti/embed-vue` — Vue 3 component / plugin
  - `@monti/embed-react` — React component + hooks
  - `@monti/embed-svelte` — Svelte component
  - `@monti/embed-web-component` — Custom element wrapping the same core
- Props aligned with S14: `embedKey`, `apiBase`, `parentOrigin`, optional agent, theme, open/close events
- TypeScript types + README examples per framework
- Update [EMBED_WEB_INTEGRATION.md](../../EMBED_WEB_INTEGRATION.md) with framework install snippets
- Tenant admin `/tenant/embed` optional “Framework” tab with copy-paste examples (Vue/React/Svelte/WC)
- Smoke demos under `poc/` (or `examples/`) for each framework
- Keep **vanilla `monti-embed.js`** as the zero-dependency path (no breaking change)

**Out:**
- Native mobile (iOS/Android) SDKs
- Changing public embed resolve security model (reuse S14 allowlist / keys)
- White-label custom domain CNAME
- Billing per embed view
- Full Storybook / design-system productization

## Acceptance criteria

1. Install + mount docs work for **Vue**, **React**, **Svelte**, and **Web Component** against a local Monti server and valid `embed_key`.
2. Each package opens the same embed surface (chat + voice when configured) with origin allowlist still enforced.
3. Packages expose open/close (or equivalent) and destroy/unmount without leaking listeners or orphan iframes.
4. TypeScript consumers get prop types; bad `embedKey` surfaces a clear error state.
5. Integrator guide documents all four plus the existing script tag path.
6. Tenant embed admin can copy at least one snippet per framework (or links to guide sections).

## Test notes

- Unit/smoke: core open/close with mock iframe
- Manual: each POC page loads agent list and one chat turn
- Regression: existing `monti-embed.js` fixture still passes S14 checklist

## Dependencies

- **SPRINT-014** Embed to Web (FEAT-0014) — public resolve, loader, `/embed` surface
- Optional polish after **SPRINT-016** locale (pass locale prop into embed)
- packages: new `packages/embed-*`, docs, tenant embed UI snippets

## Notes

- Place after Phase D go-live core (S14–18) as **Sprint 36** so commerce/ops sprints stay numbered; pull forward if integrator demand is high.
- Web Component is the interop layer for Angular and non-SPA hosts.
