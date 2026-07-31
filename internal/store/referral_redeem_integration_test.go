package store

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/libra/monti-jarvis/internal/env"
)

func TestReferralRedemptionIntegration(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("MONTI_INTEGRATION_POSTGRES_URL"))
	if dsn == "" {
		t.Skip("set MONTI_INTEGRATION_POSTGRES_URL to run Postgres integration tests")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	st, warnings := Open(ctx, env.Config{PostgresURL: dsn, PostgresSchema: "callcenter"})
	t.Cleanup(st.Close)
	for _, warning := range warnings {
		if strings.HasPrefix(warning, "postgres ") {
			t.Fatalf("open store: %s", warning)
		}
	}
	if st.pg == nil {
		t.Fatal("Postgres store is unavailable")
	}

	suffix := strings.ReplaceAll(time.Now().UTC().Format("150405.000000000"), ".", "")
	referrerID := "review_referrer_" + suffix
	redeemerID := "review_redeemer_" + suffix
	otherID := "review_other_" + suffix
	code := "review_code_" + suffix
	expiredCode := "review_expired_" + suffix
	schema := quoteIdent(st.cfg.PostgresSchema)
	for _, tenantID := range []string{referrerID, redeemerID, otherID} {
		_, err := st.pg.Exec(ctx, "INSERT INTO "+schema+".tenants (id,slug,name,status,created_by,updated_by) VALUES ($1,$1,$1,'active','test','test')", tenantID)
		if err != nil {
			t.Fatalf("insert tenant %s: %v", tenantID, err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		_, _ = st.pg.Exec(cleanupCtx, "DELETE FROM "+schema+".tenants WHERE id = ANY($1)", []string{referrerID, redeemerID, otherID})
	})
	_, err := st.pg.Exec(ctx, "INSERT INTO "+schema+".tenant_referral_codes (id,tenant_id,code,status,created_by,updated_by) VALUES ($1,$2,$3,'active','test','test')",
		"review_refcode_"+suffix, referrerID, code)
	if err != nil {
		t.Fatalf("insert referral code: %v", err)
	}
	_, err = st.pg.Exec(ctx, "INSERT INTO "+schema+".tenant_referral_codes (id,tenant_id,code,status,expires_at,created_by,updated_by) VALUES ($1,$2,$3,'active',now()-interval '1 day','test','test')",
		"review_expired_refcode_"+suffix, otherID, expiredCode)
	if err != nil {
		t.Fatalf("insert expired referral code: %v", err)
	}

	if _, _, err := st.ValidateReferralCodeForRedeem(ctx, redeemerID, expiredCode); !errors.Is(err, ErrReferralIneligible) {
		t.Fatalf("expired code error = %v, want ineligible", err)
	}
	if _, _, err := st.ValidateReferralCodeForRedeem(ctx, referrerID, code); !errors.Is(err, ErrReferralRedeemSelf) {
		t.Fatalf("self-referral error = %v, want self referral", err)
	}

	first, err := st.RedeemReferralCode(ctx, redeemerID, code, "review-request")
	if err != nil {
		t.Fatalf("redeem code: %v", err)
	}
	if first.ID == "" || len(first.Bonus) == 0 {
		t.Fatalf("redemption missing id or bonus: %+v", first)
	}
	second, err := st.RedeemReferralCode(ctx, redeemerID, code, "review-request-retry")
	if err != nil {
		t.Fatalf("retry redemption: %v", err)
	}
	if second.ID != first.ID {
		t.Fatalf("retry id = %q, want %q", second.ID, first.ID)
	}
	otherRows, err := st.ListTenantRedemptions(ctx, otherID)
	if err != nil {
		t.Fatalf("list isolated tenant redemptions: %v", err)
	}
	if len(otherRows) != 0 {
		t.Fatalf("cross-tenant redemption leak: %+v", otherRows)
	}
	platformRows, err := st.ListPlatformReferralRedemptions(ctx, PlatformReferralRedemptionFilter{TenantID: redeemerID})
	if err != nil || len(platformRows) != 1 || platformRows[0].ID != first.ID {
		t.Fatalf("platform redemption list = %+v, err=%v", platformRows, err)
	}

	reversed, err := st.ReverseReferralRedemption(ctx, first.ID, "integration review")
	if err != nil {
		t.Fatalf("reverse redemption: %v", err)
	}
	if reversed.Status != "reversed" {
		t.Fatalf("reversed status = %q", reversed.Status)
	}
	reversedAgain, err := st.ReverseReferralRedemption(ctx, first.ID, "integration review retry")
	if err != nil || reversedAgain.Status != "reversed" {
		t.Fatalf("idempotent reverse = %+v, err=%v", reversedAgain, err)
	}
}
