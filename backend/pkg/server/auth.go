package server

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	expected := os.Getenv("API_TOKEN")
	return func(c *gin.Context) {
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "API_TOKEN not configured"})
			return
		}

		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token != "" && token == expected {
			c.Next()
			return
		}

		cookie, err := c.Cookie("session")
		if err == nil && cookie == expected {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}
