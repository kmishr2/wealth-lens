package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const developmentJWTSecret = "development-only-change-me"

type Config struct {
	AppEnv            string
	HTTPAddr          string
	DatabaseURL       string
	JWTSecret         string
	AccessTokenTTL    time.Duration
	RefreshTokenTTL   time.Duration
	BcryptCost        int
	TrustedProxies    []string
	AuthRateLimit     int
	AuthRateWindow    time.Duration
	AMFINAVURL        string
	UpstoxAPIURL      string
	UpstoxAccessToken string
}

func Load() Config {
	return Config{
		AppEnv:            getEnv("APP_ENV", "development"),
		HTTPAddr:          getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://wealth_lens:wealth_lens@localhost:5432/wealth_lens?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", developmentJWTSecret),
		AccessTokenTTL:    time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL:   time.Duration(getEnvInt("REFRESH_TOKEN_TTL_HOURS", 720)) * time.Hour,
		BcryptCost:        getEnvInt("BCRYPT_COST", 12),
		TrustedProxies:    getEnvList("TRUSTED_PROXIES"),
		AuthRateLimit:     getEnvInt("AUTH_RATE_LIMIT", 20),
		AuthRateWindow:    time.Duration(getEnvInt("AUTH_RATE_WINDOW_SECONDS", 60)) * time.Second,
		AMFINAVURL:        getEnv("AMFI_NAV_URL", "https://portal.amfiindia.com/spages/NAVAll.txt"),
		UpstoxAPIURL:      getEnv("UPSTOX_API_URL", "https://api.upstox.com/v3/historical-candle"),
		UpstoxAccessToken: os.Getenv("UPSTOX_ACCESS_TOKEN"),
	}
}

func (c Config) Validate() error {
	if c.AppEnv == "production" && (c.JWTSecret == developmentJWTSecret || len(c.JWTSecret) < 32) {
		return fmt.Errorf("JWT_SECRET must be a unique secret of at least 32 characters in production")
	}
	if c.AccessTokenTTL <= 0 || c.RefreshTokenTTL <= 0 {
		return fmt.Errorf("token TTL values must be positive")
	}
	if c.BcryptCost < 10 || c.BcryptCost > 15 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 15")
	}
	if c.AuthRateLimit <= 0 || c.AuthRateWindow <= 0 {
		return fmt.Errorf("auth rate limit and window must be positive")
	}
	return nil
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func getEnvList(key string) []string {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		if value := strings.TrimSpace(part); value != "" {
			values = append(values, value)
		}
	}
	return values
}
