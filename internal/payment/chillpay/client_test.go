package chillpay

import (
	"strings"
	"testing"
)

func TestSanitizeOrderNo(t *testing.T) {
	// Legacy format with underscores + tenant — must become pure alnum ≤20.
	got := SanitizeOrderNo("mj_acme_a1b2c3d4e5f67890")
	if got != "mjacmea1b2c3d4e5f678" {
		t.Fatalf("got %q want mjacmea1b2c3d4e5f678", got)
	}
	if SanitizeOrderNo("MJ00abcdef0123456789") != "MJ00abcdef0123456789" {
		t.Fatalf("valid order mutated: %q", SanitizeOrderNo("MJ00abcdef0123456789"))
	}
	if SanitizeOrderNo("") != "" {
		t.Fatal("empty should stay empty")
	}
}

func TestSanitizeCustomerID(t *testing.T) {
	got := SanitizeCustomerID("my-tenant_slug")
	if got != "mytenantslug" {
		t.Fatalf("got %q", got)
	}
}

func TestSanitizeCustName(t *testing.T) {
	// Email must never be sent as CustName (ChillPay 2032).
	if got := SanitizeCustName("user@gmail.com", "user@gmail.com"); got == "user@gmail.com" || strings.Contains(got, "@") {
		t.Fatalf("email leaked into CustName: %q", got)
	}
	if got := SanitizeCustName("", "john.doe@acme.test"); got != "john doe" {
		t.Fatalf("email local-part fallback: got %q want john doe", got)
	}
	if got := SanitizeCustName("สมชาย ใจดี", "a@b.com"); got != "สมชาย ใจดี" {
		t.Fatalf("thai name: got %q", got)
	}
	if got := SanitizeCustName("Jane Doe", "x@y.com"); got != "Jane Doe" {
		t.Fatalf("display name: got %q", got)
	}
	if got := SanitizeCustName("", ""); got != "Customer" {
		t.Fatalf("empty fallback: got %q", got)
	}
}

func TestVerifyCallback(t *testing.T) {
	c := NewClient(Config{
		MerchantCode: "M1",
		APIKey:       "key",
		MD5Key:       "secret",
		BaseURL:      "https://example.test/Payment",
		RouteNo:      1,
		Currency:     "764",
	})

	form := CallbackForm{
		TransactionId:      "999",
		Amount:             "10000",
		OrderNo:            "ord-1",
		CustomerId:         "cust-1",
		BankCode:           "BBL",
		PaymentDate:        "20260709",
		PaymentStatus:      "0",
		BankRefCode:        "ref",
		CurrentDate:        "20260709",
		CurrentTime:        "120000",
		PaymentDescription: "pkg",
		CreditCardToken:    "",
		Currency:           "764",
		CustomerName:       "Jane",
	}
	raw := form.TransactionId + form.Amount + form.OrderNo + form.CustomerId +
		form.BankCode + form.PaymentDate + form.PaymentStatus + form.BankRefCode +
		form.CurrentDate + form.CurrentTime + form.PaymentDescription +
		form.CreditCardToken + form.Currency + form.CustomerName + "secret"
	form.CheckSum = md5Hex(raw)

	if !c.VerifyCallback(form) {
		t.Fatal("expected valid checksum")
	}
	form.CheckSum = "bad"
	if c.VerifyCallback(form) {
		t.Fatal("expected invalid checksum")
	}
}

func TestStatusEndpoint(t *testing.T) {
	got := statusEndpoint("https://sandbox-appsrv2.chillpay.co/api/v2/Payment")
	want := "https://sandbox-appsrv2.chillpay.co/api/v2/PaymentStatus/"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	// harvest-core also trims trailing /Payment/
	got2 := statusEndpoint("https://sandbox.example/api/v2/Payment/")
	if !strings.HasSuffix(got2, "/api/v2/PaymentStatus/") {
		t.Fatalf("trailing Payment/ not stripped: %q", got2)
	}
}

// TestInitPaymentChecksumOrder documents the harvest-core param order (1–20 + MD5 key).
func TestInitPaymentChecksumOrder(t *testing.T) {
	// Same concatenation as harvest-core chillpay.go InitPayment.
	merchantCode := "M0001"
	orderNo := "MJ00abcdef01234567"
	customerID := "tenant1"
	amountStr := "200000"
	phoneNumber := ""
	description := "Pro"
	channelCode := "creditcard"
	currency := "764"
	langCode := "TH"
	routeNoStr := "1"
	ipAddress := "127.0.0.1"
	apiKey := "apikey"
	tokenFlag := "N"
	creditToken, creditMonth, shopID, productImageUrl, cardType := "", "", "", "", ""
	custEmail := "a@b.com"
	custName := "Jane Doe"
	md5Key := "secret"

	raw := merchantCode + orderNo + customerID + amountStr + phoneNumber + description +
		channelCode + currency + langCode + routeNoStr + ipAddress + apiKey +
		tokenFlag + creditToken + creditMonth + shopID + productImageUrl +
		custEmail + cardType + custName + md5Key
	if md5Hex(raw) == "" {
		t.Fatal("empty checksum")
	}
	// Ensure CustName is after CardType (not omitted) — harvest includes param 20.
	if !strings.HasSuffix(raw, custName+md5Key) {
		t.Fatal("CustName must be last field before MD5 key")
	}
}