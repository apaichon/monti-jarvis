package store

import "testing"

func TestNewPromotionOrderNo(t *testing.T) {
	no := newPromotionOrderNo("acme")
	if len(no) != 20 {
		t.Fatalf("order_no length = %d, want 20 (got %q)", len(no), no)
	}
	if no[:2] != "PR" {
		t.Fatalf("prefix = %q, want PR", no[:2])
	}
	for _, c := range no {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F') || c == 'P' || c == 'R') {
			// allow full alphanum; hex id is 0-9a-f
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')) {
				t.Fatalf("non-alphanumeric char %q in %q", c, no)
			}
		}
	}
	// empty tenant still 20 chars
	no2 := newPromotionOrderNo("")
	if len(no2) != 20 {
		t.Fatalf("empty tenant order_no length = %d", len(no2))
	}
}

func TestSplitVATInclusiveZero(t *testing.T) {
	net, vat := splitVATInclusive(0, 700)
	if net != 0 || vat != 0 {
		t.Fatalf("zero amount net=%d vat=%d", net, vat)
	}
	net, vat = splitVATInclusive(10700, 700)
	if net+vat != 10700 {
		t.Fatalf("net+vat = %d, want 10700", net+vat)
	}
}
