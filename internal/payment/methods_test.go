package payment

import "testing"

func TestNormalizePaymentMethod(t *testing.T) {
	cases := []struct {
		in   string
		want string
		ok   bool
	}{
		{"", MethodCreditCard, true},
		{"credit_card", MethodCreditCard, true},
		{"CreditCard", MethodCreditCard, true},
		{"qr_promptpay", MethodQRPromptPay, true},
		{"promptpay", MethodQRPromptPay, true},
		{"qr", MethodQRPromptPay, true},
		{"bitcoin", "", false},
	}
	for _, tc := range cases {
		got, err := NormalizePaymentMethod(tc.in)
		if tc.ok {
			if err != nil {
				t.Fatalf("in %q: unexpected err %v", tc.in, err)
			}
			if got != tc.want {
				t.Fatalf("in %q: got %q want %q", tc.in, got, tc.want)
			}
		} else if err == nil {
			t.Fatalf("in %q: expected error", tc.in)
		}
	}
}

func TestChannelCodeForMethod(t *testing.T) {
	if ChannelCodeForMethod(MethodCreditCard) != ChannelCreditCard {
		t.Fatal("credit card channel")
	}
	if ChannelCodeForMethod(MethodQRPromptPay) != ChannelQRPromptPay {
		t.Fatal("promptpay channel")
	}
}
