package main

import "testing"

func TestCheckoutReturnURLForcesTenantBillingReturn(t *testing.T) {
	cases := []struct {
		name       string
		configured string
		publicBase string
		orderID    string
		want       string
	}{
		{
			name:       "default from public base",
			configured: "",
			publicBase: "http://localhost:8091",
			orderID:    "ord_abc",
			want:       "http://localhost:8091/tenant/billing/return?order_id=ord_abc",
		},
		{
			name:       "ngrok host kept path forced",
			configured: "https://embellish.ngrok-free.dev/some/old/path",
			publicBase: "http://localhost:8091",
			orderID:    "ord_1",
			want:       "https://embellish.ngrok-free.dev/tenant/billing/return?order_id=ord_1",
		},
		{
			name:       "already correct path",
			configured: "https://pay.example.com/tenant/billing/return",
			publicBase: "http://localhost:8091",
			orderID:    "ord_x",
			want:       "https://pay.example.com/tenant/billing/return?order_id=ord_x",
		},
		{
			name:       "empty public falls back localhost",
			configured: "",
			publicBase: "",
			orderID:    "ord_z",
			want:       "http://localhost:8091/tenant/billing/return?order_id=ord_z",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := checkoutReturnURL(tc.configured, tc.publicBase, tc.orderID)
			if got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestChillpayBrowserReturnURLEmbedsOrderNoInPath(t *testing.T) {
	got := chillpayBrowserReturnURL(
		"https://sardine.ngrok-free.dev/tenant/billing/return",
		"https://sardine.ngrok-free.dev/api/callbacks/chillpay",
		"http://localhost:8091",
		"MJ6b18c14228e58838b0",
	)
	want := "https://sardine.ngrok-free.dev/api/callbacks/chillpay/return/MJ6b18c14228e58838b0"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeChillPayPaymentStatus(t *testing.T) {
	if normalizeChillPayPaymentStatus("complete") != "0" {
		t.Fatal("complete → 0")
	}
	if normalizeChillPayPaymentStatus("0") != "0" {
		t.Fatal("0 → 0")
	}
	if normalizeChillPayPaymentStatus("failed") != "2" {
		t.Fatal("failed → 2")
	}
	if normalizeChillPayPaymentStatus("pending") != "1" {
		t.Fatal("pending → 1")
	}
}

func TestTenantSPAReturnURLWithParams(t *testing.T) {
	got := tenantSPAReturnURL(
		"https://sardine.ngrok-free.dev/tenant/billing/return",
		"http://localhost:8091",
		"ord_1",
		"MJab12cd34ef567890",
		"0",
		"999",
	)
	want := "https://sardine.ngrok-free.dev/tenant/billing/return?order_id=ord_1&order_no=MJab12cd34ef567890&status=0&txn_id=999"
	// url.Values.Encode sorts keys alphabetically
	want = "https://sardine.ngrok-free.dev/tenant/billing/return?order_id=ord_1&order_no=MJab12cd34ef567890&status=0&txn_id=999"
	if got != want {
		// order of query keys is alphabetical from url.Values
		if got != "https://sardine.ngrok-free.dev/tenant/billing/return?order_id=ord_1&order_no=MJab12cd34ef567890&status=0&txn_id=999" {
			t.Fatalf("got %q", got)
		}
	}
}
