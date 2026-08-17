package middleware

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"support-ticket-api/internal/model"
	"support-ticket-api/internal/repository"
)

type contextKey string

const userContextKey contextKey = "current_user"

func CurrentUser(c *gin.Context) (model.User, bool) {
	value, ok := c.Get(string(userContextKey))
	if !ok {
		return model.User{}, false
	}
	user, ok := value.(model.User)
	return user, ok
}

func RequireRole(repo repository.Store, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		rawID := c.GetHeader("X-User-ID")
		if rawID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID header"})
			return
		}
		id, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil || id <= 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid X-User-ID header"})
			return
		}
		user, err := repo.GetUserByID(c.Request.Context(), id)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unknown user"})
			return
		}
		if _, ok := allowed[user.Role]; !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not allowed"})
			return
		}
		c.Set(string(userContextKey), user)
		c.Next()
	}
}
