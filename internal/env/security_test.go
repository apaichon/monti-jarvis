package env

import "testing"

func TestValidateProductionSecurityRequiresCapabilitySeparation(t *testing.T) {
	cfg := Config{
		AppEnv:                 "production",
		JWTSecret:              "12345678901234567890123456789012",
		PostgresURL:            "postgres://writer:secret@writer-db/monti",
		PostgresKMReadURL:      "postgres://km-read:secret@read-db/monti",
		PostgresTicketWriteURL: "postgres://ticket-write:secret@ticket-db/monti",
		CookieSecure:           true,
		CookieSameSite:         "lax",
		AllowedOrigins:         []string{"https://tenant.example"},
		PostgresRLSEnforced:    true,
	}
	if err := cfg.ValidateProductionSecurity(); err != nil {
		t.Fatalf("valid production config rejected: %v", err)
	}

	cases := []struct {
		name string
		mut  func(*Config)
	}{
		{name: "auth disabled", mut: func(c *Config) { c.AuthDisabled = true }},
		{name: "weak jwt", mut: func(c *Config) { c.JWTSecret = "short" }},
		{name: "missing km read", mut: func(c *Config) { c.PostgresKMReadURL = "" }},
		{name: "missing ticket write", mut: func(c *Config) { c.PostgresTicketWriteURL = "" }},
		{name: "same writer and km", mut: func(c *Config) { c.PostgresKMReadURL = c.PostgresURL }},
		{name: "same km and ticket", mut: func(c *Config) { c.PostgresTicketWriteURL = c.PostgresKMReadURL }},
		{name: "same user different URLs", mut: func(c *Config) {
			c.PostgresKMReadURL = "postgres://writer:other-password@km-read/db"
		}},
		{name: "insecure cookie", mut: func(c *Config) { c.CookieSecure = false }},
		{name: "wildcard origin", mut: func(c *Config) { c.AllowedOrigins = []string{"*"} }},
		{name: "RLS disabled", mut: func(c *Config) { c.PostgresRLSEnforced = false }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := cfg
			tc.mut(&candidate)
			if err := candidate.ValidateProductionSecurity(); err == nil {
				t.Fatal("expected production security validation error")
			}
		})
	}
}

func TestValidateProductionSecurityAllowsDevelopmentFallback(t *testing.T) {
	for _, appEnv := range []string{"", "dev", "test"} {
		if err := (Config{AppEnv: appEnv, AuthDisabled: true}).ValidateProductionSecurity(); err != nil {
			t.Fatalf("app env %q rejected: %v", appEnv, err)
		}
	}
}
