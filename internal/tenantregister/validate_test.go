package tenantregister

import "testing"

func TestValidateSlug(t *testing.T) {
	ok := Input{
		CompanyName:      "Acme Corp",
		Slug:             "acme-corp",
		AdminEmail:       "a@acme.test",
		AdminPassword:    "secret1234",
		AdminDisplayName: "Admin",
	}
	if err := Validate(ok); err != nil {
		t.Fatalf("expected valid: %v", err)
	}

	bad := ok
	bad.Slug = "Demo"
	if err := Validate(bad); err == nil {
		t.Fatal("expected reserved/invalid slug error")
	}

	short := ok
	short.AdminPassword = "short"
	if err := Validate(short); err == nil {
		t.Fatal("expected password error")
	}
}