# Monti Multi-tenant AI Call Center as a Service Blueprint

**Version:** 2.0  
**Date:** 2026-07-07  
**Product Direction:** Multi-tenant AI Call Center as a Service  
**Primary Start Point:** Inbound AI calls for high-value businesses  
**Official Roadmap:** 35 sprints — see [Section 17](#17-official-35-sprint-roadmap) and `docs/sdlc/00-roadmap/ROADMAP.md`

---

## 1. Executive Summary

Monti AI Call Center is a multi-tenant SaaS platform that allows businesses to launch AI employees for customer service, sales, booking, support, and lead qualification. Each tenant can configure brand profiles, languages, AI employees, knowledge bases, actions, ticketing workflows, and review processes.

The first phase should focus on **inbound AI call center use cases** for high-value businesses. The core customer flow is:

```text
Customer
  ↓
Search Brand
  ↓
Select Language
  ↓
Select AI Employee
  ↓
Start Call
  ↓
AI answers using tenant knowledge
  ↓
Open Ticket / Lead / Case
  ↓
Perform Action
  ↓
Generate Call Summary
  ↓
Tenant Review
  ↓
Improve KM / Script / Workflow
```

The main platform touchpoints are:

1. **Customer Portal**
2. **Tenant Admin Portal**
3. **Platform Admin Portal**

The recommended product strategy is:

```text
Svelte + shadcn-svelte for all web portals (Customer, Tenant, Platform Admin)
Go + Fiber for API orchestration and business logic
LiveKit for realtime voice rooms (WebRTC, SIP later)
NATS.io + JetStream for event-driven workflows
Postgres as system of record (tenants, billing, tickets, metadata)
Redis 8 for session, cache, rate limits, and active call state
MinIO for recordings, KM files, avatars, exports
ClickHouse for analytics AND vector RAG (embeddings, semantic search)
Flutter later as mobile scale phase (WebView wrapper first)
```

---

## 2. Product Positioning

### 2.1 Product Name

```text
Monti AI Call Center
```

### 2.2 Core Concept

```text
AI Employees for Every Business
```

### 2.3 Simple Pitch

Monti is a multi-tenant AI Call Center platform that lets businesses create AI employees for inbound customer calls, connect their knowledge base, create tickets automatically, summarize every call, and improve service quality through review and analytics.

### 2.4 Value Proposition

```text
Let every business create AI employees that answer calls, open tickets, take actions, and improve from real customer conversations.
```

### 2.5 Why Start with Inbound Calls

Inbound calls are the best first use case because they create clear business value:

- Reduce call center workload
- Answer customers 24/7
- Capture missed opportunities
- Standardize service quality
- Convert calls into tickets and leads
- Generate searchable call summaries
- Improve the knowledge base from real customer questions
- Support multilingual customer service

---

## 3. High-value Business Prioritization

### 3.1 Best Early Target Businesses

| Priority | Business Type | Why It Is High Value |
|---:|---|---|
| 1 | Clinics / Hospitals / Wellness | Booking, appointment changes, FAQ, high call volume |
| 2 | Real Estate / Property Sales | Lead qualification, project info, appointment booking |
| 3 | Insurance / Finance Brokers | Product FAQ, lead screening, document follow-up |
| 4 | Hotels / Travel / Tourism | Multilingual calls, booking, directions, service questions |
| 5 | Education / Course Providers | Course FAQ, enrollment, schedule, payment follow-up |
| 6 | E-commerce / Retail | Order status, product questions, complaints, returns |
| 7 | Restaurants / Franchises | Reservation, branch info, promotion, delivery support |

### 3.2 Recommended First MVP Templates

Start with three vertical templates:

1. **Clinic AI Receptionist**
2. **Real Estate AI Sales Assistant**
3. **Education Course AI Advisor**

These are strong starting points because they have repetitive questions, clear workflows, measurable lead/booking value, and manageable knowledge structures.

---

## 4. Main Value Flow

```text
Customer
  ↓
Search Brand
  ↓
Select Language
  ↓
Select AI Employee
  ↓
Start Call
  ↓
AI answers using Tenant KM
  ↓
Open Ticket / Lead / Case
  ↓
Perform Action
  ↓
Call Summary
  ↓
Tenant Review
  ↓
Improve KM / Script / Workflow
```

### 4.1 Customer Journey

#### Step 1: Customer Searches Brand

Customer may come from:

- Web landing page
- QR code
- Google search
- Facebook page
- LINE OA
- Tenant website
- Embedded call widget
- Mobile app
- Phone number
- Kiosk

Example:

```text
Customer searches: ABC Dental Clinic
→ Brand profile appears
→ Customer selects preferred language
→ Customer chooses AI Receptionist
→ Customer starts voice call
```

#### Step 2: Select Language

Supported languages can be configured per tenant and per AI employee.

Example language options:

```text
Thai
English
Chinese
Japanese
Korean
Custom tenant language
```

Example employee language capability:

```text
AI Receptionist
- Thai: enabled
- English: enabled
- Japanese: disabled
```

#### Step 3: Select AI Employee

Customers choose the AI role they want to talk to.

Example for a clinic:

```text
1. AI Receptionist
2. AI Appointment Assistant
3. AI Product Advisor
4. AI Complaint Handler
5. Human Support Request
```

Example for real estate:

```text
1. AI Sales Assistant
2. AI Project Consultant
3. AI Loan Pre-check Assistant
4. AI Appointment Booking Assistant
```

#### Step 4: Start Call

Supported call channels:

```text
Web voice call
Mobile app voice call
LINE call integration
Phone call through SIP/Twilio/telephony provider
Embedded widget
Kiosk
```

The system should identify:

```text
Who is calling?
Which tenant owns this brand?
What language is selected?
Which AI employee was selected?
What does the customer need?
Is the request FAQ, lead, complaint, booking, payment, support, or escalation?
```

#### Step 5: Open Ticket

Every meaningful conversation should create a structured record.

Ticket types:

```text
Lead
Booking
Complaint
Support case
Follow-up request
Payment issue
Document request
General inquiry
Escalation to human
```

Example ticket:

```json
{
  "tenant": "ABC Dental Clinic",
  "customer_name": "Somchai",
  "language": "Thai",
  "intent": "appointment_booking",
  "priority": "high",
  "summary": "Customer wants to book teeth whitening on Sunday afternoon.",
  "next_action": "Confirm available slots",
  "status": "open"
}
```

#### Step 6: Action

The AI should not only answer. It should perform approved business actions.

| Action | Example |
|---|---|
| Create ticket | Complaint, lead, booking request |
| Check availability | Doctor schedule, course schedule, room availability |
| Book appointment | Clinic, hotel, sales visit |
| Send message | Send LINE, email, or SMS confirmation |
| Update CRM | Save lead status |
| Escalate to human | Sensitive issue, complaint, payment problem |
| Collect information | Name, phone, preferred time, issue detail |
| Generate summary | Create staff review summary |

#### Step 7: Summary Call

After every call, generate:

```text
Call summary
Customer intent
Important details
Sentiment
Next action
Responsible team
Ticket status
AI confidence score
Knowledge used
Missing knowledge
Escalation reason, if any
```

#### Step 8: Review

Tenant admin can review:

```text
Call recording
Transcript
AI summary
Ticket
Customer sentiment
AI correctness
Knowledge sources used
Human feedback
Suggested KM improvement
```

Platform admin can review across tenants:

```text
Usage
Cost
Quality
Model performance
Provider performance
Tenant health
SLA
Abuse and safety issues
```

---

## 5. Platform Touchpoints

## 5.1 Customer Portal

For end customers.

Main functions:

```text
Search brand
View brand profile
Select language
Select AI employee
Start voice call
View call status
Submit contact info
Receive ticket reference
Rate experience
Request human callback
```

Simple customer screen:

```text
[Search brand...]

ABC Dental Clinic
Language: [Thai v]
AI Employee:
[Receptionist] [Booking Assistant] [Product Advisor]

[Start Call]
```

## 5.2 Tenant Admin Portal

For each business customer.

Main functions:

```text
Manage brand profile
Manage AI employees
Upload knowledge
Configure languages
Configure call scripts
Configure business actions
Review tickets
Review call summaries
Review transcripts
Improve KM
Monitor usage and cost
Connect CRM / LINE / calendar / ticketing
```

Tenant admin dashboard:

```text
Tenant Admin Dashboard

Today
- Calls: 128
- Tickets created: 42
- Bookings: 19
- Escalations: 7
- AI resolution rate: 72%
- Avg call duration: 3m 20s
- Missing KM items: 12

[Review Calls]
[Manage KM]
[AI Employees]
[Tickets]
[Analytics]
[Settings]
```

## 5.3 Platform Admin Portal

For SaaS owner/operator.

Main functions:

```text
Manage tenants
Manage plans and billing
Manage providers
Manage model routing
Manage voice providers
Manage global guardrails
Monitor usage and cost
Monitor incidents
Review tenant health
Control feature flags
Control schema/config versions
Audit logs
```

Platform admin dashboard:

```text
Platform Admin

Tenants: 120
Active tenants today: 86
Calls today: 18,420
AI cost today: $xxx
Voice provider cost: $xxx
Avg latency: x.x sec
Escalation rate: xx%
Error rate: xx%

[Tenants]
[Providers]
[Models]
[Billing]
[Analytics]
[Audit Logs]
[System Health]
```

---

## 6. Final Tech Stack

| Layer | Tech | Role |
|---|---|---|
| Customer web | **Svelte + shadcn-svelte** | Brand search, language select, AI workforce select, start call |
| Tenant admin web | **Svelte + shadcn-svelte** | KM setup, scope, avatars, tickets, review, dashboard |
| Platform admin web | **Svelte + shadcn-svelte** | Tenants, packages, KYC, billing, quotas, audit, monitoring |
| Mobile scale phase | Flutter + WebView | Wrap Svelte portals; native LiveKit call screen later |
| Backend API | Go + Fiber | REST/gRPC APIs, auth, tenant control, call orchestration |
| Realtime media | **LiveKit** | WebRTC audio rooms, AI agent participant, human handoff, SIP later |
| Event bus | **NATS.io** | Pub/sub, request/reply, service decoupling |
| Durable events | NATS JetStream | Replayable call, ticket, KM, billing, audit event streams |
| System of record | **Postgres** | Tenants, users, roles, tickets, billing records, call metadata |
| Cache / session | **Redis 8** | Active call state, sessions, rate limits, quotas, idempotency |
| Analytics + vector RAG | **ClickHouse** | Call/usage analytics, dashboards, **embedding storage, semantic KM search** |
| Object storage | **MinIO** (R2/S3 in prod) | Recordings, KM documents, transcripts, avatars, exports |
| Voice AI | Gemini Live / Grok Voice | Realtime AI employee intelligence behind `VoiceProvider` |

### 6.1 Stack decisions (v2.0)

```text
ClickHouse replaces pgvector, pg_duckdb, and early Postgres analytics extensions.
One analytical store for: event metrics, tenant dashboards, platform dashboards,
AND vector similarity search for tenant-scoped RAG.

Postgres stays lean: relational truth, transactions, billing, RBAC.
Redis 8 handles hot path only — never the event log or KM embeddings.
NATS is the nervous system between Go services and async workers.
LiveKit owns media transport; Go owns business rules.
Svelte + shadcn is the only web UI framework for MVP through scale phase 1.
```

### 6.2 ClickHouse dual role: analytics + vector RAG

| Use | ClickHouse table pattern |
|---|---|
| Call events | `call_events` — tenant_id, session_id, turn, latency, provider |
| Usage metering | `usage_events` — voice_seconds, llm_tokens, storage_bytes |
| QA / review | `qa_events` — resolution, escalation, missing_km flags |
| Vector RAG | `km_embeddings` — tenant_id, km_version, chunk_id, embedding Array(Float32), content |
| Semantic search | `cosineDistance(embedding, query_vector)` filtered by `tenant_id` + `km_scope` |

Workers ingest KM chunks from MinIO → embed via provider → write to ClickHouse.
Conversation orchestrator queries ClickHouse before each AI turn.

---

## 7. Frontend Strategy

## 7.1 Svelte-first Web Platform

Build all screens in Svelte first:

```text
Customer Portal
Tenant Admin Portal
Platform Admin Portal
Call Review Portal
AI Employee Builder
KM Setup Wizard
Ticket Dashboard
Analytics Dashboard
Billing Dashboard
Provider Configuration
```

Reasons:

- One web codebase for customer, tenant, and platform admin
- Faster MVP development
- Easy to embed in tenant websites
- Easy to wrap in Flutter WebView later
- Good fit for admin dashboards and structured forms

## 7.2 Flutter Scale Phase

Use Flutter later as a mobile container:

```text
Flutter App
  └── WebView
        └── Svelte Customer Portal / Tenant Portal
```

Flutter initially handles:

```text
Login token bridge
Push notification
Deep link
Secure storage
Mobile app shell
WebView navigation
Camera/microphone permission bridge
```

## 7.3 Native Flutter Call Screen Later

For production-scale mobile voice calls, do not rely on WebView forever. Use Flutter WebView for admin/review screens, but build the live call screen natively with LiveKit Flutter SDK later.

Recommended split:

```text
Svelte WebView:
- Dashboard
- KM setup
- Tickets
- Review
- Analytics
- Billing

Native Flutter:
- Live call screen
- Push-to-talk / audio permission
- Background call handling
- Mobile notification
- Native LiveKit call room
```

---

## 8. Target Architecture

```text
Customer / Tenant Staff / Platform Admin
        ↓
Svelte Web App + Shadcn UI
        ↓
Optional Flutter WebView Shell
        ↓
Go Fiber API Gateway
        ↓
Auth / Tenant / Brand / AI Employee / KM / Ticket APIs
        ↓
Conversation Orchestrator
        ↓
LiveKit Room
Customer + AI Agent + Optional Human Agent
        ↓
Gemini Live / Grok Voice Provider Adapter
        ↓
NATS Event Bus
call.started, turn.created, ticket.created, action.completed
        ↓
Postgres / Redis 8 / ClickHouse / MinIO
```

---

## 9. LiveKit Architecture

LiveKit should be the **media layer**, not the business logic layer.

```text
LiveKit = realtime audio/video/data room
Go Orchestrator = business logic
Gemini/Grok = AI reasoning/voice model
Postgres = source of truth
NATS = event bus
```

## 9.1 LiveKit Call Room Model

```text
Room: tenant_abc_call_123

Participants:
- customer
- ai_employee_bot
- human_agent_optional
- supervisor_optional
```

## 9.2 Main LiveKit Use Cases

```text
Web voice call
Mobile app voice call
Human agent joins call
Supervisor listen mode
Call recording
AI agent room participant
Phone/SIP bridge later
```

## 9.3 Call Flow with LiveKit

```text
1. Customer selects brand
2. Customer selects language
3. Customer selects AI employee
4. Go creates call_session
5. Go creates LiveKit room/token
6. Customer joins LiveKit room from Svelte
7. AI agent joins LiveKit room
8. Gemini Live / Grok Voice processes audio
9. Human can join if escalation happens
10. Call ends
11. Summary and ticket are generated
```

---

## 10. NATS Architecture

NATS should be the **event backbone** between backend modules.

Use **Core NATS** for fast request/reply and pub/sub.

Use **NATS JetStream** for durable workflow events.

## 10.1 NATS Responsibilities

```text
Event bus
Async jobs
Call lifecycle events
Ticket lifecycle events
KM processing pipeline
Usage/cost metering
Notification pipeline
Audit event distribution
```

## 10.2 Recommended NATS Subjects

```text
call.started
call.ended
call.turn.created
call.transcript.created
call.summary.created

conversation.intent.detected
conversation.language.detected
conversation.escalation.required

livekit.room.created
livekit.participant.joined
livekit.participant.left
livekit.recording.completed

ticket.created
ticket.updated
ticket.escalated
ticket.closed

action.requested
action.approved
action.completed
action.failed

km.document.uploaded
km.document.extracted
km.document.chunked
km.document.embedded
km.version.published
km.feedback.created

usage.voice.recorded
usage.llm.recorded
usage.storage.recorded
usage.call.completed

review.ai.flagged
review.human.completed
review.km_improvement.created
```

## 10.3 Tenant-scoped Subjects

For tenant isolation, use tenant-scoped subjects:

```text
tenant.{tenant_id}.call.started
tenant.{tenant_id}.ticket.created
tenant.{tenant_id}.km.version.published
tenant.{tenant_id}.usage.recorded
```

Important: Do not rely on subject naming alone for security. Always enforce tenant authorization in Go and database queries.

---

## 11. Backend Architecture with Go + Fiber

## 11.1 Backend Module Responsibilities

```text
api-gateway
identity-service
tenant-service
brand-service
ai-employee-service
km-service
call-session-service
conversation-orchestrator
ticket-service
action-service
review-service
billing-service
analytics-service
provider-adapter-service
notification-service
```

## 11.2 Recommended MVP Structure: Modular Monolith

Start with a modular monolith in Go. Split services later only when necessary.

```text
/apps/api
  /internal/auth
  /internal/tenant
  /internal/brand
  /internal/employee
  /internal/km
  /internal/call
  /internal/conversation
  /internal/ticket
  /internal/action
  /internal/review
  /internal/billing
  /internal/analytics
  /internal/provider
  /internal/storage
  /internal/eventbus
```

## 11.3 Go Responsibilities

```text
REST API
WebSocket/SSE control events
Tenant authentication
Tenant authorization
LiveKit token generation
AI employee orchestration
RAG orchestration
Graph rule checking
Ticket/action APIs
NATS publisher/subscriber
Provider abstraction
Audit logging
Usage metering
```

---

## 12. Data Architecture

## 12.1 Postgres as System of Record

Use Postgres for:

```text
tenants
tenant_users
brands
ai_employees
knowledge_documents
knowledge_chunks
call_sessions
tickets
action_requests
billing
audit_logs
```

Postgres is the source of truth.

## 12.2 Redis 8 for Active State

Use **Redis 8** for:

```text
Active call state and LiveKit session mapping
Temporary conversation context (TTL)
Rate limits and quota counters
Tenant/package entitlement cache
Presence and online agent indicators
Short-lived token and idempotency keys
Realtime dashboard cache (tenant + platform)
```

Do not use Redis as the event log or vector store. Use NATS JetStream for replayable events, ClickHouse for analytics/embeddings, and Postgres for durable records.

## 12.3 ClickHouse for Analytics and Vector RAG

ClickHouse is the **single analytical and retrieval engine** from day one.

**Analytics tables:**

```text
call_events
conversation_turn_events
token_usage_events
latency_events
provider_cost_events
qa_events
customer_behavior_events
audit_events (denormalized read model)
```

**Vector RAG tables:**

```text
km_embeddings        — chunk vectors scoped by tenant_id, km_version, ai_employee_scope
km_chunk_metadata    — source document, page, language, publish status
conversation_memory  — optional short-term semantic recall per call session
```

Example RAG flow:

```text
Customer asks: "How much is teeth whitening?"

System:
1. Resolve tenant + brand + AI employee scope (Postgres)
2. Embed query via EmbeddingProvider
3. ClickHouse: cosineDistance search on km_embeddings
   WHERE tenant_id = ? AND km_scope IN (employee_allowed_scopes)
4. Return top-k chunks with scores
5. AI answers only from approved KM; log chunks used to call_events
```

Scope rules (which KM an avatar may use) live in **Postgres**; vectors live in **ClickHouse**.

## 12.4 MinIO for Object Storage

Use object storage for:

```text
Call audio recordings
Uploaded PDFs
Uploaded images
Tenant KM files
Generated transcripts
Summary reports
Export files
Avatar images
Voice samples
```

Recommended abstraction:

```go
type ObjectStorage interface {
    PutObject(ctx context.Context, key string, data io.Reader, contentType string) error
    GetObject(ctx context.Context, key string) (io.ReadCloser, error)
    DeleteObject(ctx context.Context, key string) error
    PresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
```

Provider switching:

```text
Local dev → MinIO
Production low ops → Cloudflare R2
Enterprise/on-prem → MinIO
Future → AWS S3 / GCS / Azure Blob
```

---

## 13. AI Employee Model

Each AI employee is a configurable service agent.

## 13.1 AI Employee Fields

```text
Name
Role
Avatar
Voice
Language
Knowledge scope
Allowed actions
Escalation rules
Business script
Tone of voice
Provider/model config
Working hours
Fallback human team
Review policy
Version
Status
```

Example:

```text
AI Employee: Nong Mook
Role: Clinic Receptionist
Language: Thai / English
Voice: Friendly female voice
Knowledge: Clinic FAQ, doctors, prices, promotions
Actions: Create booking request, open ticket, send LINE message
Escalation: Medical emergency, legal issue, angry customer
```

## 13.2 AI Employee Types

```text
AI Receptionist
AI Sales Assistant
AI Booking Assistant
AI Product Advisor
AI Complaint Handler
AI Technical Support
AI Payment Follow-up Assistant
AI Customer Success Assistant
```

---

## 14. Knowledge Management Design

KM = Knowledge Management.

Tenant admin should be able to set up knowledge without technical skill.

## 14.1 KM Onboarding Flow

```text
Create tenant
  ↓
Create brand profile
  ↓
Choose business template
  ↓
Upload documents / FAQ / website links
  ↓
System extracts knowledge
  ↓
System chunks and embeds knowledge
  ↓
Admin reviews generated FAQ
  ↓
Create AI employees
  ↓
Assign KM to AI employees
  ↓
Run test calls
  ↓
Publish KM version
  ↓
Monitor missing questions
  ↓
Improve KM continuously
```

## 14.2 KM Types

```text
FAQ
Product/service information
Pricing
Branch/location
Opening hours
Promotions
Policy
Booking rules
Complaint process
Human escalation rules
Legal disclaimers
Script examples
Customer objection handling
```

## 14.3 KM Versioning

Important for trust and audit.

```text
KM Version 1.0
- Initial FAQ

KM Version 1.1
- Added promotion questions
- Fixed incorrect opening hours

KM Version 1.2
- Added complaint handling flow
```

Every answer should know:

```text
Which KM version was used?
Which document was retrieved?
Which AI employee answered?
Which model/provider generated the response?
```

---

## 15. Provider Abstraction

Business logic must not depend directly on Gemini Live, Grok Voice, MinIO, R2, ClickHouse client SDKs in UI, or any single provider.

Create interfaces for:

```text
VoiceProvider
LLMProvider
EmbeddingProvider
ObjectStorageProvider
TicketProvider
NotificationProvider
AnalyticsProvider
BillingProvider
```

## 15.1 Voice Provider Interface

```go
type VoiceProvider interface {
    StartSession(ctx context.Context, req StartVoiceSessionRequest) (*VoiceSession, error)
    SendAudio(ctx context.Context, sessionID string, audio []byte) error
    ReceiveEvent(ctx context.Context, sessionID string) (*VoiceEvent, error)
    SendToolResult(ctx context.Context, sessionID string, result ToolResult) error
    EndSession(ctx context.Context, sessionID string) error
}
```

Provider adapters:

```text
GeminiLiveProvider
GrokVoiceProvider
OpenAIRealtimeProvider later
ElevenLabsProvider later
SIPProvider later
```

## 15.2 Storage Provider Interface

```go
type StorageProvider interface {
    Put(ctx context.Context, key string, body io.Reader, contentType string) error
    Get(ctx context.Context, key string) (io.ReadCloser, error)
    Delete(ctx context.Context, key string) error
    SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
```

## 15.3 Embedding Provider Interface

```go
type EmbeddingProvider interface {
    EmbedText(ctx context.Context, text string, model string) ([]float32, error)
    EmbedBatch(ctx context.Context, texts []string, model string) ([][]float32, error)
}
```

## 15.4 Notification Provider Interface

```go
type NotificationProvider interface {
    SendLine(ctx context.Context, req LineMessageRequest) error
    SendEmail(ctx context.Context, req EmailRequest) error
    SendSMS(ctx context.Context, req SMSRequest) error
    SendPush(ctx context.Context, req PushRequest) error
}
```

---

## 16. Recommended Database Domains

## 16.1 Multi-tenant Core

```text
tenants
tenant_profiles
tenant_settings
tenant_users
roles
permissions
tenant_user_roles
audit_logs
```

## 16.2 Brand and Channels

```text
brands
brand_languages
brand_channels
brand_contact_points
brand_search_index
brand_search_keywords
```

## 16.3 AI Employees

```text
ai_employees
ai_employee_versions
ai_employee_languages
ai_employee_voices
ai_employee_tools
ai_employee_km_scopes
ai_employee_guardrails
ai_employee_provider_configs
```

## 16.4 Knowledge Management

```text
knowledge_bases
knowledge_documents
knowledge_chunks
embedding_models
km_embeddings (ClickHouse)
km_versions
km_publish_logs
km_test_cases
km_feedback
```

## 16.5 Calls and Conversations

```text
call_sessions
conversation_turns
call_recordings
call_transcripts
call_summaries
call_events
call_intents
call_sentiments
call_provider_events
call_quality_reviews
```

## 16.6 Tickets and Actions

```text
tickets
ticket_events
ticket_assignments
action_requests
action_results
escalations
human_handoffs
escalation_rules
```

## 16.7 Billing and Usage

```text
plans
subscriptions
usage_records
provider_cost_records
invoices
payments
receipts
tax_invoices
```

---

## 17. Official 35-Sprint Roadmap

Sprint planning follows a **customer-first conversation wedge**, then platform foundations, tenant commerce, configuration, customer identity, operations, production tuning, and infra scale.

> **Prototype note:** `monti-jarvis` v0.1.0 delivered a Go spike for **Sprint 21** (AI workforce selection + conversation) using embedded HTML and direct Gemini relay. Official product delivery moves this to **Svelte + shadcn + LiveKit + NATS** per the stack in Section 6.

### 17.1 Sprint index

| Sprint | Platform | Feature | Phase |
| ---: | --- | --- | --- |
| 1 | Customer | Conversation | Core experience |
| 2 | Customer | Add KM and Scope | Core experience |
| 3 | Backend | Auth | Foundation |
| 4 | Platform Admin | Packages | Commercial |
| 5 | Platform Admin | Avatars | Commercial |
| 6 | Tenant | Register | Onboarding |
| 7 | Platform Admin | KYC Tenant | Onboarding |
| 8 | Platform Admin | Payment Gateway | Commerce |
| 9 | Tenant | Buy Package | Commerce |
| 10 | Platform Admin | Billing | Commerce |
| 11 | Platform Admin | Receipt | Commerce |
| 12 | Tenant | Tax Invoice | Commerce |
| 13 | Platform Admin | Quota, Rate Limit | Governance |
| 14 | Tenant | Embed to Web | Distribution |
| 15 | Tenant | Set Scope and KM | Configuration |
| 16 | Tenant | Settings, Locale, Limit, Quota | Configuration |
| 17 | Tenant | Test and Preview | Configuration |
| 18 | Tenant | Customer Tier | Configuration |
| 19 | Customer | Register | Customer identity |
| 20 | Customer | Auth | Customer identity |
| 21 | Customer | Select AI Workforce to Conversation | Core experience |
| 22 | Platform / Tenant | Conversation Records | Operations |
| 23 | Tenant | Tickets | Operations |
| 24 | Tenant | Review | Operations |
| 25 | Tenant | Dashboard | Operations |
| 26 | Tenant | Monitoring | Operations |
| 27 | Platform | Audit Log | Platform ops |
| 28 | Platform | Monitoring | Platform ops |
| 29 | Platform | Dashboard | Platform ops |
| 30 | Platform | Monitoring | Platform ops |
| 31 | Tuning | gRPC, Cache on Prod | Production |
| 32 | Tuning | Partition, Hardening | Production |
| 33 | Infra | Scale, Auto Scale | Infra |
| 34 | Infra | Canary Deployment | Infra |
| 35 | Infra | Backup Restore Archive | Infra |

### 17.2 Phase summary

**Phase A — Customer core (Sprints 1–2, 21)**  
Ship conversation first to prove value; add KM + scope so answers are tenant-grounded; formalize AI workforce picker in Svelte with LiveKit rooms.

**Phase B — Platform foundation (Sprints 3–5, 13)**  
Backend auth (tenant/platform/customer roles), commercial packages, platform-managed avatar catalog, quotas and rate limits.

**Phase C — Tenant onboarding & commerce (Sprints 6–12)**  
Registration, KYC, payment gateway, package purchase, billing, receipts, tax invoices.

**Phase D — Tenant go-live (Sprints 14–18)**  
Web embed widget, scope/KM assignment, locale/settings/limits, test/preview sandbox, customer tier rules.

**Phase E — Customer identity (Sprints 19–20)**  
Optional registered customers for history, tier benefits, and returning-call context.

**Phase F — Operations (Sprints 22–26)**  
Conversation records, tickets, QA review, tenant dashboard and monitoring (ClickHouse-backed).

**Phase G — Platform operations (Sprints 27–30)**  
Cross-tenant audit log, platform monitoring and dashboards.

**Phase H — Production & infra (Sprints 31–35)**  
gRPC internal APIs, Redis 8 cache strategy, ClickHouse partitioning, security hardening, autoscale, canary deploys, backup/archive.

### 17.3 Sprint 1 — Customer: Conversation (next official sprint)

**Goal:** End customer can start a voice conversation with an AI agent through Svelte + LiveKit.

**In scope:**
- Svelte + shadcn customer portal shell
- LiveKit room create/join/token API (Go)
- Basic conversation UI (start/end call, transcript stream)
- Call session persisted in Postgres; active state in Redis 8
- NATS `call.started` / `call.ended` events

**Out of scope:** Auth, KM RAG, tickets, billing, tenant admin.

**Stack touchpoints:** Svelte, LiveKit, Go/Fiber, Postgres, Redis 8, NATS, MinIO (recording stub).

### 17.4 Sprint 2 — Customer: Add KM and Scope

**Goal:** AI answers use tenant-approved knowledge with scope enforcement.

**In scope:**
- KM upload → MinIO → chunk → embed → **ClickHouse `km_embeddings`**
- Scope model in Postgres (which KM sets apply to which conversation context)
- RAG retrieval in conversation orchestrator before each AI turn
- Missing-KM detection flag to ClickHouse `qa_events`

**Out of scope:** Full tenant admin KM wizard (Sprint 15), billing metering.

### 17.5 Dependency graph (simplified)

```text
Sprint 1 Conversation
  → Sprint 2 KM + Scope
  → Sprint 21 Workforce Selection (enhances 1)
Sprint 3 Auth
  → Sprint 6 Register → Sprint 7 KYC
Sprint 4 Packages + Sprint 5 Avatars
  → Sprint 9 Buy Package → Sprint 10–12 Billing
Sprint 13 Quota
  → Sprint 16 Tenant Limits
Sprint 15 Scope/KM (tenant admin)
  → Sprint 17 Test/Preview
Sprint 22 Records
  → Sprint 23 Tickets → Sprint 24 Review
  → Sprint 25–26 Tenant Dashboard/Monitoring
Sprint 27 Audit
  → Sprint 28–30 Platform Monitoring/Dashboard
Sprint 31–32 Tuning
  → Sprint 33–35 Infra
```

Full SDLC tree: `docs/sdlc/README.md` — roadmap, features, design, sprints, tasks.

---

## 18. MVP Scope (mapped to sprints)

## MVP Phase 1 — Customer can talk (Sprints 1–2, 21)

```text
Svelte + shadcn-svelte
Go + Fiber
Postgres
Redis 8
NATS.io + JetStream
LiveKit
MinIO
ClickHouse (analytics + vector RAG from Sprint 2)
Gemini Live / Grok Voice
```

Features:

```text
Sprint 1:  Voice conversation via LiveKit + transcript
Sprint 2:  KM upload, scope, ClickHouse RAG retrieval
Sprint 21: AI workforce / avatar selection in conversation flow
```

## MVP Phase 2 — Platform can sell (Sprints 3–13)

```text
Auth (backend + customer + tenant roles)
Packages, avatars, registration, KYC
Payment gateway, buy package, billing, receipt, tax invoice
Quota and rate limits (Redis 8 + Postgres entitlements)
```

## MVP Phase 3 — Tenant can launch (Sprints 14–18)

```text
Web embed widget
Tenant scope + KM configuration
Locale, limits, quotas, test/preview sandbox
Customer tier rules
```

## MVP Phase 4 — Operate at scale (Sprints 19–30)

```text
Customer register/auth
Conversation records, tickets, review
Tenant + platform dashboards (ClickHouse)
Audit log, monitoring
```

## MVP Phase 5 — Production grade (Sprints 31–35)

```text
gRPC internal services, Redis 8 cache tuning
ClickHouse partitioning and hardening
Autoscale, canary deployment, backup/restore/archive
```

## Future: Mobile and telephony

```text
Flutter WebView shell (post Sprint 35 or parallel track)
SIP / phone bridge via LiveKit
Human agent handoff and supervisor listen mode
```

---

## 19. Runtime Flow

```text
1. Customer searches brand
2. Customer selects language
3. Customer selects AI employee
4. Customer starts call
5. Svelte calls Go API
6. Go creates call_session
7. Go creates LiveKit room and token
8. Redis stores active session state
9. Customer joins LiveKit room
10. AI employee joins LiveKit room
11. Voice provider opens realtime session
12. AI employee prompt is loaded
13. RAG searches tenant KM through ClickHouse vector similarity (scoped)
14. Scope and policy rules checked in Postgres (tenant entitlements, avatar KM allow-list)
15. AI answers customer
16. Tool/action request is validated
17. Ticket/action is created
18. Events are published to NATS
19. Transcript and audio are saved
20. Summary is generated
21. Tenant admin reviews call
22. Missing KM becomes improvement task
```

---

## 20. Security, Governance, and Multi-tenancy

## 20.1 Tenant Isolation

Required controls:

```text
Every table has tenant_id
Every API checks tenant permission
Every KM search filters tenant_id
Every ClickHouse vector search filters tenant_id and km_scope
Every object storage key includes tenant prefix
Every NATS event includes tenant context
Every audit log includes tenant_id, actor_id, action, and timestamp
```

## 20.2 Role-based Access Control

Example roles:

```text
Platform Owner
Platform Admin
Platform Support
Tenant Owner
Tenant Admin
Tenant Agent
Tenant Reviewer
Tenant Billing Manager
Customer
```

## 20.3 Guardrails

Guardrail examples:

```text
Answer only from approved tenant KM for company-specific questions
Escalate sensitive/legal/medical/financial advice
Do not reveal other tenant data
Do not perform high-risk actions without approval
Do not invent pricing or policy
Log every action request and result
Track which KM version was used
```

## 20.4 Audit Fields

Important tables should include:

```text
created_at
created_by
updated_at
updated_by
deleted_at
version
schema_version
tenant_id
```

---

## 21. Analytics and Metrics

## 21.1 Business Metrics

```text
Total calls
Answered calls
Missed calls reduced
Tickets created
Bookings created
Leads created
Conversion rate
Human workload reduced
Revenue influenced
```

## 21.2 AI Quality Metrics

```text
AI resolution rate
Escalation rate
Wrong answer rate
Missing KM rate
Customer satisfaction
Average confidence score
Review pass rate
Hallucination risk flag
```

## 21.3 Platform Metrics

```text
Tenant usage
Cost per call
Cost per tenant
Voice provider cost
LLM provider cost
Latency
Error rate
SLA
Provider availability
```

## 21.4 Review Metrics

```text
Calls reviewed
Calls passed
Calls failed
Human correction count
KM improvement count
Prompt improvement count
Escalation accuracy
```

---

## 22. Recommended Final Architecture Summary

```text
Frontend:
- Svelte + shadcn-svelte for Customer, Tenant, and Platform Admin portals
- Flutter WebView later for mobile scale

Realtime:
- LiveKit for call rooms and audio transport

Backend:
- Go + Fiber (REST) + gRPC (internal, Sprint 31)
- NATS.io + JetStream for event-driven services
- Redis 8 for active state, cache, quotas, rate limits

Data:
- Postgres — system of record (tenants, billing, RBAC, tickets, metadata)
- ClickHouse — analytics dashboards AND vector RAG embeddings
- MinIO — recordings, KM files, avatars, exports

AI:
- Gemini Live / Grok Voice behind VoiceProvider interface
```

The key design decisions:

```text
Svelte + shadcn is the only web UI for MVP.
LiveKit owns realtime media; Go owns orchestration.
NATS is the backend nervous system.
Postgres is transactional truth.
ClickHouse is analytical truth AND semantic retrieval.
Redis 8 is hot-path cache only.
MinIO is durable object storage.
```

---

## 23. Practical MVP Architecture Diagram

```text
[Svelte Customer Web]
       |
       | REST / WebSocket / LiveKit Token
       v
[Go Fiber API Gateway]
       |
       | create room/token
       v
[LiveKit Room] <------> [AI Agent Worker]
       |                       |
       | audio/data             | Gemini Live / Grok Voice
       v                       v
[Call Session Runtime] ---> [AI Orchestrator]
       |                       |
       |                       | RAG / Tools / Actions
       v                       v
[Redis 8 Active State]   [Postgres Metadata]
       |                       |
       | events                | scope / billing / tickets
       v                       v
[NATS / JetStream]       [ClickHouse]
       |                  analytics + km_embeddings (RAG)
       +--> KM Embed Worker
       +--> Summary Worker
       +--> Ticket Worker
       +--> Usage Worker
       +--> Notification Worker
       |
       v
[MinIO — recordings, KM files, avatars]
```

---

## 24. Recommended Development Principle

Build the platform around stable interfaces and domain modules.

```text
Do not build Gemini-specific business logic.
Do not build Grok-specific business logic.
Do not build MinIO-specific business logic.
Do not leak ClickHouse or MinIO specifics into Svelte UI — use Go service APIs.
Do not couple Flutter directly to backend internals.
```

Instead:

```text
Use provider interfaces.
Use event-driven workflows.
Use tenant-scoped data access.
Use versioned KM and AI employee configs.
Use Svelte as the first product surface.
Use Flutter only when mobile scale becomes necessary.
```

---

## 25. Final Recommendation

**Start Sprint 1 immediately:** Customer conversation (Svelte + LiveKit + NATS), not tenant admin or billing.

**Commercial wedge** (unchanged):

```text
Inbound AI Receptionist for Clinic / Real Estate / Education
```

**Build order rationale:**

| Priority | Why |
| --- | --- |
| Conversation first (Sprint 1) | Proves core product value before commerce complexity |
| KM + scope (Sprint 2) | Grounds AI answers; ClickHouse RAG from day one |
| Auth + packages (Sprints 3–5) | Enables multi-tenant SaaS once conversation works |
| Tenant commerce (Sprints 6–12) | Monetization after platform skeleton exists |
| Workforce picker (Sprint 21) | Enhances conversation UX; v0.1.0 prototype already validates this |
| Operations (Sprints 22–30) | Records, tickets, review, dashboards — ClickHouse-native |
| Tuning + infra (Sprints 31–35) | Production hardening after feature-complete beta |

**Retire from v1.0 blueprint:** pgvector, pg-apache-age, pg_duckdb as primary stores. All analytics and vector RAG consolidate into **ClickHouse**; Postgres stays relational-only.

This gives Monti a practical, scalable, and provider-agnostic foundation for becoming a multi-tenant AI Call Center platform.
