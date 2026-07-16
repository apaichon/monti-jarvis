package env

import "testing"

func TestMobileConfigDefaults(t *testing.T) {
	for _, key := range []string{
		"MOBILE_CALL_API_ENABLED",
		"MOBILE_WS_MAX_FRAME_BYTES",
		"MOBILE_PUSH_ENABLED",
		"MOBILE_PUSH_PROVIDER",
		"MOBILE_PUSH_TOKEN_TTL",
	} {
		t.Setenv(key, "")
	}
	cfg := Load()
	if cfg.MobileCallAPIEnabled {
		t.Fatal("mobile API must be disabled by default")
	}
	if cfg.MobileWSMaxFrameBytes != 32768 {
		t.Fatalf("unexpected mobile frame default: %d", cfg.MobileWSMaxFrameBytes)
	}
	if cfg.MobilePushEnabled || cfg.MobilePushProvider != "auto" || cfg.MobilePushTokenTTL.String() != "15m0s" {
		t.Fatalf("unexpected mobile push defaults: enabled=%v provider=%q ttl=%s", cfg.MobilePushEnabled, cfg.MobilePushProvider, cfg.MobilePushTokenTTL)
	}
}

func TestMobileConfigCanBeEnabled(t *testing.T) {
	t.Setenv("MOBILE_CALL_API_ENABLED", "true")
	t.Setenv("MOBILE_WS_MAX_FRAME_BYTES", "65536")
	t.Setenv("MOBILE_PUSH_ENABLED", "true")
	t.Setenv("MOBILE_PUSH_PROVIDER", "fcm")
	t.Setenv("MOBILE_PUSH_TOKEN_TTL", "30m")
	cfg := Load()
	if !cfg.MobileCallAPIEnabled || cfg.MobileWSMaxFrameBytes != 65536 || !cfg.MobilePushEnabled || cfg.MobilePushProvider != "fcm" || cfg.MobilePushTokenTTL.String() != "30m0s" {
		t.Fatalf("mobile env configuration was not loaded: %+v", cfg)
	}
}
