package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/customerimport"
	"github.com/libra/monti-jarvis/internal/store"
)

func TestWriteCustomerError(t *testing.T) {
	for _, tc := range []struct {
		err  error
		want int
		code string
	}{
		{store.ErrCustomerNotFound, http.StatusNotFound, "not_found"},
		{store.ErrCustomerConflict, http.StatusConflict, "customer_conflict"},
		{store.ErrDomainRuleTaken, http.StatusConflict, "domain_rule_exists"},
	} {
		rr := httptest.NewRecorder()
		writeCustomerError(rr, tc.err)
		if rr.Code != tc.want {
			t.Fatalf("%v status=%d body=%s", tc.err, rr.Code, rr.Body.String())
		}
		var body map[string]any
		_ = json.NewDecoder(rr.Body).Decode(&body)
		if body["code"] != tc.code {
			t.Fatalf("%v body=%#v", tc.err, body)
		}
	}
}

func TestRejectedRowsDeduplicatesMultipleFieldErrors(t *testing.T) {
	errs := []customerimport.RowError{{Row: 2}, {Row: 2}, {Row: 4}}
	if got := rejectedRows(errs); got != 2 {
		t.Fatalf("got %d", got)
	}
}
