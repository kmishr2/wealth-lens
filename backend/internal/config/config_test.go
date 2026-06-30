package config

import (
	"strings"
	"testing"
	"time"
)

func validConfig() Config {
	return Config{
		AppEnv: "development", JWTSecret: developmentJWTSecret,
		AccessTokenTTL: time.Minute, RefreshTokenTTL: time.Hour, BcryptCost: 12,
		AuthRateLimit: 20, AuthRateWindow: time.Minute,
	}
}

func TestValidateRejectsUnsafeProductionSecret(t *testing.T) {
	cfg := validConfig()
	cfg.AppEnv = "production"
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Fatalf("Validate error = %v", err)
	}
	cfg.JWTSecret = "a-unique-production-secret-with-at-least-32-characters"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid production config: %v", err)
	}
}

func TestValidateRejectsInvalidSecuritySettings(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Config)
	}{
		{"bcrypt cost", func(c *Config) { c.BcryptCost = 9 }},
		{"access TTL", func(c *Config) { c.AccessTokenTTL = 0 }},
		{"rate limit", func(c *Config) { c.AuthRateLimit = 0 }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig()
			test.mutate(&cfg)
			if err := cfg.Validate(); err == nil {
				t.Fatal("Validate returned nil")
			}
		})
	}
}
