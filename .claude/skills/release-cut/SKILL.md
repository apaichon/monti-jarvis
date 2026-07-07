---
name: release-cut
description: Cut a release for Monti Jarvis — compute next semver, update VERSION, summarize shipped sprint work, and suggest the git tag. Use at sprint close when shippable work is verified. (PM + DevOps)
---

# release-cut — record a shipped version

## Procedure
1. Confirm [`docs/sdlc/08-readiness/RELEASE-READINESS.md`](../../docs/sdlc/08-readiness/RELEASE-READINESS.md) sections A–G are green; manual UAT signed off in [`06-manual-tests/`](../../docs/sdlc/06-manual-tests/).
2. Run **`km-context`** for the closing sprint; collect completed tasks/features.
3. Compute next version from `VERSION` file (semver):
   - **patch** — fixes only
   - **minor** — new backward-compatible features (typical sprint close)
   - **major** — breaking API/schema change
4. Update `VERSION` (e.g. `0.2.0` → `0.3.0`).
5. Ensure sprint doc has `release: vX.Y.Z` and shipped summary.
6. Run **`km-sync`** to close the sprint and update ROADMAP.
7. Suggest tag (user must approve push):
   ```bash
   git tag -a vX.Y.Z -m "vX.Y.Z — <sprint title>"
   ```

## Guardrails
- Only ship verified work. Unverified tasks stay open or move to next sprint explicitly.
- Tagging/pushing requires explicit user OK.

See `VERSION`, `docs/sdlc/03-sprints/SPRINT-001.md` (v0.2.0 example).