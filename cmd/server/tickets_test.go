package main

import "testing"

func TestTicketOfferForMessage(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		topic    string
		category string
	}{
		{name: "english", message: "Please connect me to a human agent", topic: "general", category: "general"},
		{name: "topic", message: "I want to speak to a real person", topic: "billing", category: "billing"},
		{name: "thai", message: "ขอคุยกับเจ้าหน้าที่", topic: "technical", category: "technical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offer := ticketOfferForMessage(tt.message, tt.topic)
			if offer == nil {
				t.Fatal("expected ticket offer")
			}
			if offer.Category != tt.category || offer.Subject == "" || offer.Reason == "" {
				t.Fatalf("unexpected offer: %+v", offer)
			}
		})
	}
}

func TestTicketOfferForMessageIgnoresNormalRequest(t *testing.T) {
	if offer := ticketOfferForMessage("How do I update my billing address?", "billing"); offer != nil {
		t.Fatalf("unexpected ticket offer: %+v", offer)
	}
}

func TestValidateTicketDates(t *testing.T) {
	if err := validateTicketDates("2026-07-14", "2026-07-14"); err != nil {
		t.Fatalf("same-day range should be valid: %v", err)
	}
	if err := validateTicketDates("2026-07-15", "2026-07-14"); err == nil {
		t.Fatal("expected reversed range to fail")
	}
	if err := validateTicketDates("14-07-2026", "2026-07-14"); err == nil {
		t.Fatal("expected invalid date to fail")
	}
}
