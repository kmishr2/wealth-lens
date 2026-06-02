package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kaustubhmishra/wealth-lens/backend/internal/common"
)

const (
	ContextUserID = "userID"
	ContextEmail  = "email"
)

type AccessTokenVerifier interface {
	VerifyAccessToken(token string) (uuid.UUID, string, error)
}

func RequireAuth(verifier AccessTokenVerifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			common.RespondError(c, common.Unauthorized("Missing bearer token"))
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		if token == "" || token == header {
			common.RespondError(c, common.Unauthorized("Invalid bearer token"))
			c.Abort()
			return
		}

		userID, email, err := verifier.VerifyAccessToken(token)
		if err != nil {
			common.RespondError(c, common.Unauthorized("Invalid or expired bearer token"))
			c.Abort()
			return
		}

		c.Set(ContextUserID, userID)
		c.Set(ContextEmail, email)
		c.Next()
	}
}

func CurrentUserID(c *gin.Context) (uuid.UUID, error) {
	value, ok := c.Get(ContextUserID)
	if !ok {
		return uuid.Nil, common.Unauthorized("Missing authenticated user")
	}

	userID, ok := value.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, common.Unauthorized("Invalid authenticated user")
	}

	return userID, nil
}
