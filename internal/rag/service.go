package rag

import (
	"context"
	"fmt"
	"strings"

	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/scope"
)

const (
	// Slightly lower than 0.55 so bilingual product FAQs (TH question / EN KM) still match.
	minScore             = 0.42
	voiceChunkLimit      = 8
	voiceExcerptRunes    = 480
	searchCandidateLimit = 80
)

type Source struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Filename   string  `json:"filename,omitempty"`
	Scope      string  `json:"scope"`
	Excerpt    string  `json:"excerpt"`
	Score      float64 `json:"score"`
}

type Result struct {
	Sources      []Source
	ContextBlock string
	MissingKM    bool
}

type Service struct {
	ch     *clickhouse.Client
	embed  *gemini.Client
	tenant string
	cache  *preloadCache
}

func New(ch *clickhouse.Client, embed *gemini.Client, tenantID string) *Service {
	return &Service{ch: ch, embed: embed, tenant: tenantID, cache: newPreloadCache()}
}

func (s *Service) WithTenant(tenantID string) *Service {
	if s == nil {
		return nil
	}
	cp := *s
	if t := strings.TrimSpace(tenantID); t != "" {
		cp.tenant = t
	}
	return &cp
}

func (s *Service) Enabled() bool {
	return s != nil && s.embed != nil && s.embed.Enabled() && s.ch != nil && s.ch.Enabled()
}

func (s *Service) Retrieve(ctx context.Context, agentID, topic, question string) (Result, error) {
	if !s.Enabled() {
		return Result{}, nil
	}
	question = strings.TrimSpace(question)
	scopes := scope.Resolve(agentID, topic)

	var hits []clickhouse.ChunkHit
	var err error
	if question != "" {
		vec, embedErr := s.embed.EmbedQuery(ctx, question)
		if embedErr != nil {
			return Result{}, embedErr
		}
		// 1) Preferred: this agent + resolved scopes
		hits, err = s.ch.Search(ctx, s.tenant, agentID, scopes, vec, 5, searchCandidateLimit)
		if err != nil {
			return Result{}, err
		}
		result := s.buildResult(hits, question, agentID, topic)
		if !result.MissingKM {
			return result, nil
		}
		// 2) Same agent, all scopes (doc tagged technical while topic is general)
		hits, err = s.ch.Search(ctx, s.tenant, agentID, nil, vec, 5, searchCandidateLimit)
		if err != nil {
			return Result{}, err
		}
		result = s.buildResult(hits, question, agentID, topic)
		if !result.MissingKM {
			return result, nil
		}
		// 3) Tenant-wide any agent (KM uploaded under Luna but embed chat uses Ava)
		hits, err = s.ch.Search(ctx, s.tenant, "", nil, vec, 5, searchCandidateLimit)
		if err != nil {
			return Result{}, err
		}
		return s.buildResult(hits, question, agentID, topic), nil
	}
	hits, err = s.ch.ListAgentChunks(ctx, s.tenant, agentID, scopes, voiceChunkLimit)
	if err != nil {
		return Result{}, err
	}
	return s.buildResult(hits, question, agentID, topic), nil
}

// RetrieveForVoice preloads KB for Gemini Live setup (no query embedding — keep setup fast).
// Same fallbacks as chat: agent scopes → agent all scopes → tenant-wide.
func (s *Service) RetrieveForVoice(ctx context.Context, agentID, topic string) (Result, error) {
	if !s.Enabled() {
		return Result{}, nil
	}
	key := s.tenant + ":" + agentID + ":" + topic
	if cached, ok := s.cache.get(key); ok {
		return cached, nil
	}
	scopes := scope.Resolve(agentID, topic)
	hits, err := s.ch.ListAgentChunks(ctx, s.tenant, agentID, scopes, voiceChunkLimit)
	if err != nil {
		return Result{}, err
	}
	if len(hits) == 0 {
		hits, err = s.ch.ListAgentChunks(ctx, s.tenant, agentID, nil, voiceChunkLimit)
		if err != nil {
			return Result{}, err
		}
	}
	if len(hits) == 0 {
		hits, err = s.ch.ListAgentChunks(ctx, s.tenant, "", nil, voiceChunkLimit)
		if err != nil {
			return Result{}, err
		}
	}
	// Empty question keeps all hits; expand excerpts for product prices/SKUs.
	srcs := make([]Source, 0, len(hits))
	for _, h := range hits {
		srcs = append(srcs, Source{
			ChunkID:    h.ChunkID,
			DocumentID: h.DocumentID,
			Scope:      h.KMScope,
			Excerpt:    excerpt(h.Content, voiceExcerptRunes),
			Score:      1,
		})
	}
	result := Result{Sources: srcs, ContextBlock: formatVoiceContext(srcs)}
	s.cache.set(key, result)
	return result, nil
}

func (s *Service) buildResult(hits []clickhouse.ChunkHit, question, agentID, topic string) Result {
	result := Result{}
	for _, hit := range hits {
		if question != "" && hit.Score < minScore {
			continue
		}
		result.Sources = append(result.Sources, Source{
			ChunkID:    hit.ChunkID,
			DocumentID: hit.DocumentID,
			Scope:      hit.KMScope,
			Excerpt:    excerpt(hit.Content, 220),
			Score:      hit.Score,
		})
	}
	if question != "" && len(result.Sources) == 0 {
		result.MissingKM = true
		_ = s.ch.InsertQAEvent(context.Background(), s.tenant, agentID, topic, question, "missing_km")
	}
	result.ContextBlock = formatContext(result.Sources)
	return result
}

func (s *Service) AugmentPrompt(basePrompt, agentID, topic, question string, rag Result) string {
	basePrompt = strings.TrimSpace(basePrompt)
	if rag.ContextBlock == "" {
		if rag.MissingKM {
			return basePrompt + `

No approved knowledge-base chunks matched this question.
If you cannot answer from general role guidance alone, say you do not have that information in the knowledge base and offer to connect the caller with a human specialist.`
		}
		return basePrompt
	}
	return basePrompt + `

Use ONLY the following approved knowledge-base excerpts when answering. If the answer is not supported by these excerpts, say you do not have that information in the knowledge base.

` + rag.ContextBlock
}

func (s *Service) BuildVoicePrompt(basePrompt, agentID, topic string, rag Result) string {
	basePrompt = strings.TrimSpace(basePrompt)
	if rag.ContextBlock == "" {
		return basePrompt + `

Voice call — keep replies short (1–3 sentences). If you lack product facts, say you will check with a specialist.`
	}
	return basePrompt + `

You HAVE an approved knowledge base for this tenant. Prefer these facts for product prices, sizes, and promotions. Quote promo prices accurately. Reply in the caller's language. Keep spoken answers short (1–3 sentences), then offer more help.

Knowledge base excerpts:
` + rag.ContextBlock
}

func formatVoiceContext(sources []Source) string {
	if len(sources) == 0 {
		return ""
	}
	var b strings.Builder
	for i, src := range sources {
		text := excerpt(src.Excerpt, voiceExcerptRunes)
		fmt.Fprintf(&b, "[%d] %s\n", i+1, text)
	}
	return strings.TrimSpace(b.String())
}

func formatContext(sources []Source) string {
	if len(sources) == 0 {
		return ""
	}
	var b strings.Builder
	for i, src := range sources {
		fmt.Fprintf(&b, "[%d] scope=%s score=%.2f\n%s\n\n", i+1, src.Scope, src.Score, src.Excerpt)
	}
	return strings.TrimSpace(b.String())
}

func excerpt(text string, max int) string {
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return text[:max] + "…"
}