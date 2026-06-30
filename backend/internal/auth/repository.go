package auth

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateSession(session *Session) error {
	return r.db.Create(session).Error
}

func (r *Repository) GetSessionByHash(hash string) (*Session, error) {
	var session Session
	err := r.db.First(&session, "refresh_token_hash = ?", hash).Error
	return &session, err
}

func (r *Repository) RevokeSession(id uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.Model(&Session{}).Where("id = ?", id).Update("revoked_at", now).Error
}

// ConsumeSessionByHash atomically revokes and returns one active session. The
// conditional UPDATE makes refresh tokens single-use under concurrent calls.
func (r *Repository) ConsumeSessionByHash(hash string, now time.Time) (*Session, error) {
	var session Session
	result := r.db.Raw(`
		UPDATE auth_sessions
		SET revoked_at = ?
		WHERE refresh_token_hash = ?
			AND revoked_at IS NULL
			AND expires_at > ?
		RETURNING id, user_id, refresh_token_hash, expires_at, revoked_at, created_at
	`, now, hash, now).Scan(&session)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &session, nil
}
