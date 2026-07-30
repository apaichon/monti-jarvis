---
id: DES-0054
title: Portal UI Language Selector Specification
status: shipped
release: v2.31.0
updated: 2026-07-30
sprint: SPRINT-058
owner: SA
feature: FEAT-0050
release_target: v2.31.0
---

# DES-0054 — Portal UI Language Selector (EN · TH · JA)

**Sprint:** SPRINT-058 · **Release target:** v2.31.0  
**Feature:** [FEAT-0050](../01-features/FEAT-0050-portal-ui-language-selector.md)  
**Tasks:** [TASK-0208](../04-tasks/TASK-0208.md), [TASK-0209](../04-tasks/TASK-0209.md),
[TASK-0210](../04-tasks/TASK-0210.md), [TASK-0211](../04-tasks/TASK-0211.md)  
**Reference implementation:** `apps/product-web/src/lib/i18n/` (already ships EN/TH/JA)  
**Related:** [DES-0019](19-tenant-settings-limits-spec.md) (AI/workspace locale — **not** UI lang)

## 1. Goals

1. Provide a consistent **display language** control: English, Thai, Japanese.
2. Localize **portal chrome** on customer call surfaces, tenant console, and
   platform admin.
3. Persist preference in the browser; fall back safely when a key is missing.
4. Keep UI language **orthogonal** to S16 AI reply locale and tenant workspace
   locale.

## 2. Non-goals

- Machine translation of user/tenant content
- Server-side profile language API (future)
- product-web marketing rewrite (already i18n’d)
- RTL scripts
- Binding voice STT/TTS model language to the selector

## 3. Supported languages

| Code | Endonym label in selector | BCP 47 `html lang` |
| --- | --- | --- |
| `en` | English | `en` |
| `th` | ไทย | `th` |
| `ja` | 日本語 | `ja` |

Default resolution order:

1. `?lang=` query if valid  
2. `localStorage['monti_jarvis:ui_lang']`  
3. Optional browser hint (`navigator.language` starts with `th` / `ja`)  
4. `en`

## 4. Architecture

### 4.1 Per-app modules (v1)

| App | Path | Notes |
| --- | --- | --- |
| customer-web | `src/lib/i18n/{index,messages}.ts` | Directory + desk |
| tenant-web | `src/lib/i18n/{index,messages}.ts` | Shell + pages |
| platform-admin-web | `src/lib/i18n/{index,messages}.ts` | Shell + pages |
| product-web | existing `src/lib/i18n/` | Reference; optional key align later |

Shared npm package is **optional**; copy-aligned API is fine for this sprint.

### 4.2 Runtime API (align with product-web)

```ts
export type Lang = 'en' | 'th' | 'ja';
export const lang: Writable<Lang>;
export const t: Readable<Messages>; // or derived
export function setLang(next: Lang): void;
export function initLangFromUrl(searchParams: URLSearchParams): void;
export function msg(): Messages; // imperative snapshot
export function supportedLangs(): Lang[];
```

**Fallback:** when resolving key `k` for lang `L`:

```
messages[L][k] ?? messages.en[k] ?? k
```

Never render empty string for required chrome keys.

### 4.3 Storage

| Store | Key | Value |
| --- | --- | --- |
| localStorage | `monti_jarvis:ui_lang` | `en` \| `th` \| `ja` |

On `setLang`: write storage + set `document.documentElement.lang` and
`dataset.lang`.

product-web today uses `monti_product_lang`. Out of scope to force-merge this
sprint; document dual keys. Future: single key.

### 4.4 LanguageSelector UX

- Compact control: segmented buttons or `<select>` with flags/codes
- Placement:
  - Customer: header/hero chrome on list + desk rail
  - Tenant: sidebar foot (above account) or top bar
  - Admin: sidebar foot (above account)
- ARIA: `aria-label` = localized “Language” / “ภาษา” / “言語”
- Changing language is synchronous client-side (no API)

## 5. Separation from S16 locale

| Concern | Field / store | Owner | Affects |
| --- | --- | --- | --- |
| UI display language | `monti_jarvis:ui_lang` | Browser | Labels only |
| Tenant workspace locale | `tenant_settings.locale` | S16 API | Defaults / display prefs |
| AI reply language hint | `tenant_settings.ai_reply_locale` | S16 API | System prompt hint |
| Customer account locale | customer profile | S19+ | Future personalization |

**Settings copy (tenant):** show both controls with explicit names so operators
do not conflate “Display language” with “AI reply language”.

## 6. Data model

**No Postgres / ClickHouse / Redis migration.**

Optional future (out of sprint):

| Entity | Field | Notes |
| --- | --- | --- |
| `users` / customer profile | `ui_lang` | Server-persisted preference |

## 7. API summary

**No new REST endpoints required** for MVP.

Reuse existing public/tenant/admin APIs unchanged. Content strings from APIs
(brand name, package name, ticket subject) stay as returned.

Optional later:

| Method | Path | Use |
| --- | --- | --- |
| `PATCH` | `/api/me/preferences` | Persist `ui_lang` for authenticated users |

## 8. Catalog inventory (minimum keys)

### 8.1 Shared chrome (all apps)

| Key | EN example |
| --- | --- |
| `lang_label` | Language |
| `lang_en` | English |
| `lang_th` | ไทย |
| `lang_ja` | 日本語 |
| `action_save` | Save |
| `action_cancel` | Cancel |
| `action_delete` | Delete |
| `action_create` | Create |
| `action_search` | Search |
| `action_logout` | Sign out |
| `status_loading` | Loading… |
| `status_empty` | Nothing here yet |
| `status_error` | Something went wrong |
| `status_live` | Live |
| `status_ok` | OK |

### 8.2 Customer (directory + desk)

| Key group | Examples |
| --- | --- |
| Directory | choose brand, search brands, start call CTA |
| Desk | topics, start/end call, send, placeholder, audio settings, mic, speaker |
| Account | sign in, OTP, account, footer secure note |
| System | Live · OK non-technical status |

### 8.3 Tenant nav

| Key | EN |
| --- | --- |
| `nav_group_operations` | Operations |
| `nav_overview` | Overview |
| `nav_call_center` | Call center |
| `nav_monitoring` | Monitoring |
| `nav_tickets` | Tickets |
| `nav_satisfaction` | Satisfaction |
| `nav_preview` | Preview |
| `nav_group_knowledge` | Knowledge |
| `nav_knowledge` | Knowledge |
| `nav_gaps` | Gaps |
| `nav_records` | Records |
| `nav_group_commerce` | Commerce |
| `nav_billing` | Billing |
| `nav_documents` | Documents |
| `nav_tax` | Tax |
| `nav_group_channels` | Channels |
| `nav_avatars` | Avatars |
| `nav_embed` | Embed |
| `nav_theme` | Theme |
| `nav_ai_config` | AI config |
| `nav_group_directory` | Directory |
| `nav_customers` | Customers |
| `nav_tiers` | Tiers |
| `nav_group_growth` | Growth |
| `nav_referrals` | Referrals |
| `nav_group_settings` | Settings |
| `nav_settings` | Settings |
| `plan_current` | Current plan |
| `settings_display_language` | Display language |
| `settings_ai_reply_language` | AI reply language |

### 8.4 Platform admin nav

| Key | EN |
| --- | --- |
| `nav_overview` | Overview |
| `nav_packages` | Packages |
| `nav_tenants` | Tenants |
| `nav_avatars` | Avatars |
| `nav_billing` | Billing |
| `nav_quotes` | Quote requests |
| `nav_leads` | Leads |
| `nav_audit` | Audit log |
| `nav_monitoring` | Monitoring |
| `nav_call_center` | Call center |
| `nav_payment` | Payment |
| `nav_profile` | Profile |
| `system_health` | System health |
| `system_all_ok` | All systems operational |

## 9. Component → file map

| Component / module | App | Responsibility |
| --- | --- | --- |
| `lib/i18n/index.ts` | all three | store, setLang, fallback |
| `lib/i18n/messages.ts` | all three | EN/TH/JA catalogs |
| `LanguageSelector.svelte` | all three | selector UI |
| `TenantPicker.svelte` | customer | directory strings |
| `CustomerDesk.svelte` | customer | desk chrome + audio labels |
| `routes/+layout.svelte` | tenant | nav + foot |
| `routes/+layout.svelte` | admin | nav + foot |
| Settings pages | tenant | dual locale explanation |

## 10. Env vars

None required. Optional future: `PUBLIC_DEFAULT_UI_LANG=en`.

## 11. RBAC

Unaffected. Language is client-only; no authz change.

## 12. Verification

```bash
cd apps/customer-web && npm run check
cd apps/tenant-web && npm run check
cd apps/platform-admin-web && npm run check
```

Manual:

1. Open customer `/` → switch TH → labels Thai → reload still TH  
2. Open `/t/{slug}` → JA → desk chrome Japanese; brand name unchanged  
3. Tenant login → switch languages; nav groups update  
4. Settings: change UI lang; leave AI reply locale alone; confirm both independent  
5. Admin shell → EN/TH/JA nav  
6. Force-remove a JA key → EN string appears for that control  

## 13. Status gates

| Gate | Status |
| --- | --- |
| Deep spec | **approved** (implementation may start) |
| API | no new endpoints — approved as N/A for REST |
| UX wireframes | C58 / T58 / A58 in `05-ux-ui.md` |

## See also

- Workflow §130–131  
- ER Sprint 58 (client keys only)  
- API Sprint 58  
- UX C58 / T58 / A58  
- product-web i18n  
