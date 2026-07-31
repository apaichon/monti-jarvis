package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/store"
)

func TestWriteReferralRedeemError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "not found", err: store.ErrReferralNotFound, wantStatus: http.StatusNotFound, wantCode: "referral_not_found"},
		{name: "self", err: store.ErrReferralRedeemSelf, wantStatus: http.StatusConflict, wantCode: "self_referral"},
		{name: "duplicate", err: store.ErrReferralAlreadyRedeemed, wantStatus: http.StatusConflict, wantCode: "already_redeemed"},
		{name: "ineligible", err: store.ErrReferralIneligible, wantStatus: http.StatusUnprocessableEntity, wantCode: "referral_ineligible"},
		{name: "invalid", err: store.ErrReferralInvalid, wantStatus: http.StatusBadRequest, wantCode: "validation_error"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			writeReferralRedeemError(recorder, tt.err)
			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			var body map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if body["code"] != tt.wantCode {
				t.Fatalf("code = %v, want %q", body["code"], tt.wantCode)
			}
		})
	}
}
