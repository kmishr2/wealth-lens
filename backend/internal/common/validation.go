package common

import (
	"regexp"
	"strings"
	"time"
)

var (
	currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)
	emailPattern    = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
)

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func NormalizeSymbol(symbol string) string {
	return strings.ToUpper(strings.TrimSpace(symbol))
}

func NormalizeCurrency(currency string) string {
	return strings.ToUpper(strings.TrimSpace(currency))
}

func ValidateEmail(email string) bool {
	return emailPattern.MatchString(email)
}

func ValidateCurrency(currency string) bool {
	return currencyPattern.MatchString(currency)
}

func ValidateNotBlank(value string) bool {
	return strings.TrimSpace(value) != ""
}

func ValidateNotFuture(t time.Time) bool {
	return !t.After(time.Now().UTC())
}

func OneOf(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
