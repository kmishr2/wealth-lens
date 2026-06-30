package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv             string
	HTTPAddr           string
	DatabaseURL        string
	JWTSecret          string
	AccessTokenTTL     time.Duration
	RefreshTokenTTL    time.Duration
	BcryptCost         int
	TrustedProxyHeader string
	AMFINAVURL         string
	UpstoxAPIURL       string
	UpstoxAccessToken  string
}

func Load() Config {
	return Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		HTTPAddr:           getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://wealth_lens:wealth_lens@localhost:5432/wealth_lens?sslmode=disable"),
		JWTSecret:          getEnv("JWT_SECRET", "development-only-change-me"),
		AccessTokenTTL:     time.Duration(getEnvInt("ACCESS_TOKEN_TTL_MINUTES", 15)) * time.Minute,
		RefreshTokenTTL:    time.Duration(getEnvInt("REFRESH_TOKEN_TTL_HOURS", 720)) * time.Hour,
		BcryptCost:         getEnvInt("BCRYPT_COST", 12),
		TrustedProxyHeader: getEnv("TRUSTED_PROXY_HEADER", ""),
		AMFINAVURL:         getEnv("AMFI_NAV_URL", "https://portal.amfiindia.com/spages/NAVAll.txt"),
		UpstoxAPIURL:       getEnv("UPSTOX_API_URL", "https://api.upstox.com/v3/historical-candle"),
		UpstoxAccessToken:  os.Getenv("UPSTOX_ACCESS_TOKEN"),
	}
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
