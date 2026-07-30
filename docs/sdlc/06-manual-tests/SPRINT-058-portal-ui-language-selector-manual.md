# SPRINT-058 — Portal UI language selector manual UAT

**Feature:** FEAT-0050 · **Design:** DES-0054 · **Tasks:** TASK-0208–0211

## Preconditions

- Local server running (customer `/`, tenant `/tenant`, admin `/admin`)
- Tenant login and platform admin login available

## Cases

### C1 — Customer directory language switch

| Step | Action | Expected | Pass |
| --- | --- | --- | --- |
| 1 | Open customer `/` | Language selector EN/TH/JA visible | ☐ |
| 2 | Select **TH** | Title/search/call CTA in Thai | ☐ |
| 3 | Reload | Language remains TH | ☐ |
| 4 | Select **JA** | Labels Japanese; brand names unchanged | ☐ |
| 5 | `?lang=en` | Switches to English | ☐ |

### C2 — Customer call desk

| Step | Action | Expected | Pass |
| --- | --- | --- | --- |
| 1 | Open `/t/{slug}` with UI lang JA | Desk chrome Japanese | ☐ |
| 2 | Switch EN → TH | Audio settings / topics / send update | ☐ |
| 3 | Start call flow labels | End/start call localized | ☐ |
| 4 | Brand name | Still server-provided name | ☐ |

### C3 — Tenant portal

| Step | Action | Expected | Pass |
| --- | --- | --- | --- |
| 1 | Login tenant shell | Selector in sidebar foot | ☐ |
| 2 | Switch JA | Nav groups + links Japanese | ☐ |
| 3 | Settings | Display language section present | ☐ |
| 4 | Change UI to JA; leave AI reply TH | AI setting still TH after save | ☐ |
| 5 | Reload | UI stays JA | ☐ |

### C4 — Platform admin

| Step | Action | Expected | Pass |
| --- | --- | --- | --- |
| 1 | Login admin | Selector in sidebar foot | ☐ |
| 2 | Switch TH | Nav + system health Thai | ☐ |
| 3 | Reload | Preference persists | ☐ |

### C5 — Fallback

| Step | Action | Expected | Pass |
| --- | --- | --- | --- |
| 1 | Force missing JA key (dev) | EN string for that key only | ☐ |
| 2 | Clear `localStorage monti_jarvis:ui_lang` | Defaults EN or browser hint | ☐ |

## Commands

```bash
cd apps/customer-web && npm run check
cd apps/tenant-web && npm run check
cd apps/platform-admin-web && npm run check
```

## Sign-off

| Role | Name | Date | Result |
| --- | --- | --- | --- |
| Tester | | | ☐ Pass / ☐ Fail |
