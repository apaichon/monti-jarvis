package store

import (
	"errors"
	"testing"
	"time"
)

func commercialTestPackage(mode string) Package {
	deployment := DeploymentSharedCloud
	if mode == PurchaseModeQuote {
		deployment = DeploymentDedicatedVM
	}
	return Package{
		ID: "pkg-test", Slug: "test", Name: "Test plan", Status: "active",
		PriceCents: 50000, Currency: "THB", BillingPeriod: "monthly",
		DeploymentMode: deployment, PurchaseMode: mode,
		Rules: map[string]any{"max_concurrent_calls": float64(4)},
	}
}

func commercialTestVersion() CatalogVersion {
	return CatalogVersion{
		ID: "pkgv-test-v1", PackageID: "pkg-test", Version: 1,
		MonthlyPriceCents: 50000, AnnualPriceCents: 480000,
		AnnualDiscountBps: 2000, Currency: "THB", TaxRateBps: 700,
		RulesSnapshot: map[string]any{"max_concurrent_calls": float64(4)},
		EffectiveFrom: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		Status:        "active",
	}
}

func TestCalculateCatalogPriceMonthly(t *testing.T) {
	now := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	got, err := CalculateCatalogPrice(commercialTestPackage(PurchaseModeSelfServe), commercialTestVersion(), BillingIntervalMonthly, now)
	if err != nil {
		t.Fatalf("CalculateCatalogPrice: %v", err)
	}
	if got.BasePriceCents != 50000 || got.DiscountCents != 0 || got.TaxCents != 3500 || got.AmountDueCents != 53500 {
		t.Fatalf("unexpected monthly calculation: %+v", got)
	}
	if !got.CheckoutEligible || got.QuoteRequired {
		t.Fatalf("shared package routing is wrong: %+v", got)
	}
}

func TestCalculateCatalogPriceAnnual(t *testing.T) {
	got, err := CalculateCatalogPrice(
		commercialTestPackage(PurchaseModeSelfServe),
		commercialTestVersion(),
		BillingIntervalAnnual,
		time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("CalculateCatalogPrice: %v", err)
	}
	if got.BasePriceCents != 600000 || got.DiscountCents != 120000 ||
		got.TaxableAmountCents != 480000 || got.TaxCents != 33600 || got.AmountDueCents != 513600 {
		t.Fatalf("unexpected annual calculation: %+v", got)
	}
}

func TestCalculateCatalogPriceDedicatedIsIndicativeOnly(t *testing.T) {
	got, err := CalculateCatalogPrice(
		commercialTestPackage(PurchaseModeQuote),
		commercialTestVersion(),
		BillingIntervalMonthly,
		time.Now().UTC(),
	)
	if err != nil {
		t.Fatalf("CalculateCatalogPrice: %v", err)
	}
	if got.CheckoutEligible || !got.QuoteRequired || got.DeploymentMode != DeploymentDedicatedVM {
		t.Fatalf("dedicated package must be quote-only: %+v", got)
	}
}

func TestCalculateCatalogPriceRejectsBrowserIntervalAndMismatchedVersion(t *testing.T) {
	_, err := CalculateCatalogPrice(
		commercialTestPackage(PurchaseModeSelfServe),
		commercialTestVersion(),
		"weekly",
		time.Now().UTC(),
	)
	if !errors.Is(err, ErrInvalidCommercialRequest) {
		t.Fatalf("expected invalid commercial request, got %v", err)
	}
	version := commercialTestVersion()
	version.PackageID = "pkg-other"
	_, err = CalculateCatalogPrice(
		commercialTestPackage(PurchaseModeSelfServe),
		version,
		BillingIntervalMonthly,
		time.Now().UTC(),
	)
	if !errors.Is(err, ErrCatalogVersionNotFound) {
		t.Fatalf("expected catalog mismatch, got %v", err)
	}
}

func TestNormalizeDedicatedQuoteInput(t *testing.T) {
	input, err := normalizeDedicatedQuoteInput(CreateDedicatedQuoteInput{
		TenantID: " tenant-a ", PackageID: " pkg-dedicated ",
		CompanyLegalName: " Example Co., Ltd. ", ContactName: " Ada ",
		ContactEmail: " ADMIN@EXAMPLE.COM ", ContactPhone: " 0800000000 ",
		CompanySize: "11-50", ExpectedConcurrency: 25, PreferredRegion: "th-bangkok",
	})
	if err != nil {
		t.Fatalf("normalizeDedicatedQuoteInput: %v", err)
	}
	if input.TenantID != "tenant-a" || input.ContactEmail != "admin@example.com" {
		t.Fatalf("input was not normalized: %+v", input)
	}

	input.ExpectedConcurrency = 0
	if _, err := normalizeDedicatedQuoteInput(input); !errors.Is(err, ErrInvalidCommercialRequest) {
		t.Fatalf("expected concurrency validation, got %v", err)
	}
}

func TestDedicatedQuoteTransitions(t *testing.T) {
	valid := [][2]string{
		{QuoteSubmitted, QuoteUnderReview},
		{QuoteUnderReview, QuoteCapacityConfirmed},
		{QuoteCapacityConfirmed, QuoteQuoted},
		{QuoteQuoted, QuoteAccepted},
		{QuoteAccepted, QuoteProvisioning},
		{QuoteProvisioning, QuoteActive},
	}
	for _, transition := range valid {
		if !validQuoteTransition(transition[0], transition[1]) {
			t.Errorf("expected transition %s -> %s", transition[0], transition[1])
		}
	}
	if validQuoteTransition(QuoteSubmitted, QuoteActive) {
		t.Fatal("submitted quote must not jump directly to active")
	}
	for _, currency := range []string{"THB", "764", "US_DOLLAR", "thb", ""} {
		want := currency == "THB" || currency == "764"
		if got := validCommercialCurrency(currency); got != want {
			t.Fatalf("validCommercialCurrency(%q)=%v, want %v", currency, got, want)
		}
	}
}

func TestBillingIntervalPeriodAndRetryBounds(t *testing.T) {
	start := time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)
	if got := addBillingInterval(start, BillingIntervalMonthly); got.Month() != time.February || got.Day() != 28 {
		t.Fatalf("expected end-of-month clamping, got %s", got)
	}
	if got := addBillingInterval(start, BillingIntervalAnnual); got.Year() != 2027 {
		t.Fatalf("unexpected annual boundary: %s", got)
	}
	leapDay := time.Date(2024, 2, 29, 10, 0, 0, 0, time.UTC)
	if got := addBillingInterval(leapDay, BillingIntervalAnnual); got.Year() != 2025 || got.Month() != time.February || got.Day() != 28 {
		t.Fatalf("expected leap-day clamping, got %s", got)
	}
	delays := normalizeRetryDelays([]time.Duration{-1, time.Hour, 6 * time.Hour})
	if len(delays) != 2 || delays[0] != time.Hour || delays[1] != 6*time.Hour {
		t.Fatalf("unexpected retry delays: %v", delays)
	}
}
