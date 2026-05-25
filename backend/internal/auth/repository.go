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
