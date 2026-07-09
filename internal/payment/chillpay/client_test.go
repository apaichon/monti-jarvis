package chillpay

import "testing"

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
}