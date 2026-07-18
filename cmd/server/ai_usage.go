package main

import (
	"context"
	"strconv"
	"time"

	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/metering"
)

func (s *server) recordAIUsage(tenantID, callID, conversationID, message string, result gemini.ReplyResult, sequence int) {
	if s == nil || s.aiMeter == nil {
		return
	}
	event := metering.UsageEvent{
		EventID:              metering.StableEventID("text", tenantID, callID, strconv.Itoa(sequence)+":"+message),
		TenantID:             tenantID,
		CallID:               callID,
		ConversationRecordID: conversationID,
		Provider:             "gemini",
		Model:                result.Model,
		Modality:             "text",
		InputTokens:          result.Usage.PromptTokenCount,
		OutputTokens:         result.Usage.CandidatesTokenCount,
		UsageDate:            time.Now().UTC(),
		SourceUpdatedAt:      time.Now().UTC(),
	}
	go func() {
		_ = s.aiMeter.Record(context.Background(), event)
	}()
}
