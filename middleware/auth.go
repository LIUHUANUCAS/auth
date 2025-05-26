package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/LIUHUANUCAS/auth/utils"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is a middleware for authentication
type AuthMiddleware struct {
	jwtManager *utils.JWTManager
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(jwtManager *utils.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// AuthRequired is a middleware that requires authentication
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check if the header has the Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header format must be Bearer {token}",
			})
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Validate the token
		claims, err := m.jwtManager.ValidateAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Set the user ID in the context
		c.Set("userID", claims.UserID)

		// Continue
		c.Next()
	}
}

// UserContext is a key type for context values
type UserContext string

const (
	// UserIDKey is the key for user ID in context
	UserIDKey UserContext = "userID"
)

// GetUserID gets the user ID from the context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// SetUserID sets the user ID in the context
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}
