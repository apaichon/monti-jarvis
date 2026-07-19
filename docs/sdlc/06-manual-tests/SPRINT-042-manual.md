---
id: MANUAL-SPRINT-042
sprint: SPRINT-042
release_target: v2.16.0
feature: FEAT-0036
updated: 2026-07-19
status: ready
---

# SPRINT-042 Manual UAT — Tenant UX Bug Fix

**Feature:** [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md) · **Sprint:** [SPRINT-042](../03-sprints/SPRINT-042.md)  
**Tasks:** TASK-0154–0157  

Mark **[x]** pass / **[ ]** fail.

## 0. Preconditions

| Check | Done |
| --- | --- |
| S42 worktree/code on :8091 | [ ] |
| Active tenant login works | [ ] |
| Browser DevTools available | [ ] |

```bash
cd .worktrees/SPRINT-042
make restart   # or rebuild tenant-web + server
```

## 1. Session expired (TASK-0154)

1. Login as tenant admin.  
2. Open Application → Local Storage → delete `monti_tenant_access` (keep refresh to test refresh path optionally delete both).  
3. Navigate to Knowledge or click a nav link that loads API data.  

| Step | Expected | Result |
| --- | --- | --- |
| 1.1 With access deleted, refresh may restore session | Still in app if refresh works | [ ] |
| 1.2 Delete access + refresh tokens | Redirect to `/tenant/login?reason=session_expired&next=…` | [ ] |
| 1.3 Banner on login | “session expired” message | [ ] |
| 1.4 Sign in again | Lands on `next` path (e.g. /tenant/km) | [ ] |

## 2. First-login menu (TASK-0155)

1. Clear site data for localhost / tenant host.  
2. Open `/tenant/login` → sign in.  

| Step | Expected | Result |
| --- | --- | --- |
| 2.1 After login | Full sidebar shell with **grouped** nav visible **without** F5 | [ ] |
| 2.2 Refresh while logged in | Menu still present | [ ] |

## 3. Nav groups + scroll (TASK-0156)

1. Desktop width with short viewport height (~700px).  
2. Inspect sidebar groups.  

| Step | Expected | Result |
| --- | --- | --- |
| 3.1 Groups visible | Operations, Knowledge, Commerce, Channels, Directory, Settings | [ ] |
| 3.2 Scroll | Can scroll nav to Theme / Settings without clicking last item first | [ ] |
| 3.3 Active highlight | Current route still highlighted | [ ] |

## 4. Document scope (TASK-0157)

1. Open `/tenant/km`.  
2. Upload a doc with scope **billing**.  
3. Filter scope = billing / general.  
4. Change document scope via dropdown.  

| Step | Expected | Result |
| --- | --- | --- |
| 4.1 Upload with scope | Doc listed with selected scope | [ ] |
| 4.2 Filter billing | Only billing docs | [ ] |
| 4.3 Filter all | All docs | [ ] |
| 4.4 Patch scope | Scope updates and counts refresh | [ ] |

## 5. Sign-off

| Field | Value |
| --- | --- |
| Tester | |
| Date | |
| Result | PASS / FAIL |
