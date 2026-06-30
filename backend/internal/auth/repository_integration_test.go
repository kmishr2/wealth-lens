//go:build integration

package auth_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/auth"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/database"
	"gorm.io/gorm"
)

func TestRefreshSessionCanOnlyBeConsumedOnce(t *testing.T) {
	databaseURL := os.Getenv("BACKEND_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("BACKEND_TEST_DATABASE_URL is required")
	}
	db, err := database.Connect(databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	userID := uuid.New()
	if err := db.Exec(`INSERT INTO users (id, email, password_hash, display_name, base_currency, timezone)
		VALUES (?, ?, 'unused', 'Auth Test', 'INR', 'Asia/Kolkata')`, userID, userID.String()+"@example.com").Error; err != nil {
		t.Fatal(err)
	}
	repo := auth.NewRepository(db)
	session := &auth.Session{UserID: userID, RefreshTokenHash: uuid.NewString(), ExpiresAt: time.Now().UTC().Add(time.Hour)}
	if err := repo.CreateSession(session); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	consumed, err := repo.ConsumeSessionByHash(session.RefreshTokenHash, now)
	if err != nil || consumed.ID != session.ID || consumed.RevokedAt == nil {
		t.Fatalf("first consume: session=%+v err=%v", consumed, err)
	}
	if _, err := repo.ConsumeSessionByHash(session.RefreshTokenHash, now); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("second consume error = %v, want record not found", err)
	}
}
