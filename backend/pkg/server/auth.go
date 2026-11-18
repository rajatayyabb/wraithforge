package server

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"wraithforge/backend/pkg/database"
)

// AuthMiddleware authenticates API requests. A request is accepted if it
// presents a valid token via any of:
//   - Authorization: Bearer <token>
//   - a "session" cookie
//   - a "token" query parameter (needed by EventSource, which cannot set
//     request headers)
//
// A token is valid if it matches the API_TOKEN env var (a dev convenience,
// compared in constant time) or a non-revoked token in the api_tokens table.
// If neither an env token nor a database is configured, the middleware fails
// closed and rejects every request.
func AuthMiddleware(db *database.DB) gin.HandlerFunc {
	envToken := os.Getenv("API_TOKEN")

	return func(c *gin.Context) {
		presented := extractToken(c)
		if presented == "" {
			unauthorized(c)
			return
		}

		if envToken != "" && constantTimeEqual(presented, envToken) {
			c.Next()
			return
		}

		if db != nil {
			ok, err := db.ValidateAPIToken(c.Request.Context(), presented)
			if err == nil && ok {
				c.Next()
				return
			}
		}

		unauthorized(c)
	}
}

// extractToken pulls the bearer token from the Authorization header, then the
// session cookie, then the token query parameter, in that order.
func extractToken(c *gin.Context) string {
	if header := c.GetHeader("Authorization"); header != "" {
		if token := strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")); token != "" {
			return token
		}
	}
	if cookie, err := c.Cookie("session"); err == nil && cookie != "" {
		return cookie
	}
	if q := c.Query("token"); q != "" {
		return q
	}
	return ""
}

// constantTimeEqual compares two strings without leaking their relationship
// through timing. The length check short-circuits, which is acceptable: it
// leaks only length, not content.
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func unauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
}
