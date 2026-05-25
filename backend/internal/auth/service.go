package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/config"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/users"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const minPasswordLength = 12

type AccessClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

type Service struct {
	cfg       config.Config
	repo      *Repository
	userRepo  *users.Repository
	now       func() time.Time
	randomRaw func(int) ([]byte, error)
}

func NewService(cfg config.Config, repo *Repository, userRepo *users.Repository) *Service {
	return &Service{
		cfg:      cfg,
		repo:     repo,
		userRepo: userRepo,
		now: func() time.Time {
			return time.Now().UTC()
		},
		randomRaw: randomBytes,
	}
}

func (s *Service) Register(req RegisterRequest) (AuthResponse, error) {
	email := common.NormalizeEmail(req.Email)
	if !common.ValidateEmail(email) {
		return AuthResponse{}, common.BadRequest("Valid email is required")
	}

	if len(req.Password) < minPasswordLength {
		return AuthResponse{}, common.BadRequest("Password must be at least 12 characters")
	}

	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		return AuthResponse{}, common.BadRequest("Display name is required")
	}

	baseCurrency := common.NormalizeCurrency(req.BaseCurrency)
	if !common.ValidateCurrency(baseCurrency) {
		return AuthResponse{}, common.BadRequest("Base currency must be a three-letter uppercase code")
	}

	timezone := strings.TrimSpace(req.Timezone)
	if timezone == "" {
		timezone = "UTC"
	}

	if _, err := s.userRepo.GetByEmail(email); err == nil {
		return AuthResponse{}, common.Conflict("Email is already registered")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthResponse{}, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.cfg.BcryptCost)
	if err != nil {
		return AuthResponse{}, err
	}

	user := &users.User{
		Email:        email,
		PasswordHash: string(passwordHash),
		DisplayName:  displayName,
		BaseCurrency: baseCurrency,
		Timezone:     timezone,
	}

	if err := s.userRepo.Create(user); err != nil {
		if common.IsUniqueViolation(err) {
			return AuthResponse{}, common.Conflict("Email is already registered")
		}
		return AuthResponse{}, err
	}

	return s.issueTokens(*user)
}

func (s *Service) Login(req LoginRequest) (AuthResponse, error) {
	email := common.NormalizeEmail(req.Email)
	user, err := s.userRepo.GetByEmail(email)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthResponse{}, common.Unauthorized("Invalid email or password")
	}
	if err != nil {
		return AuthResponse{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{}, common.Unauthorized("Invalid email or password")
	}

	return s.issueTokens(*user)
}

func (s *Service) Refresh(refreshToken string) (AuthResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return AuthResponse{}, common.BadRequest("Refresh token is required")
	}

	session, err := s.repo.GetSessionByHash(hashRefreshToken(refreshToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthResponse{}, common.Unauthorized("Invalid refresh token")
	}
	if err != nil {
		return AuthResponse{}, err
	}
	if session.RevokedAt != nil || session.ExpiresAt.Before(s.now()) {
		return AuthResponse{}, common.Unauthorized("Invalid refresh token")
	}

	user, err := s.userRepo.GetByID(session.UserID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthResponse{}, common.Unauthorized("Invalid refresh token")
	}
	if err != nil {
		return AuthResponse{}, err
	}

	if err := s.repo.RevokeSession(session.ID); err != nil {
		return AuthResponse{}, err
	}

	return s.issueTokens(*user)
}

func (s *Service) Logout(refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return common.BadRequest("Refresh token is required")
	}

	session, err := s.repo.GetSessionByHash(hashRefreshToken(refreshToken))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}

	return s.repo.RevokeSession(session.ID)
}

func (s *Service) ParseAccessToken(token string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	parsed, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return []byte(s.cfg.JWTSecret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid || claims.UserID == uuid.Nil {
		return nil, common.Unauthorized("Invalid bearer token")
	}
	return claims, nil
}

func (s *Service) VerifyAccessToken(token string) (uuid.UUID, string, error) {
	claims, err := s.ParseAccessToken(token)
	if err != nil {
		return uuid.Nil, "", err
	}
	return claims.UserID, claims.Email, nil
}

func (s *Service) issueTokens(user users.User) (AuthResponse, error) {
	now := s.now()
	accessExpiresAt := now.Add(s.cfg.AccessTokenTTL)

	claims := AccessClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExpiresAt),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return AuthResponse{}, err
	}

	refreshRaw, err := s.randomRaw(32)
	if err != nil {
		return AuthResponse{}, err
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(refreshRaw)
	session := &Session{
		UserID:           user.ID,
		RefreshTokenHash: hashRefreshToken(refreshToken),
		ExpiresAt:        now.Add(s.cfg.RefreshTokenTTL),
	}
	if err := s.repo.CreateSession(session); err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.cfg.AccessTokenTTL.Seconds()),
		User:         users.ToResponse(user),
	}, nil
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomBytes(size int) ([]byte, error) {
	buf := make([]byte, size)
	_, err := rand.Read(buf)
	return buf, err
}
