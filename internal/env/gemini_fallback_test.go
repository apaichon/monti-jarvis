package env

import "testing"

func TestPlatformGeminiFallbackAllowed(t *testing.T) {
	prod := Config{AppEnv: "production", AllowPlatformGeminiFallback: true}
	if prod.PlatformGeminiFallbackAllowed() {
		t.Fatal("production must never allow platform Gemini fallback")
	}
	devOn := Config{AppEnv: "dev", AllowPlatformGeminiFallback: true}
	if !devOn.PlatformGeminiFallbackAllowed() {
		t.Fatal("dev with flag should allow fallback")
	}
	devOff := Config{AppEnv: "dev", AllowPlatformGeminiFallback: false}
	if devOff.PlatformGeminiFallbackAllowed() {
		t.Fatal("dev with flag false should not allow fallback")
	}
}
