package store

import "testing"

func TestReferralCodeRedeemable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		codeStatus  string
		expired     bool
		ownerStatus string
		want        bool
	}{
		{name: "active code and tenant", codeStatus: ReferralCodeActive, ownerStatus: "active", want: true},
		{name: "expired", codeStatus: ReferralCodeActive, expired: true, ownerStatus: "active"},
		{name: "disabled code", codeStatus: ReferralCodeDisabled, ownerStatus: "active"},
		{name: "inactive owner", codeStatus: ReferralCodeActive, ownerStatus: "suspended"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := referralCodeRedeemable(tt.codeStatus, tt.expired, tt.ownerStatus); got != tt.want {
				t.Fatalf("referralCodeRedeemable() = %v, want %v", got, tt.want)
			}
		})
	}
}
