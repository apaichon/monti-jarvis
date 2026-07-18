---
id: SPRINT-037
status: completed
start: 2026-07-18
end: 2026-07-18
updated: 2026-07-18
design_pack: existing
release_target: v2.14.0
release: v2.14.0
closed: 2026-07-18
goal: "Ship first-class embed framework SDKs (Vue · React · Svelte · Web Component) on shared @monti/embed-core."
roadmap_sprint: 37
feature: FEAT-0017
platform: Tenant / Integrator
depends_on: [SPRINT-014]
---

# SPRINT-037 — Embed Framework SDKs (Vue · React · Svelte · Web Component)

## Goal

Give host integrators npm-ready packages for Vue 3, React, Svelte, and a framework-agnostic Web Component, all wrapping the S14 public resolve + `/embed` iframe surface, without breaking the zero-dependency `monti-embed.js` loader.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed (S30, S31, S32) | 16, 16, 3 → **avg ~11.7** |
| **Committed / completed** | **14** |

## Commitment

| Work package | Points | Owner | Status | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0146](../04-tasks/TASK-0146.md) `@monti/embed-core` + lifecycle tests | 5 | dev | completed | Resolve, iframe URL, floating/inline widget, open/close/destroy, clear error states |
| [TASK-0147](../04-tasks/TASK-0147.md) Framework packages Vue/React/Svelte/WC | 6 | dev | completed | Four publishable wrappers over core with typed props/events |
| [TASK-0148](../04-tasks/TASK-0148.md) Docs, examples, tenant Framework tab | 3 | dev | completed | Integrator guide § Framework SDKs, smoke demos, `/tenant/embed` copy snippets |

**Committed:** 14 points · **Completed:** 14 points · **Task IDs:** TASK-0146–TASK-0148.

## Scope boundary

**In**

- `packages/embed-core`, `embed-vue`, `embed-react`, `embed-svelte`, `embed-web-component`
- Props aligned with S14: `embedKey`, `apiBase`, `parentOrigin`, position, optional agent/theme/locale
- TypeScript types, package READMEs, `examples/embed-sdks`
- Update `docs/EMBED_WEB_INTEGRATION.md`
- Tenant admin Framework SDKs tab with copy-paste snippets
- Keep vanilla `monti-embed.js` path

**Out**

- Native mobile SDKs (covered by SPRINT-027)
- Changing public embed resolve security model
- White-label CNAME, billing per embed view, Storybook productization
- npm registry publish automation (packages are monorepo-ready; registry TBD)

## Design

| Artifact | Status |
| --- | --- |
| Feature [FEAT-0017](../01-features/FEAT-0017-embed-framework-sdks.md) | shipped |
| Embed to web [DES-0017](../02-design/17-embed-to-web-spec.md) | reused (S14) |
| Integrator guide [EMBED_WEB_INTEGRATION.md](../../EMBED_WEB_INTEGRATION.md) | updated |

## Verification

```bash
./packages/build-embed-sdks.sh
# core unit tests: open/close/destroy + embed_not_found
# optional: serve examples/embed-sdks against local Monti + valid emb_ key
```

Regression: existing `monti-embed.js` / S14 checklist unchanged.

## Release

**v2.14.0** — Embed Framework SDKs (FEAT-0017 / SPRINT-037).
