package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type fakeAccessTokenVerifier struct {
	userID uuid.UUID
	email  string
	err    error
	token  string
}

func (f *fakeAccessTokenVerifier) VerifyAccessToken(token string) (uuid.UUID, string, error) {
	f.token = token
	if f.err != nil {
		return uuid.Nil, "", f.err
	}
	return f.userID, f.email, nil
}

func TestRequireAuthAcceptsBearerTokenAndSetsContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.New()
	verifier := &fakeAccessTokenVerifier{
		userID: userID,
		email:  "investor@example.com",
	}

	router := gin.New()
	router.Use(RequireAuth(verifier))
	router.GET("/protected", func(c *gin.Context) {
		currentUserID, err := CurrentUserID(c)
		if err != nil {
			t.Fatalf("CurrentUserID returned error: %v", err)
		}
		if currentUserID != userID {
			t.Fatalf("CurrentUserID = %s, want %s", currentUserID, userID)
		}

		email, ok := c.Get(ContextEmail)
		if !ok || email != "investor@example.com" {
			t.Fatalf("email context = %v, %t; want investor@example.com, true", email, ok)
		}

		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if verifier.token != "access-token" {
		t.Fatalf("verifier token = %q, want %q", verifier.token, "access-token")
	}
}

func TestRequireAuthRejectsInvalidBearerTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		header        string
		verifierError error
	}{
		{
			name:   "missing header",
			header: "",
		},
		{
			name:   "missing bearer prefix",
			header: "access-token",
		},
		{
			name:          "verifier rejects token",
			header:        "Bearer access-token",
			verifierError: errors.New("invalid token"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			verifier := &fakeAccessTokenVerifier{err: tt.verifierError}
			router := gin.New()
			router.Use(RequireAuth(verifier))
			router.GET("/protected", func(c *gin.Context) {
				t.Fatal("protected handler should not run")
			})

			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.header != "" {
				request.Header.Set("Authorization", tt.header)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
			}
		})
	}
}
