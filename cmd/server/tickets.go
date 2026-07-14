package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/store"
)

type createTicketRequest struct {
	ConversationRecordID string `json:"conversation_record_id"`
	CallID               string `json:"call_id"`
	ConfirmEscalation    bool   `json:"confirm_escalation"`
	Subject              string `json:"subject"`
	Description          string `json:"description"`
	Category             string `json:"category"`
	ContactName          string `json:"contact_name"`
	ContactEmail         string `json:"contact_email"`
}

type patchTicketRequest struct {
	Status         *string `json:"status"`
	Priority       *string `json:"priority"`
	AssigneeUserID *string `json:"assignee_user_id"`
}

type ticketEventRequest struct {
	Type string `json:"type"`
	Note string `json:"note"`
}

type ticketOffer struct {
	Subject  string `json:"subject"`
	Category string `json:"category"`
	Reason   string `json:"reason"`
}

func ticketOfferForMessage(message, topic string) *ticketOffer {
	originalMessage := strings.TrimSpace(message)
	message = strings.ToLower(originalMessage)
	if originalMessage == "" {
		return nil
	}
	for _, signal := range []string{
		"human agent", "live agent", "real person", "speak to a person", "talk to a person",
		"speak with someone", "talk to someone", "escalate", "มนุษย์", "เจ้าหน้าที่", "คุยกับคน",
		"ขอคน", "ติดต่อคน",
	} {
		if strings.Contains(message, signal) {
			category := strings.TrimSpace(strings.ToLower(topic))
			if category != "billing" && category != "technical" {
				category = "general"
			}
			return &ticketOffer{
				Subject:  "Human follow-up requested",
				Category: category,
				Reason:   "Customer context: " + boundedTicketContext(originalMessage),
			}
		}
	}
	return nil
}

func (s *server) createCustomerTicket(w http.ResponseWriter, r *http.Request) {
	var req createTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTicketError(w, http.StatusBadRequest, "invalid JSON", "validation_error")
		return
	}
	if !req.ConfirmEscalation {
		writeTicketError(w, http.StatusBadRequest, "customer confirmation is required", "validation_error")
		return
	}
	tenantID := s.publicCustomerTenantID(r)
	ctx, ok := s.resolveCustomerPortalContext(w, r, true)
	if !ok {
		return
	}
	if ctx.TenantID != tenantID {
		tenantID = ctx.TenantID
	}
	if strings.TrimSpace(req.ContactEmail) == "" && ctx.Customer == nil {
		writeTicketError(w, http.StatusBadRequest, "contact email is required", "contact_required")
		return
	}
	if req.ContactEmail != "" && !strings.Contains(req.ContactEmail, "@") {
		writeTicketError(w, http.StatusBadRequest, "contact email is invalid", "validation_error")
		return
	}

	input := store.TicketInput{
		TenantID:             tenantID,
		ConversationRecordID: strings.TrimSpace(req.ConversationRecordID),
		CallID:               strings.TrimSpace(req.CallID),
		Subject:              req.Subject,
		Description:          req.Description,
		Category:             req.Category,
		ContactName:          req.ContactName,
		ContactEmail:         req.ContactEmail,
		IdempotencyKey:       strings.TrimSpace(r.Header.Get("Idempotency-Key")),
		ActorType:            "system",
	}
	var sourceRecord *store.ConversationRecord
	if ctx.Customer != nil {
		input.CustomerID = ctx.Customer.ID
		input.ActorType = "customer"
		input.ActorID = ctx.Customer.ID
		if input.ContactName == "" {
			input.ContactName = ctx.Customer.DisplayName
		}
		if input.ContactEmail == "" {
			input.ContactEmail = ctx.Customer.Email
		}
	}
	if input.ConversationRecordID != "" {
		record, err := s.store.GetConversationRecord(r.Context(), tenantID, input.ConversationRecordID)
		if err != nil {
			writeTicketError(w, http.StatusNotFound, "source conversation not found", "not_found")
			return
		}
		if input.CallID == "" {
			input.CallID = record.CallID
		}
		input.AvatarID = record.AvatarID
		if input.CustomerID == "" {
			input.CustomerID = record.CustomerID
		}
		sourceRecord = &record
	}
	if input.CallID != "" && sourceRecord == nil {
		if record, err := s.store.GetConversationRecordByCallID(r.Context(), tenantID, input.CallID); err == nil {
			input.ConversationRecordID = record.ID
			input.AvatarID = record.AvatarID
			if input.CustomerID == "" {
				input.CustomerID = record.CustomerID
			}
			sourceRecord = &record
		}
	}
	if input.CallID == "" {
		writeTicketError(w, http.StatusBadRequest, "call_id or conversation_record_id is required", "validation_error")
		return
	}
	if err := s.store.ValidateCallReference(r.Context(), tenantID, input.CallID); err != nil {
		writeTicketError(w, http.StatusNotFound, "source conversation not found", "not_found")
		return
	}
	if sourceRecord != nil {
		input.SourceSummary = ticketSourceSummary(*sourceRecord, req.Description)
	}
	ticket, existed, err := s.store.CreateTicket(r.Context(), input)
	if err != nil {
		writeTicketStoreError(w, err)
		return
	}
	if existed {
		writeJSON(w, http.StatusOK, map[string]any{"ticket": ticketPublicJSON(ticket), "idempotent": true})
		return
	}
	s.publishTicketEvent("ticket.created", ticket, "customer")
	writeJSON(w, http.StatusCreated, map[string]any{"ticket": ticketPublicJSON(ticket)})
}

func (s *server) listTenantTickets(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeTicketError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	filters := store.TicketFilters{
		StartDate:      strings.TrimSpace(r.URL.Query().Get("start_date")),
		EndDate:        strings.TrimSpace(r.URL.Query().Get("end_date")),
		Status:         strings.TrimSpace(r.URL.Query().Get("status")),
		Priority:       strings.TrimSpace(r.URL.Query().Get("priority")),
		Category:       strings.TrimSpace(r.URL.Query().Get("category")),
		AvatarID:       strings.TrimSpace(r.URL.Query().Get("avatar_id")),
		CustomerID:     strings.TrimSpace(r.URL.Query().Get("customer_id")),
		AssigneeUserID: strings.TrimSpace(r.URL.Query().Get("assignee_user_id")),
	}
	if err := validateTicketDates(filters.StartDate, filters.EndDate); err != nil {
		writeTicketError(w, http.StatusBadRequest, err.Error(), "validation_error")
		return
	}
	tickets, err := s.store.ListTickets(r.Context(), tenantID, filters)
	if err != nil {
		writeTicketError(w, http.StatusBadGateway, err.Error(), "tickets_unavailable")
		return
	}
	rows := make([]map[string]any, 0, len(tickets))
	for _, ticket := range tickets {
		rows = append(rows, ticketListJSON(ticket))
	}
	writeJSON(w, http.StatusOK, map[string]any{"tickets": rows, "next_cursor": nil})
}

func (s *server) getTenantTicket(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeTicketError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	ticket, err := s.store.GetTicket(r.Context(), tenantID, r.PathValue("id"))
	if err != nil {
		writeTicketError(w, http.StatusNotFound, "ticket not found", "not_found")
		return
	}
	if len(ticket.SourceSummary) == 0 && ticket.CallID != "" {
		if record, sourceErr := s.store.GetConversationRecordByCallID(r.Context(), tenantID, ticket.CallID); sourceErr == nil {
			ticket.SourceSummary = ticketSourceSummary(record, ticket.Description)
			if ticket.ConversationRecordID == "" {
				ticket.ConversationRecordID = record.ID
			}
		}
	}
	events, err := s.store.ListTicketEvents(r.Context(), tenantID, ticket.ID)
	if err != nil {
		writeTicketError(w, http.StatusBadGateway, err.Error(), "tickets_unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ticket": ticketTenantJSON(ticket), "events": events})
}

func (s *server) patchTenantTicket(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeTicketError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req patchTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTicketError(w, http.StatusBadRequest, "invalid JSON", "validation_error")
		return
	}
	ticket, event, err := s.store.UpdateTicket(r.Context(), tenantID, r.PathValue("id"), req.Status, req.Priority, req.AssigneeUserID)
	if err != nil {
		writeTicketStoreError(w, err)
		return
	}
	if event.ID != "" {
		s.publishTicketEvent("ticket.updated", ticket, event.EventType)
	}
	writeJSON(w, http.StatusOK, map[string]any{"ticket": ticketTenantJSON(ticket)})
}

func (s *server) addTenantTicketEvent(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeTicketError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return
	}
	var req ticketEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeTicketError(w, http.StatusBadRequest, "invalid JSON", "validation_error")
		return
	}
	if strings.TrimSpace(req.Type) != "note" {
		writeTicketError(w, http.StatusBadRequest, "only note events are supported", "validation_error")
		return
	}
	ticket, event, err := s.store.AddTicketNote(r.Context(), tenantID, r.PathValue("id"), req.Note)
	if err != nil {
		writeTicketStoreError(w, err)
		return
	}
	s.publishTicketEvent("ticket.updated", ticket, event.EventType)
	writeJSON(w, http.StatusCreated, map[string]any{"event": event})
}

func (s *server) publishTicketEvent(subject string, ticket store.Ticket, action string) {
	if s.bus == nil || !s.bus.Enabled() {
		return
	}
	_ = s.bus.PublishJSON(subject, map[string]any{
		"event":                  subject,
		"tenant_id":              ticket.TenantID,
		"ticket_id":              ticket.ID,
		"conversation_record_id": ticket.ConversationRecordID,
		"call_id":                ticket.CallID,
		"status":                 ticket.Status,
		"action":                 action,
		"at":                     time.Now().UTC(),
	})
}

func ticketPublicJSON(ticket store.Ticket) map[string]any {
	return map[string]any{
		"id":                     ticket.ID,
		"status":                 ticket.Status,
		"priority":               ticket.Priority,
		"category":               ticket.Category,
		"source":                 ticket.Source,
		"call_id":                ticket.CallID,
		"conversation_record_id": ticket.ConversationRecordID,
		"created_at":             ticket.CreatedAt,
	}
}

func ticketListJSON(ticket store.Ticket) map[string]any {
	label := ticket.ContactName
	if label == "" {
		label = "Anonymous"
	}
	return map[string]any{
		"id":               ticket.ID,
		"subject":          ticket.Subject,
		"category":         ticket.Category,
		"priority":         ticket.Priority,
		"status":           ticket.Status,
		"customer_id":      ticket.CustomerID,
		"customer_label":   label,
		"avatar_id":        ticket.AvatarID,
		"avatar_name":      ticket.AvatarName,
		"source":           ticket.Source,
		"call_id":          ticket.CallID,
		"assignee_user_id": ticket.AssigneeUserID,
		"last_activity_at": ticket.LastActivityAt,
	}
}

func ticketTenantJSON(ticket store.Ticket) map[string]any {
	row := ticketListJSON(ticket)
	row["description"] = ticket.Description
	row["call_id"] = ticket.CallID
	row["conversation_record_id"] = ticket.ConversationRecordID
	row["source_summary"] = ticket.SourceSummary
	row["contact_email_masked"] = maskEmail(ticket.ContactEmail)
	row["resolved_at"] = ticket.ResolvedAt
	row["closed_at"] = ticket.ClosedAt
	return row
}

func ticketSourceSummary(record store.ConversationRecord, customerRequest string) map[string]any {
	result := map[string]any{
		"channel":          record.Channel,
		"status":           record.Status,
		"duration_seconds": record.DurationSeconds,
		"started_at":       record.StartedAt,
		"customer_context": boundedTicketContext(customerRequest),
	}
	if record.AvatarName != "" {
		result["avatar_name"] = record.AvatarName
	}
	if topic, ok := record.Summary["topic"].(string); ok && strings.TrimSpace(topic) != "" {
		result["topic"] = strings.TrimSpace(topic)
	}
	return result
}

func boundedTicketContext(value string) string {
	value = strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
	runes := []rune(value)
	if len(runes) > 500 {
		return string(runes[:500]) + "…"
	}
	return value
}

func validateTicketDates(start, end string) error {
	for name, value := range map[string]string{"start_date": start, "end_date": end} {
		if value == "" {
			continue
		}
		if _, err := time.Parse("2006-01-02", value); err != nil {
			return fmt.Errorf("%s must use YYYY-MM-DD", name)
		}
	}
	if start != "" && end != "" && start > end {
		return fmt.Errorf("start_date must be before or equal to end_date")
	}
	return nil
}

func writeTicketStoreError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrTicketNotFound):
		writeTicketError(w, http.StatusNotFound, "ticket not found", "not_found")
	case errors.Is(err, store.ErrTicketConflict):
		writeTicketError(w, http.StatusConflict, "an open ticket already exists", "ticket_already_open")
	case errors.Is(err, store.ErrTicketIdempotency):
		writeTicketError(w, http.StatusConflict, "idempotency key conflicts with an existing request", "idempotency_conflict")
	case errors.Is(err, store.ErrInvalidTransition):
		writeTicketError(w, http.StatusBadRequest, err.Error(), "validation_error")
	case errors.Is(err, store.ErrInvalidTicket):
		writeTicketError(w, http.StatusBadRequest, "invalid ticket fields", "validation_error")
	default:
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			writeTicketError(w, http.StatusBadRequest, err.Error(), "validation_error")
			return
		}
		writeTicketError(w, http.StatusBadGateway, err.Error(), "tickets_unavailable")
	}
}

func writeTicketError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, map[string]any{"error": message, "code": code})
}
