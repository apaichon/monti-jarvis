# Feature: AI Avatar Workforce + Inbound Q&A   (feature:workforce-qa)
**Sprint:** SPRINT-001   **Owner:** DEV

## Problem

Inbound callers need a staffed call center experience without waiting for human agents. Monti must offer selectable AI avatar agents that can hold a conversation and answer common questions by voice or text.

## Scope

In:
- Four AI avatar agents with distinct roles, voices, and greetings
- Workforce catalog API and selection UI
- Text chat with per-agent system prompts and topic tags (general, billing, technical)
- Voice-to-voice inbound calls via Gemini Live
- Isolated local infra for call sessions

Out:
- Authentication, KYC, ticketing, CRM lookup, knowledge-base RAG, call recording, supervisor tools (→ backlog)

## Acceptance criteria

1. Caller can list and select any workforce agent before starting a conversation.
2. Text messages receive answers shaped by the selected agent's role prompt.
3. Voice calls use the selected agent's Gemini voice and stay connected for multi-turn Q&A.
4. Transcripts appear in the caller desk for both text and voice turns.
5. Sessions persist exchange metadata when Postgres is available.

## Test notes

- Functional: `GET /api/workforce`, browser agent selection, `POST /api/chat` per agent, voice call smoke with mic permission.
- Languages: Thai + English where user-facing (agent prompts support both).
- Safety: agents must not solicit passwords, OTPs, or full payment credentials.

## Dependencies

- components: `internal/workforce`, `internal/gemini`, `internal/live`, `internal/web`
- reference: Jarvis Chat patterns at `/Users/ar677018/Projects/libra/jarvis`