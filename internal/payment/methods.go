package payment

import (
	"fmt"
	"strings"
)

// Tenant-facing payment methods for Buy Package checkout.
const (
	MethodCreditCard  = "credit_card"
	MethodQRPromptPay = "qr_promptpay"
)

// ChillPay ChannelCode values (form field ChannelCode on InitPayment).
const (
	ChannelCreditCard  = "creditcard"
	ChannelQRPromptPay = "promptpay"
)

// NormalizePaymentMethod returns a canonical method or an error if unsupported.
func NormalizePaymentMethod(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", MethodCreditCard, "creditcard", "card":
		return MethodCreditCard, nil
	case MethodQRPromptPay, "promptpay", "qr", "qrcode", "qr_code":
		return MethodQRPromptPay, nil
	default:
		return "", fmt.Errorf("unsupported payment_method %q (use credit_card or qr_promptpay)", raw)
	}
}

// ChannelCodeForMethod maps tenant payment_method → ChillPay ChannelCode.
func ChannelCodeForMethod(method string) string {
	switch method {
	case MethodQRPromptPay:
		return ChannelQRPromptPay
	default:
		return ChannelCreditCard
	}
}

// MethodLabel is a human-readable label for documents/UI.
func MethodLabel(method string) string {
	switch method {
	case MethodQRPromptPay:
		return "QR PromptPay"
	case MethodCreditCard:
		return "Credit Card"
	default:
		if method == "" {
			return "Credit Card"
		}
		return method
	}
}
