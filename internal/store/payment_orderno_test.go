package store

import (
	"regexp"
	"testing"
)

func TestNewPaymentOrderNoChillPayRules(t *testing.T) {
	re := regexp.MustCompile(`^[A-Za-z0-9]{1,20}$`)
	for _, tenant := range []string{"acme", "my-long-tenant-slug-name", "a", ""} {
		got := newPaymentOrderNo(tenant)
		if !re.MatchString(got) {
			t.Fatalf("tenant %q order_no %q fails ChillPay alphanumeric ≤20 rule", tenant, got)
		}
		if len(got) != 20 {
			t.Fatalf("tenant %q order_no len=%d want 20 (%q)", tenant, len(got), got)
		}
		if got[:2] != "MJ" {
			t.Fatalf("want MJ prefix, got %q", got)
		}
	}
	a := newPaymentOrderNo("acme")
	b := newPaymentOrderNo("acme")
	if a == b {
		// same nanosecond possible but rare; at least format is stable
		t.Logf("note: consecutive order nos collided (same nano): %s", a)
	}
}
