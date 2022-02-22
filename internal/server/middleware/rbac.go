package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// RBACMiddleware Validates if a given token has the needed permissions to
// continue the request. It accepts a list of allowed roles for that endpoint.
// If the user don't own the resource the role will determine if the user can
// access it or not.
// The path must have an owner parameter.
func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {

	return func(c *gin.Context) {
		owner := c.Param("owner")
		if owner != "" {
			userID, ok := c.Get(SubClaim)
			if !ok {
				zap.L().Warn("invalid user context")
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}

			if owner == userID {
				c.Next()
			}
		}

		role, ok := c.Get(RoleClaim)
		if !ok {
			zap.L().Warn("invalid role context")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		c.AbortWithStatus(http.StatusForbidden)
	}

}
